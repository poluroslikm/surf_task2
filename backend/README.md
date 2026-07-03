# Шеф-стол — backend (Go)

Реализация REST API по контракту `../api/{auth,slots,bookings,common}` и плану
`../6-development/BE_IMPLEMENTATION_PLAN.md`.

## Реализовано в этой итерации

- **BE-00** — каркас (`cmd/api`, `internal/{config,domain,service,storage,httpapi}`).
- **BE-02** (частично) — request id / recovery / access-log middleware, единый error mapper
  по `api/common/models.yaml` (коды `bad_request`, `unauthorized`, `forbidden`, `not_found`,
  `internal_error`).
- **BE-03** — миграция `migrations/00001_init.sql` (`clients`, `auth_sessions`, `programs`,
  `chefs`, `slots`, `bookings`) и dev-seed `migrations/00002_seed_dev.sql`.
- **BE-04** — `register`, `login`, `logout`: email/password, `bcrypt`, непрозрачный
  bearer-токен (`auth_sessions.token_hash`) — **не JWT**, как и зафиксировано в
  `auth/models.yaml` (`bearerFormat: opaque-token`).
- **BE-05** — `listSlots` (окно по умолчанию 7 дней, FR-3/FR-4/R-027; заполненные и
  отменённые студией слоты не скрываются — FR-6a/FR-18), `getSlot`.

## Не реализовано

- **BE-01** — генерация типов через `oapi-codegen` не выполнялась: в среде, где писался
  этот код, нет установленного Go. DTO в `internal/httpapi/dto/dto.go` написаны вручную по
  образцу `api/*/models.yaml`. Замените на настоящую генерацию, как только появится
  тулчейн — иначе контракт и код могут разойтись незаметно.
- **BE-03** — `sqlc` тоже не запускался; репозитории в `internal/storage/` — обычный `pgx`
  с SQL-строками, без кодогенерации. Схема при этом полностью соответствует финальному
  дизайну из `BE_IMPLEMENTATION_PLAN.md` → BE-03 (никаких `seats_count`, лимитов проката,
  `capacity_cap` на программе).
- **BE-06…BE-15** — создание/список/детали/отмена записей, оценка шефа, push-инфраструктура
  (FR-21/FR-22), контрактные тесты, Go-тесты, k6, финальная сверка — не начаты.
- Per-client scoping (кто есть текущий клиент) нигде не используется — он не нужен ни
  `auth`, ни `slots` эндпоинтам, но потребуется для `bookings` (BE-07+, NFR-8).
- **Ничего не скомпилировано и не запускалось.**

## Известные ограничения окружения, в котором писался код

В песочнице, где создавался этот код, нет Go, Docker и `psql` — весь код написан вручную,
без единого прогона `go build`/`go vet`/`go test`, и `go.sum` тоже не существует (зависимости
ни разу не разрешались). Возможны опечатки в именах пакетов/сигнатурах, которые
обнаружатся только при первой сборке. Перед тем как доверять этому коду:

```
go mod tidy
go build ./...
go vet ./...
```

## Запуск локально

```
cp .env.example .env
docker compose up -d db
# установить goose (https://github.com/pressly/goose) и применить миграции:
goose -dir migrations postgres "$DATABASE_URL" up
go mod tidy
go run ./cmd/api
```

## Переменные окружения

| Переменная | Обязательна | По умолчанию | Смысл |
|---|---|---|---|
| `DATABASE_URL` | да | — | DSN PostgreSQL |
| `HTTP_ADDR` | нет | `:8080` | адрес HTTP-сервера |
| `SESSION_TTL_HOURS` | нет | `720` (30 дней) | срок жизни bearer-токена |

## Эндпоинты

| Метод | Путь | operationId | Auth |
|---|---|---|---|
| POST | `/auth/register` | `register` | нет |
| POST | `/auth/login` | `login` | нет |
| POST | `/auth/logout` | `logout` | да |
| GET | `/slots` | `listSlots` | да |
| GET | `/slots/{slotId}` | `getSlot` | да |
| GET | `/healthz` | — | нет |
