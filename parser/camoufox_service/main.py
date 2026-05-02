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
    #port = os.getenv("CAMOUFOX_PORT")
    #if port is None:
    #    raise ValueError("CAMOUFOX_PORT environment variable is not set")
    #host = os.getenv("CAMOUFOX_HOST")
    #if host is None:
    #    raise ValueError("CAMOUFOX_HOST environment variable is not set")
    max_bro = os.getenv("MAX_BRO")
    if max_bro is None:
        raise ValueError("MAX_BRO environment variable is not set")
    timeout = os.getenv("CAMOUFOX_TIMEOUT")
    if timeout is None:
        raise ValueError("CAMOUFOX_TIMEOUT environment variable is not set")
    CONFIG["camoufox"]["timeout"] = int(timeout)
    CONFIG["camoufox"]["max_bro"] = int(max_bro)
    #CONFIG["config"]["port"] = int(port)
    #CONFIG["config"]["host"] = host
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
    # 1. сформировать запрос
    query = f"{data.title} {data.performer} скачать"

    # 2. найти ссылку (заглушка)
    music_url = find_url_by_camoufox(query)

    if music_url == "":
        raise HTTPException(status_code=404, detail="Music not found")

    # 3. отправить обратно в Go
    return {
        "title": data.title,
        "performer": data.performer,
        "url": music_url,
        "chat_id": data.chat_id,
    }      
    

async def find_url_by_camoufox(query: str) -> str:
    i = None
    try:
        i, bro = await get_free_browser()
        await bro.goto("https://www.google.com/search?q=" + query)
        # открыть страницу camoufox
    except Exception as e:
        logger.error(f"error in find_url_by_camoufox: {e}")
        raise HTTPException(status_code=500, detail=f"error in find_url_by_camoufox: {e}")
    finally:
        if i is not None:
            await release_browser(i)
    return "хуй" #тест


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


#if __name__ == "__main__":
#    import uvicorn
#    uvicorn.run(
#        app,
#        host=CONFIG["config"]["host"],
#        port=CONFIG["config"]["port"],
#        log_level="info"
#    )
    
