FROM ghcr.io/astral-sh/uv:python3.11-bookworm-slim

WORKDIR /app/ml-service
ENV UV_HTTP_TIMEOUT=180

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        libgl1 \
        libglib2.0-0 \
        libsm6 \
        libxext6 \
        libxcb1 \
    && rm -rf /var/lib/apt/lists/*

COPY pyproject.toml uv.lock README.md ./
COPY app ./app
COPY scripts ./scripts
COPY src ./src
COPY data ./data

RUN uv sync --frozen

EXPOSE 8000

CMD ["/app/ml-service/.venv/bin/uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
