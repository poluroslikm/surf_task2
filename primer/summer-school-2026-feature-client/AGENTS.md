# AGENTS.md

## Repo State
- This repo currently contains BA/analysis artifacts for the SUP club project; BE/FE application code has not been created yet.
- The root `README.md` is still the GitLab template; use `01-analysis/README.md` and downstream specs instead.
- Preserve Russian requirement IDs and traceability links (`BR-*`, `FR-*`, `NFR-*`, `UC-*`, `SCR-*`, `BS-*`, `LOGIC-*`) when changing docs or generating code.

## Source Of Truth
- Start with `01-analysis/5-mobile-app-spec/README.md` for the current client app spec and screen inventory.
- Use `01-analysis/5-mobile-app-spec/*.md` and `01-analysis/5-mobile-app-spec/09_Логики/*.md` for screen flows, API operationIds, validation, states, and acceptance criteria.
- Use `01-analysis/api/redocly.yaml` plus domain files under `01-analysis/api/{auth,slots,bookings,profile,instructors}/` for API contracts.
- Use `01-analysis/2-requirements/` for requirements and `01-analysis/4-design/data-model.md` for resource model and invariants.
- Older references to `../api/openapi.yaml` are stale; there is no such file now. Use the multi-file OpenAPI domains registered in `redocly.yaml`.
- References to `rigla_network/...` in `_SCREEN_TEMPLATE.md` are template leftovers from another project; do not copy them into real specs or code.

## Target Implementation Constraints
- BE target stack: Go RESTful API, generated from or aligned to the existing OpenAPI contracts.
- BE performance testing target: k6 load tests sized for up to 300 concurrent users, especially booking/cancel concurrency.
- FE target stack: Kotlin Compose Multiplatform with Android, iOS, and Web targets.
- Do not introduce a different primary BE/FE stack unless the user explicitly asks.

## Backend Architecture Baseline
- Use a classic layered Go REST API architecture: HTTP transport -> application/use cases -> domain -> repositories/infrastructure.
- Keep the backend in a dedicated Go module, preferably under `backend/`, unless the user requests another layout.
- Use `cmd/api/main.go` only for composition: load config, initialize logger, database, repositories, services/use cases, HTTP router, middleware, and graceful shutdown.
- Put non-public application code under `internal/`; do not expose reusable packages unless there is a real external consumer.
- Recommended package layout:
  - `internal/config`: environment/config parsing and validation.
  - `internal/http`: router, middleware, request/response mapping, error mapping, generated OpenAPI handlers if used.
  - `internal/http/handlers`: thin HTTP handlers/controllers; no business rules here.
  - `internal/app` or `internal/usecase`: application services/use cases that orchestrate domain rules, repositories, transactions, idempotency, and external services.
  - `internal/domain`: entities, value objects, domain errors, and pure business rules; no database, HTTP, framework, or logging dependencies.
  - `internal/repository`: repository interfaces owned by use cases/domain needs, plus persistence implementations when the project is small.
  - `internal/storage/postgres`: SQL queries, migrations integration, transaction helpers, row locking, and PostgreSQL-specific code if persistence grows beyond simple repositories.
  - `internal/auth`: OTP/session/JWT/token logic and current-user context helpers.
  - `internal/clock`: injectable time source for booking/cancellation boundary tests.
  - `internal/observability`: structured logging, request IDs, metrics, tracing, health/readiness checks.
  - `migrations`: database schema migrations.
  - `tests` or package-local `*_test.go`: integration and contract tests where needed.
- Dependency direction must point inward: HTTP depends on use cases, use cases depend on domain abstractions/repository interfaces, infrastructure implements interfaces. Domain must not import HTTP, SQL drivers, generated API types, or framework packages.
- HTTP handlers must stay thin: validate/parse transport-level input, call one use case, convert result/errors to OpenAPI-compatible responses.
- Keep OpenAPI operationIds and schemas aligned with `01-analysis/api`; generated types are acceptable at the transport boundary, but do not leak them into domain models.
- Use explicit DTO mapping between OpenAPI/generated models, application commands/results, and domain entities.
- Put business invariants in use cases/domain, not in handlers or SQL alone. SQL constraints and transactions should protect the same invariants at persistence level.
- Booking and cancellation flows must run inside explicit database transactions and use row-level locking or equivalent concurrency control to prevent double booking, overbooking, and incorrect board inventory updates.
- Implement idempotent `createBooking` handling with persisted idempotency keys bound to the authenticated client and request payload semantics.
- Prefer context-aware methods (`context.Context`) for all request-scoped operations, repository calls, and external integrations.
- Use structured errors: domain/application errors should be typed or comparable and mapped once at the HTTP boundary to the documented API error responses.
- Do not use global mutable state for config, logger, database, clock, or services; pass dependencies through constructors.
- Prefer standard library plus small, established libraries. Avoid heavy frameworks unless explicitly justified.
- Tests should cover domain rules with unit tests, use cases with mocked/fake repositories where useful, and booking/cancel concurrency with integration tests against PostgreSQL or an equivalent test container.

## API Docs Commands
- API docs are the only npm project: run commands with `npm --prefix 01-analysis/api ...` from the repo root, or from `01-analysis/api` directly.
- First-time setup needs `npm --prefix 01-analysis/api install`; there is no lockfile and `package-lock.json` is intentionally ignored.
- Lint OpenAPI after contract changes: `npm --prefix 01-analysis/api run lint`.
- Bundle domain specs to ignored `dist/`: `npm --prefix 01-analysis/api run bundle`.
- Preview docs locally: `npm --prefix 01-analysis/api run preview`.
- Current domains are `auth`, `slots`, `bookings`, `profile`, and `instructors`; add new domains to `redocly.yaml` or Redocly will not lint/bundle them.

## MVP Scope Traps
- In scope: client role only, phone/OTP auth, slot list/filtering, slot card, booking self plus 1-2 guests, own/rental board choice, own bookings, cancel booking, profile, reminders/push registration where specified.
- Out of scope: instructor/admin UI, schedule CRUD, slot creation/editing, ratings, public reviews, online payment, auto-weather cancellation, loyalty, no-show handling.
- Slots, routes, and instructors are read-only projections from existing infrastructure; client code/API must not create or edit them.
- Payment is offline; the product only shows price and records booking details.

## Domain Invariants To Protect
- Booking must be atomic and must prevent double booking and overbooking under parallel requests.
- `createBooking` supports `Idempotency-Key`; use it for safe retry after network failure.
- `seats_count` is 1..3 and `rental_count` is 0..`seats_count`; own-board places consume group seats but not rental boards.
- Do not hardcode route caps or board inventory in FE; use slot/route data even though docs mention 8/12 and 12 boards.
- Late cancellation is `< 2h` before slot start and does not free seats or rental boards; exactly `2h` is early cancellation and frees them.
- "Past" is derived from `slot.start_at`, not a stored status. Booking status is `active`, `cancelled`, or `late_cancel`; slot status is `scheduled` or `cancelled`.
- Client data access is only to the current user's profile/bookings; cross-client access must return forbidden/unauthorized behavior per common API responses.
