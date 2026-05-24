# Agent Service — техническая документация

## Обзор

Agent Service — это отдельный микросервис, отвечающий за генерацию текстовых отчётов, объяснений и черновиков задач на основе LLM (Ollama). Он выделен из основного backend'а для независимого масштабирования, изоляции промптов и упрощения замены модели.

## Архитектура

```
Backend ──HTTP──▶ Agent-Service (8090) ──HTTP──▶ Ollama (11434)
                        │
                    internal/
                     ├── agent/        — ядро: клиент, промпты, генерация, guardrails
                     ├── config/       — env-конфиг
                     └── handlers/     — HTTP handlers
```

## Структура проекта

```
services/agent-service/
├── cmd/server/main.go              # entry point
├── internal/
│   ├── agent/
│   │   ├── client.go               # HTTP-клиент к Ollama
│   │   ├── prompts.go              # 13 промптов (на русском)
│   │   └── service.go              # Generate* + Guardrails + helpers
│   ├── config/
│   │   └── config.go               # загрузка переменных окружения
│   └── handlers/
│       └── handlers.go             # HTTP-хендлеры + регистрация маршрутов
├── Dockerfile
├── go.mod / go.sum
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `HTTP_ADDR` | `:8090` | Адрес HTTP-сервера |
| `AGENT_BASE_URL` | `http://localhost:11434` | URL Ollama API |
| `AGENT_MODEL` | `gemma2:2b` | Модель Ollama |
| `AGENT_TIMEOUT_SECONDS` | `120` | Таймаут запроса к Ollama |
| `LOG_LEVEL` | `info` | Уровень логирования |

## Эндпоинты

Все эндпоинты принимают `POST` с `Content-Type: application/json`.

### Level 1 — Базовые сводки

| Путь | Метод | Описание |
|---|---|---|
| `/api/agent/dataset-summary` | POST | Сводка качества датасета |
| `/api/agent/recommendations-summary` | POST | Объяснение рекомендаций |
| `/api/agent/roadmap-summary` | POST | Объяснение дорожной карты |

### Level 2 — Ролевые сводки

| Путь | Метод | Описание |
|---|---|---|
| `/api/agent/admin-summary` | POST | Сводка для администратора |
| `/api/agent/ml-engineer-summary` | POST | Сводка для ML-инженера |
| `/api/agent/annotator-summary` | POST | Сводка для разметчика |
| `/api/agent/expert-summary` | POST | Сводка для эксперта предметной области |
| `/api/agent/analyst-summary` | POST | Сводка для аналитика данных |

### Level 2 — Объяснение объекта

| Путь | Метод | Описание |
|---|---|---|
| `/api/agent/object-explanation` | POST | Почему объект требует внимания |

### Level 2 — Планы

| Путь | Метод | Описание |
|---|---|---|
| `/api/agent/collection-plan` | POST | План сбора реальных данных |
| `/api/agent/synthetic-plan` | POST | План синтетических данных |

### Level 2 — Генерация задач

| Путь | Метод | Описание |
|---|---|---|
| `/api/agent/generate-collection-tasks` | POST | Черновики задач на сбор данных (JSON) |
| `/api/agent/generate-synthetic-tasks` | POST | Черновики задач на синтетику (JSON) |
| `/api/agent/generate-annotator-brief` | POST | Brief для разметчиков |
| `/api/agent/generate-expert-brief` | POST | Brief для экспертов |

### Health

| Путь | Метод | Описание |
|---|---|---|
| `/health` | GET | Проверка работоспособности |

### Формат ответа (AgentResponse)

```json
{
  "summary": "текстовый ответ модели",
  "key_findings": ["вывод 1", "вывод 2"],
  "risks": ["риск 1", "риск 2"],
  "recommended_next_steps": ["шаг 1", "шаг 2"],
  "used_context_fields": ["поле1", "поле2"]
}
```

Эндпоинты `generate-collection-tasks` и `generate-synthetic-tasks` возвращают сырой JSON-массив, а не `AgentResponse`:

```json
[
  {
    "target_class": "cat",
    "target_count": 500,
    "priority": "high",
    "risk": "medium",
    "justification": "обоснование на русском"
  }
]
```

## Guardrails (защита от галлюцинаций)

После получения ответа от Ollama каждый метод проходит через `guardResponse`:

1. **Пустой ответ** — если `Summary` пуст, заменяется на заглушку
2. **Проверка чисел** — для каждого числового параметра контекста ищутся числа в ответе рядом с меткой; если отношение ответ/факт > 1.5 или < 0.5 — добавляется предупреждение
3. **Порог срабатывания** — при ≥ 3 предупреждениях в начало `Summary` добавляется префикс `[Quality check]`
4. **Пустые массивы** — `KeyFindings`, `Risks`, `RecommendedNextSteps` инициализируются пустыми массивами, если модель их не заполнила

## Язык промптов

Все промпты написаны на **русском языке**. Ключевые правила, вшитые в каждый промпт:
- "Каждое число должно быть из раздела Контекст выше. Никогда не выдумывай проценты, количества или оценки."
- "Тебе предоставлены: [поля]. Не ссылайся на другие поля."
- Если данных нет — ответ "Недостаточно данных для ответа."

## Безопасность

- Сервис не имеет доступа к базе данных — все контекстные данные передаются в теле запроса
- Таймаут ответа от Ollama — 180 секунд (конфигурируется через `AGENT_TIMEOUT_SECONDS`)
- Каждый ответ проходит post-processing (guardrails) перед возвратом

## Локальный запуск

```bash
# Сборка
cd services/agent-service
go build ./...

# Запуск
AGENT_BASE_URL=http://localhost:11434 \
AGENT_MODEL=gemma2:2b \
HTTP_ADDR=:8090 \
go run ./cmd/server
```

Через Makefile (из корня backend):

```bash
make agent-build
make agent-run
```

## Docker

```bash
docker compose up --build agent-service
```

В docker-compose сервис использует `host.docker.internal:11434` для доступа к Ollama на хосте.

## Клиент для backend

В `internal/agentsvc/client.go` определён HTTP-клиент, полностью повторяющий сигнатуры методов `agent.Service`. Используется в `internal/services/agent.go` для отправки контекстных данных в agent-service вместо прямого вызова Ollama.
