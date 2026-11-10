from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
import os
from camoufox.async_api import AsyncCamoufox 
import asyncio
from contextlib import asynccontextmanager
from urllib.parse import urlsplit

CONFIG = {
    "camoufox":{
        "timeout": int,
        "handless": True,
        "max_bro": int, 
    },
    "sites": [],
}

browser_pool = {}
pool_lock = asyncio.Lock()

app = FastAPI(title="Camoufox Scraping Service")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)
class Data(BaseModel):
    title: str
    performer: str
    chat_id: int


def config_loder():
    # подтягиваем
    max_bro = os.getenv("MAX_BRO")
    if max_bro is None:
        raise ValueError("MAX_BRO environment variable is not set")
    timeout = os.getenv("CAMOUFOX_TIMEOUT")
    if timeout is None:
        raise ValueError("CAMOUFOX_TIMEOUT environment variable is not set")
    site0 = os.getenv("SITE0")
    # подтягиваем ссылки на сайты для парсинга из конфигурации
    if site0 is None:
        raise ValueError("SITE1 environment variable is not set")

    # заполняем
    CONFIG["camoufox"]["timeout"] = int(timeout)
    CONFIG["camoufox"]["max_bro"] = int(max_bro)
    CONFIG["sites"].append(site0)

    logger.info(
        f"Configuration loaded: timeout={CONFIG['camoufox']['timeout']}, "
        f"headless={CONFIG['camoufox']['handless']}, "
        f"max_bro={CONFIG['camoufox']['max_bro']}, "
        f"sites={CONFIG['sites']}"
    )


async def init_browser_pool():
    global browser_pool
    try:
        for i in range(CONFIG["camoufox"]["max_bro"]):
            manager = AsyncCamoufox(
                headless = CONFIG["camoufox"]["handless"],
                timeout = CONFIG["camoufox"]["timeout"]
            )
            browser = await manager.__aenter__() 
            logger.info(f"Browser№{i} initialized")
            browser_pool[i] = {
                "manager": manager,
                "browser": browser,
                "busy": False
            }
    except Exception as e:
        logger.error(f"error init browser pool: {e}")
        raise HTTPException(status_code=500, detail=f"init error: {e}")

async def close_browser_pool():
    async with pool_lock:
        if not browser_pool:
            return
        
        # список задач на закрытие
        tasks = []
        for i, bro in browser_pool.items():
            tasks.append(bro["manager"].__aexit__(None, None, None))
        
        logger.info(f"Closing {len(tasks)} browsers concurrently...")
        
        # запускаем всё сразу и ждем завершения всех задач
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        for i, res in enumerate(results):
            if isinstance(res, Exception):
                logger.error(f"Browser {i} failed to close: {res}")
            else:
                logger.info(f"Browser {i} closed successfully")
    
        browser_pool.clear()
    


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Camoufox service is starting up...")
    try:
        config_loder()
        await init_browser_pool()
        yield
    finally:
        await close_browser_pool()
        logger.info("Camoufox service is shutting down...")


app = FastAPI(title="Camoufox Scraping Service", lifespan=lifespan)


async def get_free_browser():
    async with pool_lock:
        for i, bro in browser_pool.items():
            if not bro["busy"]:
                bro["busy"] = True
                logger.info(f"Browser№{i} allocated")
                return i, bro["browser"]
    raise HTTPException(status_code=503, detail="No free browsers available")

async def release_browser(index: int):
    async with pool_lock:
        if index in browser_pool:
            browser_pool[index]["busy"] = False
            logger.info(f"Browser№{index} released")
        else:
            logger.warning(f"Attempted to release non-existent browser index: {index}")


@app.post("/parse")
async def parse_music(data: Data):
    # найти ссылку 
    music_url = await find_url_by_camoufox(data.title, data.performer)

    if music_url == "":
        raise HTTPException(status_code=404, detail="Music not found")

    # отправить обратно в Go
    return {
        "title": data.title,
        "performer": data.performer,
        "url": music_url,
        "chat_id": data.chat_id,
    }      

async def find_url_by_camoufox(title, performer: str) -> str:
    i = None
    file_url = ""
    try:
        # подтягиваем свободный браузер из пула
        i, bro = await get_free_browser()
        page = await bro.new_page()

        # ищем ссылку на трек на всех сайтах из конфигурации
        for site_index, site_url in enumerate(CONFIG["sites"]):
            if site_index == 0:
                logger.info(f"Searching for track on site_index {site_index} using browser№{i}")
                file_url = await for_site0(site_url, title, performer, page)
                if file_url:
                    return file_url
        return "" 

    except Exception as e:
        logger.error(f"error in find_url_by_camoufox: {e}")
        raise HTTPException(status_code=500, detail=f"error in find_url_by_camoufox: {e}")
    finally:
        if page is not None:
            await page.close()
            logger.info(f"Page closed for browser№{i}")
        if i is not None:
            await release_browser(i)


async def for_site0(site_url, title, performer: str, page) -> str:
    #создаем поисковый запрос
    query = f"{performer} {title}"

    # открываем страницу сайта
    await page.goto(site_url, wait_until="networkidle")
    
    # ищем строку поиска на странице и вводим запрос
    search_field = page.locator("#story")
    await search_field.wait_for(state="visible")
    await search_field.fill(query)
    await search_field.press("Enter")
    logger.info("Find track")

    # ищем ссылку на трек в результатах поиска
    track_link = page.locator("a.track-item__desc").filter(has_text=f"{performer} - {title}").first
    await track_link.wait_for(state="visible")
    await track_link.click()
    await page.wait_for_load_state("networkidle")
    logger.info("Opened track page")

    # скрапим ссылку на скачивание трека
    download_link = page.locator('a.page__dl[download][href*="/nos/"]')
    
    if await download_link.count() == 0:
        logger.info("Download link not found on %s", page.url)
        return ""

    file_url = await download_link.first.evaluate("(element) => element.href")
    parsed_url = urlsplit(file_url)

    if (
        parsed_url.scheme != "https"
        or not parsed_url.path.lower().endswith(".mp3")
    ):
        raise ValueError(f"Unexpected download URL on {page.url}: {file_url}")

    logger.info("Found track download link on %s", page.url)
    return file_url
