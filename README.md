# Тестовое задание Luna
## Стек

Go, MySQL 8, Redis 7, Docker Compose, JWT, Swagger, Prometheus

## Быстрый старт

### Docker Compose 

```bash
docker compose up --build
```

Поднимает MySQL, Redis и приложение. Миграции применяются автоматически при старте.

Приложение доступно на `http://localhost:8080`.

### Локально (без Docker)

Требования: Go 1.24+, MySQL 8, Redis 7.

1. Запустить MySQL и Redis:
```bash
docker compose up mysql redis
```

2. Запустить приложение:
```bash
make run
```

Или собрать бинарник:
```bash
make build
./bin/task_vault
```

### Конфигурация

Приоритет: переменные окружения > `config.yaml` > значения по умолчанию.

Основные переменные:

| Переменная | Описание | По умолчанию |
|------------|----------|-------------|
| `SERVER_PORT` | Порт сервера | `8080` |
| `MYSQL_DSN` | DSN подключения к MySQL | `root:root@tcp(localhost:3306)/task_vault?parseTime=true` |
| `REDIS_ADDR` | Адрес Redis | `localhost:6379` |
| `JWT_SECRET` | Секрет для JWT-токенов | `dev-secret-change-me` |
| `JWT_TTL` | Время жизни токена | `24h` |

Пример `config.yaml`:

```yaml
server:
  port: "8080"

mysql:
  dsn: "root:root@tcp(localhost:3306)/task_vault?parseTime=true"
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 5m

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

jwt:
  secret: "dev-secret-change-me"
  ttl: 24h
```

## API документация

Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

Перегенерация после изменения аннотаций:
```bash
make swagger
```

## Эндпоинты

### Аутентификация
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/register` | Регистрация |
| POST | `/api/v1/login` | Авторизация (возвращает JWT) |

### Команды (требуется JWT)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/teams` | Создать команду |
| GET | `/api/v1/teams` | Список команд пользователя |
| POST | `/api/v1/teams/{id}/invite` | Пригласить в команду (owner/admin) |

### Задачи (требуется JWT)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/tasks` | Создать задачу |
| GET | `/api/v1/tasks` | Список задач (фильтры: `team_id`, `status`, `assignee_id`, `limit`, `offset`) |
| PUT | `/api/v1/tasks/{id}` | Обновить задачу |
| GET | `/api/v1/tasks/{id}/history` | История изменений задачи |

### Служебные
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe (MySQL + Redis) |
| GET | `/metrics` | Prometheus-метрики |

## Пример использования

```bash
# Регистрация
curl -s -X POST http://localhost:8080/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret123","name":"John"}'

# Логин
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret123"}' | jq -r '.data.token')

# Создать команду
curl -s -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"My Team"}'

# Список задач
curl -s http://localhost:8080/api/v1/tasks?team_id=<uuid> \
  -H "Authorization: Bearer $TOKEN"
```

## Тесты

```bash
# Unit-тесты
make test

# Интеграционные тесты (требуется Docker)
go test ./internal/adapter/mysql/... -v
```

## Архитектура

```
cmd/task_vault/         — точка входа
internal/
  domain/               — бизнес-сущности (User, Team, Task, ...)
  ports/                — интерфейсы (репозитории, кеш)
  app/
    command/            — write-операции (CQRS)
    query/              — read-операции
    auth/               — JWT
  adapter/
    mysql/              — реализация репозиториев
    redis/              — кеш
    http/
      handler/          — HTTP-обработчики + DTO + валидация
      middleware/       — JWT-авторизация, rate limiting, логирование, метрики
    logging/            — декораторы с логированием
  config/               — конфигурация (YAML + ENV)
migrations/             — SQL-миграции
docs/                   — сгенерированная Swagger-документация
```

## Makefile

```bash
make build      # Сборка бинарника
make run        # Запуск
make test       # Тесты
make swagger    # Перегенерация Swagger
```
