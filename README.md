# DataForge AI Platform

Платформа для курирования датасетов с ML-анализом и LLM-агентом. Позволяет загружать, анализировать, проверять и экспортировать датасеты классификации изображений.

## Архитектура

```
┌──────────┐     ┌───────────┐     ┌──────────────┐     ┌──────────┐
│ Frontend │────▶│  Backend  │────▶│  ML-Service  │     │  Ollama  │
│ (Next.js)│     │  (Go/chi) │     │  (Python)    │     │  (Gemma) │
└──────────┘     └─────┬─────┘     └──────────────┘     └──────────┘
                       │                                      ▲
                       ▼                                      │
                 ┌──────────┐                          ┌──────────────┐
                 │ Postgres │                          │ Agent-Service │
                 └──────────┘                          │  (Go/Ollama)  │
                                                        └──────────────┘
```

**Flow данных:**

1. Фронт загружает датасет (CSV + ZIP) → Backend
2. Backend отправляет данные в ML-Service для анализа → получает метрики
3. Backend собирает контекст (dashboard, рекомендации, roadmap) и отправляет в Agent-Service
4. Agent-Service формирует промпт, вызывает Ollama (Gemma), возвращает ответ на русском
5. Backend возвращает результат на фронт

## Состав проекта

| Компонент | Технология | Порт |
|---|---|---|
| **Backend** | Go 1.22 + chi + pgx | `:8080` |
| **Frontend** | Next.js 15 + React 19 + Tailwind | `:3000` |
| **ML-Service** | Python 3.11 + FastAPI | `:8000` |
| **Agent-Service** | Go 1.22 (отдельный модуль) | `:8090` |
| **Postgres** | PostgreSQL 15 | `:5432` |
| **Ollama** | LLM-сервер (gemma2:2b) | `:11434` |

## Зависимости

- **Docker** + Docker Compose — для запуска всей инфраструктуры
- **Go 1.22+** — для локальной разработки backend / agent-service
- **Node.js 18+** — для локальной разработки frontend
- **Python 3.11** — только для ML-Service
- **Ollama** — для работы Agent-Service

## Быстрый старт (Docker Compose)

```bash
# 1. Клонировать репозиторий
git clone <repo>
cd Hakaton_MIREA

# 2. Убедиться, что Ollama запущена (нужна на хосте, не в Docker)
ollama serve
ollama pull gemma2:2b

# 3. Запустить все сервисы
cd backend
docker compose up --build
```

После запуска:
- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Agent-Service: `http://localhost:8090`
- ML-Service: `http://localhost:8000`

## Локальная разработка

### Backend

```bash
cd backend

# Установка зависимостей
go mod tidy

# Настройка окружения (скопировать и отредактировать)
cp .env.example .env

# Запуск (требуется PostgreSQL и Agent-Service)
go run ./cmd/api
```

Переменные окружения:

| Переменная | По умолчанию | Описание |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Порт backend |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/hakaton?sslmode=disable` | Подключение к БД |
| `STORAGE_DIR` | `storage` | Директория для файлов |
| `ML_SERVICE_URL` | `http://ml-service:8000` | URL ML-сервиса |
| `ML_TIMEOUT_SECONDS` | `15` | Таймаут ML-сервиса |
| `JWT_SECRET` | `change-me-in-production` | Секрет для JWT |
| `AGENT_SERVICE_URL` | `http://agent-service:8090` | URL Agent-Service |

### Agent-Service

```bash
cd backend/services/agent-service

# Сборка и запуск
go build ./...
AGENT_BASE_URL=http://localhost:11434 \
AGENT_MODEL=gemma2:2b \
HTTP_ADDR=:8090 \
go run ./cmd/server
```

Или через Makefile из корня backend:

```bash
make agent-run
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### ML-Service

```bash
cd ml-service
uv sync --frozen
uv run uvicorn app.main:app --reload
```

### Postgres (через Docker)

```bash
docker run -d \
  --name dataforge-pg \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=hakaton \
  -p 5432:5432 \
  postgres:15-alpine
```

## Основные эндпоинты

### Публичные (без авторизации)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/health` | Проверка работоспособности |
| POST | `/api/auth/register` | Регистрация |
| POST | `/api/auth/login` | Вход |

