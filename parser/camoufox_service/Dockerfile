# Этап 1: Установка зависимостей
FROM python:3.11-slim AS builder

WORKDIR /app

COPY parser/camoufox_service/requirements.txt ./requirements.txt

RUN pip install --no-cache-dir --prefix=/install -r requirements.txt \
    && python -m spacy download ru_core_news_sm

# Этап 2: Финальный образ
FROM python:3.11-slim

WORKDIR /app

# Копируем зависимости из первого этапа
COPY --from=builder /install /usr/local

COPY parser/camoufox_service/ ./camoufox_service/

EXPOSE 8082

CMD ["uvicorn", "camoufox_service.main:app", "--host", "0.0.0.0", "--port", "8082"]
