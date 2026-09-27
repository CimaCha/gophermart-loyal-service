# gophermart-loyal-service

# Gophermart — накопительная система лояльности

Курсовой проект Яндекс.Практикума (командная версия): HTTP API системы лояльности
«Гофермарт» и отдельный сервис расчёта баллов вознаграждения (accrual).

## Содержание

- [Архитектура](#архитектура)
- [Структура репозитория](#структура-репозитория)
- [Быстрый старт](#запустить-приложение)
- [Конфигурация](#конфигурация)
- [API](#api)
- [Background worker](#background-worker)
- [Этапы реализации](#этапы-реализации)
- [Тестирование](#тестирование)

## Архитектура

Проект состоит из двух независимых HTTP-сервисов, которые можно запускать
и деплоить отдельно друг от друга.

| Сервис | Назначение | Хранилище | Порт (по умолчанию) |
|---|---|---|---|
| `gophermart` | Регистрация/логин, приём номеров заказов, баланс баллов, списания | PostgreSQL | `RUN_ADDRESS` |
| `accrual` | Расчёт вознаграждений по составу заказа | PostgreSQL (может быть отдельная БД/схема) | `RUN_ADDRESS` (свой) |

**Взаимодействие:** фоновый воркер внутри `gophermart` опрашивает `accrual`
по HTTP (`GET /api/orders/{number}`), используя адрес из `ACCRUAL_SYSTEM_ADDRESS`.
Аутентификация между сервисами не требуется — `accrual` работает в доверенном контуре.

```
gophermart API ──▶ gophermart DB
       │
       ▼
background worker ──HTTP──▶ accrual API ──▶ accrual DB
```

Общая схема gophermart:

![Gophermart architecture](docs/gophermart.svg)

## Background worker

Worker отвечает за фоновую обработку заказов: получение статусов из accrual,
обновление состояния заказа и начисление баллов пользователю.

Он запускается внутри `gophermart` и работает независимо от HTTP API.

Основные этапы:

1. Получение заказа из очереди уведомлений или через polling.
2. Запрос текущего статуса в accrual.
3. Обновление статуса заказа.
4. Начисление баллов после успешной обработки.

Схема работы worker:

![Worker](docs/worker.svg)

### Требования

- Go 1.22+
- PostgreSQL 15+
- `migrate` CLI (или используемый в проекте инструмент миграций)

### Переменные окружения

**gophermart**

| Переменная | Флаг | Описание | По умолчанию |
|---|---|---|---|
| `RUN_ADDRESS` | `-a` | Адрес и порт запуска | `:8080` |
| `DATABASE_URI` | `-d` | Строка подключения к PostgreSQL | — |
| `ACCRUAL_SYSTEM_ADDRESS` | `-r` | Адрес сервиса accrual | — |

**accrual**

| Переменная | Флаг | Описание | По умолчанию |
|---|---|---|---|
| `RUN_ADDRESS` | `-a` | Адрес и порт запуска | `:8081` |
| `DATABASE_URI` | `-d` | Строка подключения к PostgreSQL | — |

## Конфигурация

Помимо флагов/переменных окружения, часть настроек (rate limit по роутам,
логирование) задаётся через `config/*.yaml`:

```yaml
directory: logs/gophermart

stdout:
  enabled: true
  format: text
  level: info

files:
  - name: app
    enabled: true
    format: json
    level: debug

cleanup_interval: 5m

register:
  key_prefix: register
  window: 1m
  max_requests: 5
login:
  key_prefix: login
  window: 3m
  max_requests: 10
```

## API

### Gophermart — публичное API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| POST | `/api/user/register` | нет | Регистрация пользователя |
| POST | `/api/user/login` | нет | Аутентификация |
| POST | `/api/user/orders` | да | Загрузка номера заказа |
| GET | `/api/user/orders` | да | Список загруженных заказов |
| GET | `/api/user/balance` | да | Текущий баланс баллов |
| POST | `/api/user/balance/withdraw` | да | Списание баллов |
| GET | `/api/user/withdrawals` | да | История списаний |

### Accrual — приватное API (доверенный контур)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/api/orders/{number}` | Получение статуса расчёта по заказу |
| POST | `/api/orders` | Регистрация нового заказа для расчёта |
| POST | `/api/goods` | Регистрация правила вознаграждения за товар |

Статусы расчёта: `REGISTERED → PROCESSING → PROCESSED | INVALID`.

## Этапы реализации

- [x] Фундамент: конфигурация, логгер, подключение к БД, graceful shutdown
- [x] Middleware: logging, request ID, rate limit, recovery
- [x] JWT: построение и валидация токена
- [x] User: регистрация и аутентификация
- [x] Order: приём и выдача списка заказов
- [x] Balance: баланс и списание баллов
- [ ] Accrual: HTTP API (`/api/orders/{number}`, `/api/orders`, `/api/goods`)
- [ ] Accrual: асинхронный расчёт вознаграждений (матчинг по `match`)
- [x] Background worker в gophermart: опрос accrual, обработка `429 Retry-After`
- [ ] Rate limiting на accrual API
- [x] Интеграционные тесты (БД, HTTP)
- [ ] Юнит-тесты — покрытие ≥ 60%
- [ ] Документация по экспортируемым сущностям

## Тестирование

```bash
go test ./... -cover
```

Требование по ТЗ — покрытие тестами не менее 60% по всей системе.
Для интеграционных тестов с БД используется `testcontainers` (см. `*_test.go`
в пакетах `repository`).

## Запустить приложение

Перед запуском убедитесь, что PostgreSQL запущен и настроены файлы
`.env.gophermart` и `.env.accrual`.

### Через Makefile

Запустить сервис **gophermart**:

```bash
make run-gophermart
```

Запустить сервис **accrual**:

```bash
make run-accrual
```

### Напрямую через Go

Запустить **gophermart**:

```bash
go run ./cmd/gophermart \
  -a :8080 \
  -d postgres://user:pass@localhost:5432/gophermart_db?sslmode=disable
```

Запустить **accrual**:

```bash
go run ./cmd/accrual \
  -a :8081 \
  -d postgres://user:pass@localhost:5432/accrual_db?sslmode=disable
```

### Запустить тесты (интеграционные и юниты)

```bash
make run-test
```