### Проекты (требуют JWT)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/projects` | Создать проект |
| GET | `/api/projects` | Список проектов |
| GET | `/api/projects/{id}` | Детали проекта |
| POST | `/api/projects/{id}/upload` | Загрузить датасет (CSV + ZIP) |
| POST | `/api/projects/{id}/analyze` | Запустить анализ |
| GET | `/api/projects/{id}/dashboard` | Дашборд |
| GET | `/api/projects/{id}/probabilistic-analysis` | Вероятностный анализ |
| GET | `/api/projects/{id}/review-queue` | Очередь проверки |
| GET | `/api/projects/{id}/objects` | Список объектов (с фильтрами/сортировкой) |
| GET | `/api/projects/{id}/objects/{oid}` | Детали объекта |
| POST | `/api/projects/{id}/objects/{oid}/action` | Действие над объектом |
| POST | `/api/projects/{id}/objects/{oid}/comments` | Добавить комментарий |
| GET | `/api/projects/{id}/recommendations` | Рекомендации |
| GET | `/api/projects/{id}/roadmap` | Дорожная карта |
| GET | `/api/projects/{id}/collection-tasks` | Задачи на сбор данных |
| POST | `/api/projects/{id}/collection-tasks` | Создать задачу на сбор |
| GET | `/api/projects/{id}/synthetic-tasks` | Задачи на синтетику |
| POST | `/api/projects/{id}/synthetic-tasks` | Создать задачу на синтетику |
| GET | `/api/projects/{id}/class-action-plan` | План действий по классам |
| POST | `/api/projects/{id}/export` | Создать экспорт |
| GET | `/api/projects/{id}/exports/{eid}/download` | Скачать экспорт |

### Agent-эндпоинты (через Backend, требуется JWT)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/projects/{id}/agent/dataset-summary` | Сводка датасета |
| POST | `/api/projects/{id}/agent/recommendations-summary` | Объяснение рекомендаций |
| POST | `/api/projects/{id}/agent/roadmap-summary` | Объяснение roadmap |
| POST | `/api/projects/{id}/agent/admin-summary` | Для администратора |
| POST | `/api/projects/{id}/agent/ml-engineer-summary` | Для ML-инженера |
| POST | `/api/projects/{id}/agent/annotator-summary` | Для разметчика |
| POST | `/api/projects/{id}/agent/expert-summary` | Для эксперта |
| POST | `/api/projects/{id}/agent/analyst-summary` | Для аналитика |
| POST | `/api/projects/{id}/objects/{oid}/agent/explain` | Объяснить объект |
| POST | `/api/projects/{id}/agent/collection-plan` | План сбора данных |
| POST | `/api/projects/{id}/agent/synthetic-plan` | План синтетики |
| POST | `/api/projects/{id}/agent/generate-collection-tasks` | Черновики задач на сбор |
| POST | `/api/projects/{id}/agent/generate-synthetic-tasks` | Черновики задач на синтетику |
| POST | `/api/projects/{id}/agent/generate-annotator-brief` | Brief для разметчиков |
| POST | `/api/projects/{id}/agent/generate-expert-brief` | Brief для экспертов |

## Пример curl-запроса (полный цикл)

```bash
# Регистрация
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@test.com","name":"User","password":"test123"}' | jq -r .token)

# Создать проект
PROJECT_ID=$(curl -s -X POST http://localhost:8080/api/projects \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Demo","modality":"image","task_type":"classification","classes":["cat","dog"]}' | jq -r .id)

# Загрузить датасет
curl -X POST "http://localhost:8080/api/projects/$PROJECT_ID/upload" \
  -H "Authorization: Bearer $TOKEN" \
  -F dataset=@dataset.csv \
  -F images=@images.zip

# Запустить анализ
curl -X POST "http://localhost:8080/api/projects/$PROJECT_ID/analyze" \
  -H "Authorization: Bearer $TOKEN"

# Получить сводку от агента
curl -X POST "http://localhost:8080/api/projects/$PROJECT_ID/agent/dataset-summary" \
  -H "Authorization: Bearer $TOKEN" | jq .summary
```

## Тестирование

```bash
cd backend
go test ./...
go vet ./...
```

## Makefile (backend)

```bash
make tidy          # go mod tidy
make fmt           # форматирование кода
make build         # сборка
make run           # запуск backend
make agent-build   # сборка agent-service
make agent-run     # запуск agent-service
make docker-up     # запуск через Docker Compose
make docker-down   # остановка Docker Compose
```
