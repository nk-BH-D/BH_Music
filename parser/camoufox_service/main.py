from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
import os
from camoufox.async_api import AsyncCamoufox 
import asyncio
#import requests

CONFIG = {
    "camoufox":{
        "timeout": int,
        "handless": True,
        "max_bro": int, 
    },
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
    max_bro = os.getenv("MAX_BRO")
    if max_bro is None:
        raise ValueError("MAX_BRO environment variable is not set")
    timeout = os.getenv("CAMOUFOX_TIMEOUT")
    if timeout is None:
        raise ValueError("CAMOUFOX_TIMEOUT environment variable is not set")
    CONFIG["camoufox"]["timeout"] = int(timeout)
    CONFIG["camoufox"]["max_bro"] = int(max_bro)
    logger.info(f"Configuration loaded: {CONFIG}")


async def init_browser_pool():
    global browser_pool
    try:
        for i in range(CONFIG["camoufox"]["max_bro"]):
            browser = AsyncCamoufox(
                headless = CONFIG["camoufox"]["handless"],
                timeout = CONFIG["camoufox"]["timeout"]
            )
            await browser.start() 
            logger.info(f"Browser№{i} initialized")
            browser_pool[i] = {
                "browser": browser,
                "busy": False
            }
    except Exception as e:
        logger.error(f"error init browser pool: {e}")
        raise HTTPException(status_code=500, detail=f"init error: {e}")


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
    music_url = find_url_by_camoufox(data.title, data.performer)

    if music_url == "":
        raise HTTPException(status_code=404, detail="Music not found")

    # отправить обратно в Go
    return {
        "title": data.title,
        "performer": data.performer,
        "url": music_url,
        "chat_id": data.chat_id,
    }      

# ----------------!!!--------------------   
# надо сделать защиту от редиректа
# надо сделать функцию которая проверяет ссылки на мусор
# надо сделать поведение парсера болие универсальным
# ----------------!!!--------------------

async def find_url_by_camoufox(title, performer: str) -> str:
    i = None
    file_url = ""
    try:
        i, bro = await get_free_browser()
        page = await bro.new_page()
        query = f"{title} {performer} скачать"
        await page.goto(f"https://www.google.com/search?q={query}" , wait_until="networkidle")
        # берём все выданные поиском ссылки
        links = await page.query_selector_all('a:has(h3)')
        search_results = []
        for link in links:
            href = await link.get_attribute('href')
            if href and href.startswith('http'):
                search_results.append(href)
        logger.info(f"Search results for query '{query}': {search_results}")
        # открываем каждую ссылку и ищем музыку
        for url in search_results:
            await page.goto(url, wait_until="networkidle")
            # закрываем все всплывающие окна
            page.on("popup", lambda p: p.close())

            # пытаемся сразу найти ссылку на музыку
            file_url = await dounload_btn(page, url)
            if file_url and file_url.startswith("http"):
                logger.info(f"Found music URL: {file_url} on page {url}")
                break

            # если не нашли, то ищем поле поиска на сайте
            search_field = await page.query_selector(
                'input[name="q"], input[name="search"], input[type="search"], input[placeholder*="поиск" i], input[placeholder*="search" i]'
            )
            if search_field:
                # печатаем название песни
                await search_field.fill(f"{title} {performer}")
                await search_field.press("Enter")

                # ждем загрузки результатов уже на самом сайте
                await page.wait_for_load_state("networkidle")

                # ищем ссылки на музыку
                file_url = await dounload_btn(page, url)
                if file_url != "":
                    logger.info(f"Found music URL: {file_url} on page {url}")
                    break
                else:
                    logger.info(f"No music URL found on page {url}")
                    continue

            else:
                logger.info(f"No search field found on page {url}")
                continue      
    except Exception as e:
        logger.error(f"error in find_url_by_camoufox: {e}")
        raise HTTPException(status_code=500, detail=f"error in find_url_by_camoufox: {e}")
    finally:
        if page:
            await page.close()
        if i is not None:
            await release_browser(i)
    return file_url


async def dounload_btn(page, url) -> str:
    file_url = ""
    download_btn  = await page.query_selector('a[href*=".mp3"]')
    if download_btn:
        file_url = await download_btn.get_attribute("href")
    else:
        selectors = [
            "a.download", 
            "a.btn-download", 
            "button.download", 
            "[title*='Скачать']", 
            ".download-link"
        ]
        for selector in selectors:
            btn = await page.query_selector(selector)
            if btn:
                file_url = await btn.get_attribute("href")

    if file_url and file_url.startswith("http"):
        logger.info(f"Found music URL: {file_url} on page {url}")
        return file_url


@app.on_event("startup")
async def startup_event():
    config_loder()
    await init_browser_pool()
    logger.info("Сервис Camoufox Scraping запущен")


@app.on_event("shutdown")
async def shutdown_event():
    async with pool_lock:
        if not browser_pool:
            return
        
        # список задач на закрытие
        tasks = []
        for i, bro in browser_pool.items():
            tasks.append(bro["browser"].__aexit__(None, None, None))
        
        logger.info(f"Closing {len(tasks)} browsers concurrently...")
        
        # запускаем всё сразу и ждем завершения всех задач
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        for i, res in enumerate(results):
            if isinstance(res, Exception):
                logger.error(f"Browser {i} failed to close: {res}")
            else:
                logger.info(f"Browser {i} closed successfully")

    browser_pool.clear()
