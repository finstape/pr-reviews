# PR Review Service

Сервис назначения ревьюеров для Pull Request'ов.

## Описание

Сервис автоматически назначает ревьюеров на Pull Request'ы из команды автора, позволяет выполнять переназначение ревьюверов и получать список PR'ов, назначенных конкретному пользователю, а также управлять командами и активностью пользователей.

## Требования

- Docker
- Docker Compose

## Запуск

Для запуска проекта выполните команду:

```bash
docker-compose up
```

После запуска сервис будет доступен на порту `8080`.

### Переменные окружения

Все необходимые переменные окружения заданы в `docker-compose.yml`. При необходимости их можно переопределить через файл `.env` или напрямую в `docker-compose.yml`.

## API Endpoints

### Teams

- `POST /team/add` - Создать команду с участниками
- `GET /team/get?team_name=<name>` - Получить команду с участниками

### Users

- `POST /users/setIsActive` - Установить флаг активности пользователя
- `GET /users/getReview?user_id=<id>` - Получить PR'ы, где пользователь назначен ревьювером

### Pull Requests

- `POST /pullRequest/create` - Создать PR и автоматически назначить до 2 ревьюверов
- `POST /pullRequest/merge` - Пометить PR как MERGED (идемпотентная операция)
- `POST /pullRequest/reassign` - Переназначить конкретного ревьювера

### Health

- `GET /healthz` - Health check endpoint

## Примеры использования

### Создание команды

```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {
        "user_id": "u1",
        "username": "Alice",
        "is_active": true
      },
      {
        "user_id": "u2",
        "username": "Bob",
        "is_active": true
      }
    ]
  }'
```

### Создание PR

```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1"
  }'
```

### Получение PR'ов пользователя

```bash
curl "http://localhost:8080/users/getReview?user_id=u2"
```

## Архитектура

Проект использует Clean Architecture с разделением на слои:

- `internal/entity` - доменные сущности
- `internal/usecase` - бизнес-логика
- `internal/repo` - интерфейсы репозиториев
- `internal/repo/persistent` - реализация репозиториев для PostgreSQL
- `internal/controller/http` - HTTP контроллеры
- `pkg` - вспомогательные пакеты (logger, postgres, httpserver)

## База данных

Используется PostgreSQL. Миграции запускаются автоматически при старте приложения.

## Тестирование

### Unit тесты

Для запуска unit тестов:

```bash
go test ./internal/...
```

Для запуска с подробным выводом:

```bash
go test -v ./internal/...
```

Для запуска с покрытием:

```bash
go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/...
go tool cover -html=coverage.txt
```

### Интеграционные тесты

Интеграционные тесты проверяют работу репозиториев с реальной базой данных.

Для запуска интеграционных тестов:

```bash
docker-compose -f docker-compose-integration-test.yml up --build --abort-on-container-exit
```

Или локально (требуется запущенная тестовая БД):

```bash
PG_URL="postgres://test_user:test_password@localhost:5433/db_test?sslmode=disable" go test -v -tags=integration ./integration-test/...
```

### Структура тестов

- **Unit тесты** (`internal/usecase/*/..._test.go`) - тестируют бизнес-логику с моками репозиториев
- **Табличные тесты** (`internal/usecase/*/..._table_test.go`) - тестируют различные сценарии в табличном формате
- **Тесты контроллеров** (`internal/controller/http/v1/*_test.go`) - тестируют HTTP handlers
- **Интеграционные тесты** (`integration-test/integration_test.go`) - тестируют работу с реальной БД

### Покрытие тестами

Проект включает:
- Unit тесты для всех usecase слоев (30+ тестов)
- Интеграционные тесты для репозиториев
- Тесты для HTTP контроллеров
- Табличные тесты для критических сценариев
- Покрытие edge cases и обработки ошибок

## Линтинг

Для проверки кода линтером:

```bash
golangci-lint run
```

## Разработка

### Структура проекта

```
.
├── cmd/app/              # Точка входа приложения
├── config/               # Конфигурация приложения
├── internal/
│   ├── entity/           # Доменные сущности
│   ├── usecase/          # Бизнес-логика
│   ├── repo/             # Интерфейсы репозиториев
│   │   └── persistent/   # Реализация репозиториев (PostgreSQL)
│   └── controller/       # HTTP контроллеры
│       └── http/v1/      # API версии 1
├── pkg/                  # Вспомогательные пакеты
│   ├── logger/           # Логирование
│   ├── postgres/         # Подключение к БД
│   └── httpserver/       # HTTP сервер
├── migrations/           # SQL миграции
├── integration-test/     # Интеграционные тесты
└── docker-compose.yml    # Docker Compose конфигурация
```

### Бизнес-логика

#### Назначение ревьюверов

При создании PR автоматически назначаются до 2 ревьюверов из команды автора:
- Выбираются только активные пользователи (`is_active = true`)
- Автор PR исключается из списка кандидатов
- Если кандидатов меньше 2, назначаются все доступные
- Если кандидатов нет, PR создается без ревьюверов

#### Мерж PR

Операция мержа идемпотентна:
- Если PR уже в статусе `MERGED`, возвращается текущее состояние без изменений
- При мерже устанавливается `merged_at` timestamp

#### Переназначение ревьювера

- Можно переназначить только для PR в статусе `OPEN`
- Новый ревьювер выбирается из активных участников команды старого ревьювера
- Старый ревьювер должен быть назначен на PR

### База данных

#### Схема БД

- `teams` - команды с участниками
- `users` - пользователи (связь с командами через `team_name`)
- `pull_requests` - Pull Request'ы
- `pr_reviewers` - связь многие-ко-многим между PR и ревьюверами

#### Миграции

Миграции применяются автоматически при старте приложения. Для ручного управления:

```bash
# Применить миграции
migrate -path migrations -database "$PG_URL?sslmode=disable" up

# Откатить последнюю миграцию
migrate -path migrations -database "$PG_URL?sslmode=disable" down
```

## Переменные окружения

Основные переменные окружения (заданы в `docker-compose.yml`):

- `APP_NAME` - имя приложения
- `APP_VERSION` - версия приложения
- `HTTP_PORT` - порт HTTP сервера (по умолчанию: 8080)
- `LOG_LEVEL` - уровень логирования (по умолчанию: info)
- `PG_URL` - строка подключения к PostgreSQL
- `PG_POOL_MAX` - максимальный размер пула соединений (по умолчанию: 10)

## Troubleshooting

### Проблемы с подключением к БД

Если приложение не может подключиться к БД:
1. Убедитесь, что PostgreSQL запущен: `docker-compose ps`
2. Проверьте переменную `PG_URL` в `docker-compose.yml`
3. Проверьте логи: `docker-compose logs app`

### Проблемы с миграциями

Если миграции не применяются:
1. Проверьте логи приложения на наличие ошибок
2. Убедитесь, что приложение собирается с тегом `migrate`: `go build -tags migrate`
3. Проверьте права доступа к БД

## Лицензия

См. файл LICENSE.
