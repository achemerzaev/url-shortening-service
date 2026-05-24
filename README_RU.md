# URL Shortening Service

Go-сервис для создания и управления короткими ссылками. В проекте есть JWT-авторизация, PostgreSQL, Redis-кеш, синхронизация Redis с PostgreSQL, Prometheus-метрики и Grafana.

## Стек

- Go, Gin
- PostgreSQL
- Redis
- Docker Compose
- Prometheus и Grafana

## Запуск

Создайте нужные secret-файлы в папке `secrets/`:

```text
db_user.txt
db_password.txt
db_name.txt
redis_password.txt
grafana_admin_user.txt
grafana_admin_password.txt
```

Запустите сервис:

```bash
docker compose up --build
```

API будет доступно на `http://localhost:8080`.

Полезные адреса:

- API: `http://localhost:8080`
- Метрики: `http://localhost:8080/metrics`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Примеры запросов

Регистрация:

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret"}'
```

Логин:

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret"}'
```

Оба запроса возвращают `access_token` и `refresh_token`. Access token нужен для защищенных эндпоинтов:

```bash
TOKEN="paste_access_token_here"
```

Создать короткую ссылку:

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```

Открыть короткую ссылку:

```bash
curl -i -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/shorten/abc123
```

Получить статистику:

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/shorten/abc123/stats
```

Обновить ссылку:

```bash
curl -X PUT http://localhost:8080/shorten/abc123 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev"}'
```

Удалить ссылку:

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/shorten/abc123
```

Обновить токены:

```bash
curl -X POST http://localhost:8080/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"paste_refresh_token_here"}'
```

## Тесты

Unit и e2e тесты:

```bash
go test ./...
```

Интеграционные тесты сервисного слоя:

```bash
go test -tags=integration ./internal/service
```

## Производительность

Результат нагрузочного теста:

- RPS: 2000
- Средняя задержка: 47.5 ms
- p90: 109 ms
- p95: 131 ms
