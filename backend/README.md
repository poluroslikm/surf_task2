# Шеф-стол — backend (Go)

Реализация REST API по контракту `../api/{auth,slots,bookings,common}` и плану
`../6-development/BE_IMPLEMENTATION_PLAN.md`.

## Реализовано

- **BE-00** — каркас (`cmd/api`, `internal/{config,domain,service,storage,httpapi}`).
- **BE-02/BE-11** — request id / recovery / access-log / CORS middleware, единый error mapper по
  `api/common/models.yaml` (весь enum `Error.code`: `bad_request`, `unauthorized`, `forbidden`,
  `not_found`, `internal_error`, `slot_full`, `slot_cancelled`, `slot_started`,
  `already_cancelled`, `not_ratable`, `already_rated`).
- **BE-03** — миграция `migrations/00001_init.sql` (`clients`, `auth_sessions`, `programs`,
  `chefs`, `slots`, `bookings`) и dev-seed `migrations/00002_seed_dev.sql`.
- **BE-04** — `register`, `login`, `logout`: email/password, `bcrypt`, непрозрачный
  bearer-токен (`auth_sessions.token_hash`) — **не JWT** (`auth/models.yaml` →
  `bearerFormat: opaque-token`). `RequireAuth` прокидывает аутентифицированного клиента в
  контекст запроса (`ClientFromContext`) — нужно для per-client scoping в `bookings` (NFR-8).
- **BE-05** — `listSlots` (окно по умолчанию 7 дней, FR-3/FR-4/R-027; заполненные и
  отменённые студией слоты не скрываются — FR-6a/FR-18), `getSlot`.
- **BE-06…BE-09** — весь домен `bookings`:
  - `createBooking` — атомарно (одна транзакция: CAS-декремент `free_seats` + `INSERT`), снимок
    `price_total`, `free_cancellation_until` вычисляется на чтении (`slot.start_at − 24ч`), не
    хранится. Неизвестный `slot_id` → `400` (у `createBooking` в контракте нет `404`, в отличие
    от `getSlot`); отменённый слот → `410 slot_cancelled`; нет мест → `409 slot_full` с
    `details.available_seats`. Повторная активная запись того же клиента на тот же слот
    (`bookings_active_client_slot_uidx`) → `409 bad_request` — в контракте нет отдельного кода
    для этого случая, решено по аналогии с `ErrEmailTaken`.
  - `listBookings`/`getBooking` — только свои записи (`403` на чужие, NFR-8).
  - `cancelBooking` — правило 24 часов (граница включительна), после старта класса — `422
    slot_started`; и проверка владения, и переход статуса — под `SELECT ... FOR UPDATE` в одной
    транзакции, чтобы параллельные отмены не могли обе вернуть место в слот.
  - `submitRating` — только для прошедших активных записей без ранее поставленной оценки;
    порядок проверок: не найдена → чужая → уже оценена → не подходит по времени/статусу →
    диапазон `1–5`.
- **BE-12 (частично)** — юнит-тесты `auth`/`slots` на сервисном слое (фейковые репозитории) +
  HTTP-тесты через `httptest` поверх реальных хендлеров/роутера.

## Не реализовано

- **BE-01** — генерация типов через `oapi-codegen` не выполнялась (изначально — из-за
  отсутствия Go в среде на тот момент; сейчас Go есть, но генерацию так и не подключили). DTO в
  `internal/httpapi/dto/dto.go` написаны вручную по образцу `api/*/models.yaml` — контракт и код
  могут разойтись незаметно, сверяйте вручную при изменении `api/`.
- **BE-03** — `sqlc` тоже не подключён; репозитории — обычный `pgx` с SQL-строками.
- **BE-10** — push-инфраструктура (FR-21/FR-22): регистрация push-подписки не имеет эндпоинта в
  контракте (открытый вопрос, см. `api/README.md`); ничего не реализовано.
- **BE-12 (bookings)** — новый домен пока без автотестов, только живая ручная/скриптовая
  проверка (см. ниже). Concurrency-тест для `createBooking` был прогнан вживую один раз, но не
  оформлен как повторяемый Go-тест.
- **BE-13** — k6/нагрузочное тестирование не начато.
- **submitRating: success/`already_rated`** — проверены только по коду, не вживую: в seed-данных
  нет слота с `start_at` в прошлом, а фабриковать его вручную агент, писавший эту часть, не стал
  (см. «Проверено вживую» ниже). `not_ratable`-отказы (будущий слот, отменённая запись) — проверены
  вживую.

## Проверено вживую

Бэкенд сейчас реально работает на машине пользователя: настоящий (не портативный) PostgreSQL 17
через `winget`, настоящий Go через `winget`, реальный запущенный клиент на React обращается к
этому бэкенду через браузер. Основной поток (регистрация → список слотов) подтверждён визуально
самим пользователем.

Дополнительно, для `bookings`, отдельный прогон (агент, отдельный порт `:8081`, не пересекаясь с
работающим на `:8080` инстансом пользователя; только новые тестовые клиенты `booking-agent-*`,
существующие seed-данные не тронуты):

- `createBooking` (свой/прокатный инвентарь), корректные `price_total`/`free_cancellation_until`.
- Повторная активная запись на тот же слот → `409 bad_request`.
- Полностью занятый слот → `409 slot_full` с `details.available_seats: 0`.
- Неизвестный `slot_id` → `400`; неверный `equipment_choice` → `400`.
- Ранняя (`≥24ч`) отмена → `cancelled`, место возвращено; поздняя (`<24ч`) → `late_cancel`, место
  не возвращено; повторная отмена → `409 already_cancelled`.
- Оценка будущего/отменённого класса → `422 not_ratable`.
- Доступ к чужой записи (`getBooking`/`cancelBooking`) → `403`; неизвестная запись → `404`.
- **Конкурентный тест (NFR-6):** 8 параллельных `createBooking` на слот с 3 свободными местами →
  ровно 3×`201` и 5×`409 slot_full`, итоговый `free_seats = 0`, без ухода в минус и без
  оверселлинга. Тестовые брони отменены после проверки, слот возвращён к исходному состоянию.

`go build ./...`, `go vet ./...`, `go test ./...` — чисто на текущем коде (auth+slots тесты
проходят; bookings без автотестов, см. выше).

## Запуск локально

```
docker compose up -d db
# или реальный PostgreSQL (см. историю разработки — здесь ставился через winget)
goose -dir migrations postgres "$DATABASE_URL" up
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
| POST | `/bookings` | `createBooking` | да |
| GET | `/bookings` | `listBookings` | да |
| GET | `/bookings/{bookingId}` | `getBooking` | да |
| POST | `/bookings/{bookingId}/cancel` | `cancelBooking` | да |
| POST | `/bookings/{bookingId}/rating` | `submitRating` | да |
| GET | `/healthz` | — | нет |
