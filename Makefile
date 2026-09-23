include .env
export

# Запустить gophermart
run-gophermart:
	go run ./cmd/gophermart \
	-a $(RUN_ADDRESS) \
  	-config $(CONFIG_PATH) \
  	-d $(DATABASE_URI)

# Очистить кэш тестов
test-clean:
	go clean -testcache

# Запускает тесты во всем приложение
test:
	go test -v ./...

# Создать новую миграцию
goose-new:
	@if [ -z "$(name)" ]; then \
		echo "There is no name param. Example: make migrate-create name=name_of_migrate"; \
		exit 1; \
	fi;	\

	goose \
	create $(name) sql

goose-status:
	goose \
	status


goose-up:
	goose \
	up

goose-up-by-one:
	goose \
	up-by-one

goose-down:
	goose \
	down

# Перейти на старую версию миграции по её id
goose-down-to: 
	@if [ -z "$(id)" ]; then \
		echo "There is no migrate id. Example: make goose-down-to id=20170614145246"; \
		exit 1; \
	fi; \

	goose \
	down-to $(id)

# Перейти на новую версию миграции по её id
goose-up-to:
	@if [ -z "$(id)" ]; then \
		echo "There is no migrate id. Example: make goose-down-to id=20170614145246"; \
		exit 1; \
	fi;	\

	goose \
	up-to $(id)
