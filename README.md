# gophermart-loyal-service

### Запустить приложение

```bash
go go run ./cmd/gophermart \
  -config config.yaml \
  -d postgres://postgres:password@localhost:5432/database
```

или

```bash
make run-app
```


### Запустить тесты (интеграционные и юниты)

```bash
make run-test
```
