FROM python:3.11-slim AS builder

WORKDIR /app

COPY requirements.txt .
 
# ставим python-зависимости + браузер
RUN pip install --no-cache-dir -r requirements.txt \
    && python -m camoufox fetch

FROM python:3.11-slim

WORKDIR /app/parser

# системные зависимости для Firefox 
RUN apt-get update && apt-get install -y --no-install-recommends \
    libnss3 \
    libatk1.0-0 \
    libx11-6 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libgbm1 \
    libasound2 \
    libpangocairo-1.0-0 \
    libgtk-3-0 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# перенос python пакетов
COPY --from=builder /usr/local /usr/local

# перенос браузера
COPY --from=builder /root/.cache/camoufox /root/.cache/camoufox

COPY . ./camoufox_service/

CMD ["uvicorn", "camoufox_service.main:app"]
