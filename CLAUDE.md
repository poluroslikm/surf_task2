# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

This is an **analyst-to-implementation pipeline** for «Шеф-стол» — a client-only PWA + REST API for a
cooking-studio booking system (replaces manual WhatsApp/Google-Sheets booking). There is **no
application code in this repository yet**: the root holds staged analysis/design documentation, a
multi-file OpenAPI contract, a backend implementation plan, and a large reference example used only
as a structural template.

## Folder structure and reading order

Numbered folders are sequential pipeline stages; each stage's documents cite their sources from the
previous stage (see the "Источники" links at the top of most files). When you need to know *why* a
requirement or field exists, follow the citations backward through stages rather than guessing:

```
0-customer-brief/  → raw customer brief (intentionally incomplete/contradictory)
1-elicitation/     → clarifying Q&A + domain description (01-questions.md, 02-domain.md)
2-requirements/    → business/functional/non-functional requirements, use cases, user stories
                      (5 separate files — never merge them)
3-design-brief/    → per-screen/bottom-sheet draft requirements for a designer (00-foundations.md
                      holds cross-cutting UI decisions)
4-design/          → ER data model + Create-Booking sequence diagram
api/               → OpenAPI 3.0.3 contract, split by domain (see below) — this is the canonical
                      data contract, not 4-design/data-model.md
5-mobile-app-spec/ → final per-screen/bottom-sheet spec (SCR-*.md, BS-*.md) + reusable cross-screen
                      logic docs in 09_Логики/ — entry point is 5-mobile-app-spec/README.md
6-development/     → implementation plans (checklist-style, iterative). Currently just
                      BE_IMPLEMENTATION_PLAN.md; nothing in it is implemented yet.
inf/, promt/       → meta: the analyst methodology and prompts this whole pipeline was built with
primer/            → a full reference implementation (Go backend + Kotlin CMP client) for a
                      DIFFERENT but structurally similar course example — see warning below
```

## The `primer/` reference — template only, not a source of truth

`primer/summer-school-2026-feature-client/` is a complete worked example for a **different domain**
(a river sup-boarding club, "Волна": instructors/routes/boards) built with the same methodology. It's
useful for folder layout and tech-stack patterns (Go + chi + oapi-codegen + sqlc + goose + Postgres on
the backend, Kotlin Compose Multiplatform + MVI on the client), but several of its business rules and
schema choices **do not apply to «Шеф-стол» and must not be copied verbatim** — this has already
caused real mismatches once (see `6-development/BE_IMPLEMENTATION_PLAN.md` intro and BE-03/BE-08 for
the full list). The recurring pattern: primer's numbers/fields are specific to sabboarding and were
carried over uncritically before being caught — always check the current repo's own `2-requirements/`
and `api/` before trusting anything from `primer/`.

## Commands

Root-level: no build/lint/test yet — the only executable artifact right now is the OpenAPI contract.

Validate an OpenAPI domain file (each `api.yaml` is a self-contained document with relative `$ref`s):

```
npx --yes @redocly/cli@latest lint api/<domain>/api.yaml
```

(`api/auth/api.yaml` is pre-approved in `.claude/settings.local.json`; the same command works for
`slots`, `bookings`, `common`.)

`primer/summer-school-2026-feature-client/backend/` and `.../client/` are self-contained Go and
Gradle/Kotlin projects with their own `Makefile` / `gradlew` — only relevant if you're actively
consulting or running the reference example, not part of this repo's own build.

## API contract structure (`api/`)

Split by domain, not one file — see `api/README.md` for the full rationale. 5 domains exist on
purpose: **auth**, **slots**, **bookings**, **push**, **common**. There is no `profile` domain (no
client profile view/edit beyond login) and no standalone `programs`/`chefs` domain — `Program` and
`Chef` are read-only objects nested inside `Slot`, not independently listable resources. `push`
(added later, closing the gap noted below) is a real resource with its own lifecycle — a client
registers a Web Push subscription — not a read-only projection, same rationale that keeps
`bookings` its own domain instead of a `slots` sub-path. Rating is a sub-path
of bookings (`POST /bookings/{bookingId}/rating`), not its own resource.

## Domain invariants that require cross-file synthesis

These are easy to get wrong because no single file states them plainly — they only become clear by
reading `2-requirements/`, `4-design/data-model.md`, and `api/*/models.yaml` together:

- **Auth is email + password**, not phone/OTP. The bearer token is **opaque**
  (`auth/models.yaml` → `bearerFormat: opaque-token`) — explicitly not JWT.
- **Cancellation threshold is 24 hours**, not a shorter window. `free_cancellation_until =
  slot.start_at − 24h` is computed and returned by the server; it must never be recomputed
  client-side.
- **Equipment rental fund is unconstrained** (FR-9) — no capacity/counter field exists anywhere in
  the model for it. `equipment_choice` is a plain `own`/`rental` enum on the booking, never blocked.
- **One booking = one seat.** There is no `seats_count` / group-booking concept anywhere in this
  project (open domain question #1 in `1-elicitation/02-domain.md`, resolved to the simple MVP
  variant).
- **`Booking.status` has 4 values**: `active`, `cancelled`, `late_cancel`, `cancelled_by_studio`. The
  last one is a server-initiated transition (studio force-majeure cancellation of a `Slot`), distinct
  from client-initiated cancellation, and blocks re-booking that slot for everyone.
- **`price_total` is a frozen snapshot** of `slot.price` at booking creation time — never
  recalculated if the slot's price changes later.
- **Rating is embedded 1:1 in `Booking`**, not a standalone resource — settable at most once, only
  when `slot.start_at` is in the past and `booking.status = active`.
- Seat-capacity limits (8 for complex-equipment classes / 12 studio-wide) are backend business logic
  living only in `Slot.total_seats`/`free_seats` — never re-derive or hardcode this rule client-side
  (FR-12), and don't tie it to `Program.difficulty` via a rigid formula (unlike the primer's
  `routes_capacity_chk`, which does exactly that for a different, unrelated reason).

## Known open contract gaps

Documented but unresolved — don't silently invent an answer, they need an explicit decision that
also updates `api/`:

- **Equipment rental pricing** (`rental_price`) is not part of the contract — mentioned as a possible
  backend field but no FR/US defines how it would affect `price_total`.
- `createBooking` has no documented `422 slot_started` response (unlike `cancelBooking` and
  `submitRating`, which do) — booking a slot that has already started is not explicitly guarded by
  the contract.
