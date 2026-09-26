include .env.gophermart
include .env.accrual
export

# ------------------------------------------------------------------------------
# Run
# ------------------------------------------------------------------------------

# Запустить gophermart сервис
run-gophermart:
	set -a && . ./.env.gophermart && set +a && \
	go run ./cmd/gophermart \
		-a $$RUN_ADDRESS \
		-d $$DATABASE_URI

# Запустить accrual сервис (пока недоступен)
run-accrual:
	set -a && . ./.env.accrual && set +a && \
	go run ./cmd/accrual \
		-a $$RUN_ADDRESS \
		-d $$DATABASE_URI

# ------------------------------------------------------------------------------
# Tests
# ------------------------------------------------------------------------------

# Почистить кэш тестов
test-clean:
	go clean -testcache

# Запустить тесты во всем приложение
test:
	go test -v ./...

# ------------------------------------------------------------------------------
# Gophermart migrations
# ------------------------------------------------------------------------------

# Создать новую миграцию для gophermart
goose-gophermart-new:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make goose-gophermart-new name=create_users"; \
		exit 1; \
	fi; \
	set -a && . ./.env.gophermart && set +a && \
	goose create $(name) sql

# Показать текущую версию миграции у gophermart сервиса
goose-gophermart-status:
	set -a && . ./.env.gophermart && set +a && \
	goose status

# Применить последнюю версию миграций в gophermart сервисе
goose-gophermart-up:
	set -a && . ./.env.gophermart && set +a && \
	goose up

# Применить следующую версию миграций для gophermart сервиса
goose-gophermart-up-by-one:
	set -a && . ./.env.gophermart && set +a && \
	goose up-by-one

# Вернуть предыдущую версию миграций для gophermart сервиса
goose-gophermart-down:
	set -a && . ./.env.gophermart && set +a && \
	goose down

# Вернуться к конкретной предыдущей версии (по ID) миграций для gophermart сервиса
goose-gophermart-down-to:
	@if [ -z "$(id)" ]; then \
		echo "Usage: make goose-gophermart-down-to id=20260924153000"; \
		exit 1; \
	fi; \
	set -a && . ./.env.gophermart && set +a && \
	goose down-to $(id)

# Перейти к конкретной новой версии (по ID) миграций для gophermart сервиса
goose-gophermart-up-to:
	@if [ -z "$(id)" ]; then \
		echo "Usage: make goose-gophermart-up-to id=20260924153000"; \
		exit 1; \
	fi; \
	set -a && . ./.env.gophermart && set +a && \
	goose up-to $(id)

# ------------------------------------------------------------------------------
# Accrual migrations
# ------------------------------------------------------------------------------

# Создать новую миграцию для accrual
goose-accrual-new:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make goose-accrual-new name=create_orders"; \
		exit 1; \
	fi; \
	set -a && . ./.env.accrual && set +a && \
	goose create $(name) sql

# Показать текущую версию миграции у accrual сервиса
goose-accrual-status:
	set -a && . ./.env.accrual && set +a && \
	goose status

# Применить последнюю версию миграций в accrual сервисе
goose-accrual-up:
	set -a && . ./.env.accrual && set +a && \
	goose up

# Применить следующую версию миграций для accrual сервиса
goose-accrual-up-by-one:
	set -a && . ./.env.accrual && set +a && \
	goose up-by-one

# Вернуть предыдущую версию миграций для accrual сервиса
goose-accrual-down:
	set -a && . ./.env.accrual && set +a && \
	goose down

# Вернуться к конкретной предыдущей версии (по ID) миграций для accrual сервиса
goose-accrual-down-to:
	@if [ -z "$(id)" ]; then \
		echo "Usage: make goose-accrual-down-to id=20260924153000"; \
		exit 1; \
	fi; \
	set -a && . ./.env.accrual && set +a && \
	goose down-to $(id)

# Перейти к конкретной новой версии (по ID) миграций для accrual сервиса
goose-accrual-up-to:
	@if [ -z "$(id)" ]; then \
		echo "Usage: make goose-accrual-up-to id=20260924153000"; \
		exit 1; \
	fi; \
	set -a && . ./.env.accrual && set +a && \
	goose up-to $(id)