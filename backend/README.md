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

## Проверено вживую (не только статически)

В среде разработки не было предустановленного Go/Docker, но нашёлся сетевой доступ — портативный
Go 1.26.4 и портативная сборка PostgreSQL 16.4 (EDB zip) были скачаны, проверены по SHA-256/MD5 и
распакованы во временный каталог, без установки в систему. На этом стенде реально проверено:

- `go mod tidy`, `go build ./...`, `go vet ./...`, `gofmt -l .` — все чисто, без единой ошибки;
  зависимости в `go.mod`/`go.sum` (chi v5.1.0, pgx v5.6.0, x/crypto v0.26.0) резолвятся и
  собираются как есть.
- Локальный `postgres` (initdb + pg_ctl start), обе миграции применены как есть (`00001_init.sql`,
  `00002_seed_dev.sql`) — ни один `CHECK`/индекс/`CREATE EXTENSION` не потребовал правок.
- Скомпилированный бинарник запущен против этой БД и опрошен вживую: `GET /healthz` (200),
  `POST /auth/register` (201 + сохранённый токен), повторная регистрация того же email (409,
  `code: bad_request` — именно так, как задокументирован этот нюанс в
  `5-mobile-app-spec/09_Логики/LOGIC-001_...md`), короткий пароль (400), `POST /auth/login`
  верный/неверный пароль (200/401), `GET /slots` без и с токеном (401/200, дефолтные 7 дней,
  верно отсортировано, все 3 seed-слота на месте), `GET /slots?date_to=...` (200), `GET
  /slots/{unknown}` (404), `POST /auth/logout` (204) и повторное использование того же токена
  после логаута (401 — сессия действительно инвалидируется).
- Без правок кода: с первого запуска всё отработало так, как описано в контракте.

**Не проверено даже так:** `BE-06`+ (создание/отмена записи, оценка, атомарность `free_seats`
под конкурентной нагрузкой), сборка Docker-образа (`Dockerfile`) — Docker в среде так и нет,
автоматические тесты (`BE-11`/`BE-12` не начаты).

## Известные ограничения окружения, в котором писался код

Go/Docker/`psql` не были предустановлены — весь код был написан вручную, без возможности собрать
его в момент написания. Раздел выше подтверждает, что после появления сети это удалось исправить
постфактум: код скомпилировался и заработал с первого раза без правок. Тем не менее `Dockerfile`/
`compose.yaml` по-прежнему не проверялись — Docker-движка в среде нет.

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
