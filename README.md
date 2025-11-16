# PR Reviewer Assignment Service

## Быстрый старт

1. **Клонируйте репозиторий** (если еще не сделано):
```bash
git clone <repository-url>
cd avito-tech-internship
```

2. **Запустите сервис и базу данных**:
```bash
make docker-up
# или
docker-compose up --build
```

3. **Проверьте, что сервис работает**:
```bash
curl http://localhost:8080/health
# Должен вернуть: {"status":"ok"}
```

4. **Остановка сервиса**:
```bash
make docker-down
# или
docker-compose down
```

## База данных

### Автоматические миграции

При запуске сервиса миграции применяются автоматически. Файлы миграций находятся в `internal/migrations/`:
- `20251114_init.up.sql` - создание таблиц
- `20251114_init.down.sql` - откат миграций

### Подключение к БД

**Через Docker Compose:**
```bash
# Подключение к PostgreSQL контейнеру
docker exec -it avito-postgres psql -U avito -d avito_db
```

**Локально:**
```bash
psql -h localhost -p 5432 -U avito -d avito_db
# Пароль: avito
```

### Структура БД

- `teams` - команды
- `users` - пользователи
- `pull_requests` - Pull Request'ы
- `pr_reviewers` - связь PR и ревьюверов
- `schema_migrations` - таблица для отслеживания миграций


## API Endpoints

### Teams

- `POST /team/add` - Создать команду с участниками
- `GET /team/get?team_name=<name>` - Получить команду

### Users

- `POST /users/setIsActive` - Установить флаг активности пользователя
- `GET /users/getReview?user_id=<id>` - Получить PR'ы, где пользователь назначен ревьювером

### Pull Requests

- `POST /pullRequest/create` - Создать PR и автоматически назначить ревьюверов
- `POST /pullRequest/merge` - Пометить PR как MERGED (идемпотентная операция)
- `POST /pullRequest/reassign` - Переназначить ревьювера

### Statistics

- `GET /stats` - Получить статистику по назначениям ревьюверов

### Bulk Operations

- `POST /users/bulkDeactivate` - Массовая деактивация пользователей команды с безопасной переназначаемостью PR

### Health Check

- `GET /health` - Проверка здоровья сервиса

### Swagger UI

- `GET /swagger` - Интерактивная документация API (Swagger UI)
- `GET /swagger/openapi.yml` - OpenAPI спецификация в формате YAML

После запуска сервиса откройте в браузере: http://localhost:8080/swagger

## Примеры использования

### 1. Создание команды

```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }'
```

### 2. Создание PR

```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search feature",
    "author_id": "u1"
  }'
```

### 3. Получение команды

```bash
curl http://localhost:8080/team/get?team_name=backend
```

### 4. Мерж PR

```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001"
  }'
```

### 5. Получение статистики

```bash
curl http://localhost:8080/stats
```

### 6. Массовая деактивация пользователей

```bash
curl -X POST http://localhost:8080/users/bulkDeactivate \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "user_ids": ["u1", "u2"]
  }'
```

## Makefile команды

```bash
make build          # Собрать приложение
make run            # Запустить локально
make test           # Запустить тесты
make test-race      # Запустить тесты с race detector
make docker-up      # Запустить через Docker Compose
make docker-down    # Остановить Docker Compose
make migrate-up     # Применить миграции (CLI)
make migrate-down   # Откатить миграции (CLI)
make clean          # Очистить артефакты сборки
make deps           # Установить зависимости
make fmt            # Форматировать код
make lint           # Проверить код линтером
make install-linter # Установить golangci-lint
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SERVER_PORT` | Порт HTTP сервера | `8080` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `avito` |
| `DB_PASSWORD` | Пароль БД | `avito` |
| `DB_NAME` | Имя БД | `avito_db` |
| `DB_SSLMODE` | SSL режим | `disable` |

## Структура проекта

```
avito-tech-internship/
├── cmd/
│   └── server/          # Точка входа приложения
├── internal/
│   ├── config/          # Конфигурация
│   ├── domain/          # Доменные модели
│   ├── handler/         # HTTP обработчики
│   ├── repository/      # Интерфейсы и реализации репозиториев
│   ├── service/         # Бизнес-логика
│   └── migrations/      # SQL миграции
├── pkg/
│   └── migrate/         # Утилита для миграций
├── docker-compose.yml   # Docker Compose конфигурация
├── Dockerfile           # Docker образ приложения
├── Makefile             # Команды для сборки и запуска
└── openapi.yml          # OpenAPI спецификация
```

### Линтинг кода

```bash
# Установить линтер (если еще не установлен)
make install-linter

# Запустить линтер
make lint
```
### Нагрузочное тестирование

См. подробности в [test/load/README.md](test/load/README.md)

```bash
# Установить k6 (см. test/load/README.md)
# Запустить тесты
k6 run test/load/k6_test.js
```