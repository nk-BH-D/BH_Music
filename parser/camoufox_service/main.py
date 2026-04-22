from fastapi import FastAPI
from pydantic import BaseModel
import logging
import os
#import requests

app = FastAPI(title="Camoufox Scraping Service")


class Data(BaseModel):
    title: str
    performer: str
    chat_id: int

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@app.post("/parse")
def parse_music(data: Data):
    try:
        # 1. сформировать запрос
        query = f"{data.title} {data.performer}"

        # 2. найти ссылку (заглушка)
        music_url = find_url(query)

        if music_url == "":
            return {"error": "not found"}

        # 3. отправить обратно в Go
        return {
            "title": data.title,
            "performer": data.performer,
            "url": music_url,
            "chat_id": data.chat_id,
        }      
    except Exception as e:
        return {"error": str(e)}


def find_url(query: str) -> str:
    # тут будет код с реализацией
    return "хуй" #тест


@app.on_event("startup")
async def startup_event():
    #await init_browser_pool()
    logger.info("Сервис Camoufox Scraping запущен")


#@app.on_event("shutdown")
#async def shutdown_event():
    #async with browser_lock:
    #    for browser in browser_pool:
    #       try:
    #           if hasattr(browser, "stop"):
    #                browser.stop()
    #        except Exception as e:
    #           logger.error(f"Ошибка закрытия браузера: {e}")
    #  browser_pool.clear()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=int(os.getenv("CAMOUFOX_PORT")),
        log_level="info"
    )
    