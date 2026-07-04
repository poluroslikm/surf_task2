# План реализации FE (веб-клиента) для «Шеф-стол»

> **Смена стека (см. историю правок).** Изначально этот план и реализация были на Kotlin
> Compose Multiplatform (wasmJs). После воспроизводимого краша компилятора Kotlin/Wasm на Windows
> (подтверждённые открытые баги JetBrains — **KT-79120**, **KT-79144**: «Symbol for Any not found»,
> «It happens only... on Js and WasmJs (JVM and Native are fine)» — фикс не помог даже на
> обновлённых версиях Kotlin 2.2.20/2.2.21), решено отказаться от Kotlin/Compose и переписать
> клиент на обычном вебовом стеке — **React + TypeScript + Vite**. Заодно это устраняет расхождение
> с исходным ТЗ: `5-mobile-app-spec/README.md` (NFR-3) с самого начала требовал именно **PWA**,
> а не Kotlin/CMP — переход на Kotlin был отдельным отклонением от ТЗ, теперь снятым.
>
> Вся содержательная часть плана (декомпозиция экранов, доменные правила, отличия от референса
> `primer/`) не пересматривалась — эти разделы про продукт, а не про технологию.

## Отличия домена от референса (не копировать бездумно)

| Аспект | Референс («Волна») | Наш проект («Шеф-стол») |
|---|---|---|
| Авторизация | Телефон + OTP, 3 шага (phone/otp/name) | Email + пароль, 1 шаг, вкладки «Вход»/«Регистрация» (SCR-001) |
| Бронь | Группа 1–3 места + прокат досок с лимитом | Одна запись — одно место, `equipment_choice: own\|rental` без лимита (FR-9) |
| Порог отмены | 2 часа | **24 часа** |
| Статусы брони | `active/cancelled/late_cancel` | + `cancelled_by_studio` (4 значения) |
| Домены API | auth, profile, slots, instructors, bookings | auth, slots, bookings, common (без `profile`, без `instructors` — Chef вложен в Slot) |
| Профиль | Полный экран SCR-007 (имя, смена телефона, удаление аккаунта) | Нет профиля вообще — только `BS-004` (выход), см. `feature-list.md` §6 |
| Карта / маршрут | `BS-004` карта маршрута, `LOGIC-006`, `Route.geometry` | Не существует — класс проходит в студии, адреса/карты нет в контракте |
| Оценка | Не входит в MVP референса | **`SCR-007` Оценка шефа** — есть у нас (FR-19/FR-20), которой нет в референсе |
| Логик (09_Логики) | 8 (включая цену, доступность, карту) | **5**: `LOGIC-001` авторизация, `LOGIC-002` отмена 24ч, `LOGIC-003` фильтрация дат, `LOGIC-004` push-разрешение, `LOGIC-005` паттерн состояний экрана |
| Цена | Клиент считает `price * seats + rentalPrice * rental` (`LOGIC-003` референса) | **Клиент никогда не считает цену** — `price_total` всегда из API, снимок на момент записи (FR-23, P3) |
| Idempotency-Key | Есть в контракте `createBooking` | **Нет в нашем контракте** — открытый пробел, см. `BE_IMPLEMENTATION_PLAN.md` → BE-06 |

## TOC / Todo реализации

- [x] [FE-00. Создать каркас Vite+React+TS проекта](#fe-00-создать-каркас-vitereactts-проекта)
- [ ] [FE-01. Ядро: состояния экрана, конфиг, ошибки](#fe-01-ядро-состояния-экрана-конфиг-ошибки)
- [x] [FE-02. Сетевой слой и типы по контракту](#fe-02-сетевой-слой-и-типы-по-контракту)
- [x] [FE-03. Хранилище токена и сессии](#fe-03-хранилище-токена-и-сессии)
- [x] [FE-04. Экран авторизации SCR-001 (LOGIC-001)](#fe-04-экран-авторизации-scr-001-logic-001)
- [ ] [FE-05. Корневая навигация и таб-бар](#fe-05-корневая-навигация-и-таб-бар)
- [x] [FE-06. Список слотов SCR-002 + фильтр дат BS-001 (LOGIC-003, LOGIC-005)](#fe-06-список-слотов-scr-002--фильтр-дат-bs-001-logic-003-logic-005)
- [ ] [FE-07. Карточка слота SCR-003](#fe-07-карточка-слота-scr-003)
- [ ] [FE-08. Оформление записи SCR-004 + подтверждение BS-002 (LOGIC-004)](#fe-08-оформление-записи-scr-004--подтверждение-bs-002-logic-004)
- [ ] [FE-09. Мои бронирования SCR-005](#fe-09-мои-бронирования-scr-005)
- [ ] [FE-10. Детали брони SCR-006 + отмена BS-003 (LOGIC-002)](#fe-10-детали-брони-scr-006--отмена-bs-003-logic-002)
- [ ] [FE-11. Оценка шефа SCR-007](#fe-11-оценка-шефа-scr-007)
- [ ] [FE-12. Аккаунт / выход BS-004](#fe-12-аккаунт--выход-bs-004)
- [ ] [FE-13. Тема оформления](#fe-13-тема-оформления)
- [ ] [FE-14. Тесты](#fe-14-тесты)
- [ ] [FE-15. Финальная проверка готовности FE](#fe-15-финальная-проверка-готовности-fe)

FE-00, FE-02 (без кодогенерации типов), FE-03, FE-04, FE-06 (без BS-001/фикс. таб-бара с реальной
навигацией) сделаны и **реально собраны и протестированы** (`npm run build`, см. `client/README.md`
→ «Проверено вживую») — в отличие от Kotlin-попытки, которая не компилировалась вообще.

## Стек приложения

- **React + TypeScript + Vite** — не подбирались вручную, а сгенерированы официальным
  скаффолдером `npm create vite@latest -- --template react-ts`: версии (React 19.2, Vite 8.1,
  TypeScript 6.0 на момент написания) берутся из реального шаблона, а не из памяти — именно
  ручной подбор версий стал причиной многочасового провала с Kotlin/Gradle.
- **vite-plugin-pwa** — манифест + service worker, закрывает NFR-3 (PWA) по-настоящему.
- Обычный `fetch` — без axios/ky, минимум зависимостей.
- Без React Router и без стейт-менеджер-библиотек, пока экранов всего 2 — простое условное
  переключение `Auth ↔ Slots` по наличию токена (`App.tsx`). Полноценная навигация с таб-баром —
  отдельный пункт (`FE-05`), намеренно отложен, не переусложняем раньше времени.
- Простые кастомные React-хуки на экран (`useAuthScreen`, `useSlotsScreen`) вместо MVI-хранилищ —
  тот же паттерн состояний (см. ниже), просто без отдельной библиотеки на 2 экрана.

## Структура

```
client/src/
  core/
    config.ts       — API_BASE_URL
    screenState.ts  — ScreenState<T> (LOGIC-005)
    errors.ts       — ErrorCode, ApiError, NetworkError/ApiRequestError, сквозные тексты
  api/
    types.ts        — DTO 1:1 с api/auth/models.yaml, api/slots/models.yaml
    httpClient.ts    — fetch-обёртка: Bearer из session, маппинг ошибок, 401 → session.clear()
    authApi.ts       — register, login, logout
    slotsApi.ts      — listSlots, getSlot
  session/
    session.ts        — localStorage-токен (LOGIC-001), Flow-подобная подписка через слушателей
  screens/
    auth/AuthScreen.tsx, useAuthScreen.ts
    slots/SlotsScreen.tsx, useSlotsScreen.ts
  App.tsx            — условное Auth ↔ Slots по сессии
```

Пакетов `profile`, `map`, `notifications` как отдельных доменов нет (см. таблицу отличий выше).

## Domain-типы (TS, зеркалят Kotlin-версию из истории правок 1:1)

```ts
type ProgramDifficulty = 'novice' | 'experienced'
interface Program { id: string; name: string; difficulty: ProgramDifficulty; photo_url: string; ingredients: string[]; allergens: string[] }
interface Chef { id: string; name: string }
type SlotStatus = 'scheduled' | 'cancelled'
interface Slot { id: string; program: Program; chef: Chef; start_at: string; total_seats: number; free_seats: number; price: number; status: SlotStatus; cancellation_reason?: string | null }
type EquipmentChoice = 'own' | 'rental'
type BookingStatus = 'active' | 'cancelled' | 'late_cancel' | 'cancelled_by_studio'
```

`Booking`/`Rating` появятся в FE-08+ вместе с доменом `bookings`.

## Паттерн состояний экрана (прямое отражение LOGIC-005)

```ts
type ScreenState<T> =
  | { kind: 'loading' }
  | { kind: 'content'; value: T; refreshing: boolean }
  | { kind: 'empty'; reason: string }
  | { kind: 'error'; message: string }
```

Правила (буквально из `LOGIC-005`, не расширять сверх документированного):
- Первичная загрузка: `loading -> content|empty|error`.
- Обновление (аналог pull-to-refresh — кнопка «Обновить», см. «Не реализовано» в
  `client/README.md`): контент/empty сохраняется, `refreshing=true`; провал — **только** снек
  (`snack`-состояние хука), `error` не показывается (LOGIC-005 AC-007).
- Действие (`createBooking`/`cancelBooking`/`submitRating`, будущие FE-08+) не имеет своего
  Loading/Error — только `submitting`-флаг и снек по результату.
- `401` на любом запросе, кроме `login`, — централизованно в `httpClient` очищает токен;
  экран/хук ничего специфичного не делает, переключение на `AuthScreen` происходит в `App.tsx`
  реактивно через подписку на `session`.

## Data layer

- `AuthApi`: `register`, `login`, `logout`.
- `SlotsApi`: `listSlots`, `getSlot`.
- В FE-08+ добавится `BookingsApi`: `createBooking`, `listBookings`, `getBooking`,
  `cancelBooking`, `submitRating` — без `ProfileApi`/`InstructorsApi`.

Типы ошибок — по факту `common/models.yaml → Error.code` enum (список воспроизведён 1:1 с
`backend/internal/httpapi/errors.go`, не изобретать новые коды):
`bad_request`, `unauthorized`, `forbidden`, `not_found`, `internal_error`,
`slot_full`, `slot_cancelled`, `slot_started`, `already_cancelled`, `not_ratable`, `already_rated`.

**Idempotency-Key — не реализуется**: контракт `createBooking` его не принимает (см. таблицу
отличий выше) — не изобретать сверх задокументированного `CreateBookingRequest`.

---

## Декомпозиция FE

### FE-00. Создать каркас Vite+React+TS проекта ✅

Сделано: `npm create vite@latest client -- --template react-ts`, добавлен `vite-plugin-pwa`,
удалены demo-компоненты скаффолда.

Готово, когда: `npm run build` проходит — **проверено вживую**.

### FE-01. Ядро: состояния экрана, конфиг, ошибки

Частично сделано в рамках FE-02/FE-06 (`core/screenState.ts`, `core/errors.ts`,
`core/config.ts`) — отдельный юнит-тест на маппинг `Error{code,message}` → сообщение для всех
кодов из `common/models.yaml` ещё не написан (см. FE-14).

### FE-02. Сетевой слой и типы по контракту ✅ (без кодогенерации)

Сделано: DTO в `api/types.ts` вручную по образцу `api/auth/models.yaml`/`api/slots/models.yaml`
(аналогично `backend/internal/httpapi/dto` — генератора клиента из OpenAPI не подключали);
`authApi.ts`/`slotsApi.ts` — тонкие обёртки над `httpClient` по `operationId`.

Готово, когда: типы совпадают с контрактом, `tsc --noEmit` проходит — **проверено вживую**.

### FE-03. Хранилище токена и сессии ✅

Сделано: `session.ts` поверх `localStorage` (LOGIC-001 — не `sessionStorage`), подписка через
простой набор слушателей (аналог `Flow<Boolean>`), централизованная очистка токена на `401`
(кроме самого `login`) — в `httpClient.ts`.

Не сделано: юнит-тест «токен переживает пересоздание модуля» (FE-14).

### FE-04. Экран авторизации SCR-001 (LOGIC-001) ✅

Сделано по `5-mobile-app-spec/SCR-001-auth.md` буквально: вкладки Вход/Регистрация (Вход —
активна по умолчанию), email/password валидация по blur (E1/E2, E2 только на «Регистрации»),
disabled/submitting-CTA, нюанс «409 при регистрации = `code: bad_request`, различать по
HTTP-статусу» (не по `code`), `401` на `login` → очистка только пароля + фокус на поле пароля,
сохранение токена и переход на `SlotsScreen` без промежуточного экрана успеха.

Готово, когда: все AC из `SCR-001-auth.md` покрыты тестами — тесты (FE-14) ещё не написаны, само
поведение реализовано и собирается.

### FE-05. Корневая навигация и таб-бар

Не начато. Сейчас `App.tsx` — простое `if (authenticated)`-переключение без роутера; вкладка
«Мои записи» и иконки фильтра/аккаунта на `SlotsScreen` отрисованы по дизайну, но `onClick`
не ведут никуда (соответствующих экранов ещё нет).

### FE-06. Список слотов SCR-002 + фильтр дат BS-001 (LOGIC-003, LOGIC-005) ✅ (без BS-001 UI)

Сделано по `5-mobile-app-spec/SCR-002-slot-list.md`: `ScreenState` loading/content/empty/error,
бейдж «Отменён студией» с приоритетом над «Мест нет» (независимо от `free_seats`), тег
сложности «Новичковый»/«Для опытных», два варианта Empty («Выбрать другой период» для дефолтного
периода / «Сбросить фильтр» для применённого), хук поддерживает `resetDateFilter` по правилу
LOGIC-003 (клиент никогда не передаёт `date_from`, только опционально `date_to`).

Не сделано: сама шторка `BS-001` (поле «Показать по», кнопки «Применить»/«Сбросить») — только
контракт хука (`resetDateFilter`); чип периода показывает сырой ISO-timestamp, а не «3–20 июля»;
`refresh` — обычная кнопка, а не жест pull-to-refresh.

### FE-07. Карточка слота SCR-003

Не начато.

### FE-08. Оформление записи SCR-004 + подтверждение BS-002 (LOGIC-004)

Не начато. При реализации: цена — всегда `slot.price` как есть (нет группы мест/проката с
лимитом); push-разрешение — `Notification.requestPermission()` (Web Push API, не нативный SDK,
NFR-3), показывается один раз при первой успешной записи на устройстве.

### FE-09. Мои бронирования SCR-005

Не начато. Группировка «Предстоящие»/«Прошедшие» — на клиенте по `slot.start_at`, не по
`status` (такого статуса `past` не существует).

### FE-10. Детали брони SCR-006 + отмена BS-003 (LOGIC-002)

Не начато. Превью на BS-003 — по `booking.free_cancellation_until` (уже посчитан сервером):
**≥ 24ч до старта** → бесплатно, **< 24ч** → предупреждение без штрафа (не 2 часа, как в
референсе). Финальный статус — только из ответа `cancelBooking`.

### FE-11. Оценка шефа SCR-007

Не начато. `422 not_ratable`/`409 already_rated` — снек, без краша формы; редактирование ранее
поставленной оценки не предусмотрено (UC-4 E2).

### FE-12. Аккаунт / выход BS-004

Не начато. Минимальная шторка: `logout` → `session.clear()` → реактивный переход на `AuthScreen`.

### FE-13. Тема оформления

Частично: нейтральные CSS-токены в `index.css` (бренд не зафиксирован Figma-макетом, см.
`00-foundations.md`) — не выделены в отдельный модуль темы, захардкожены в глобальный CSS-файл;
вынести в CSS custom properties/токены отдельным пунктом при добавлении новых экранов.

### FE-14. Тесты

Не начато. Нужны: юнит-тесты границ LOGIC-002 (24ч ровно, ±1с), LOGIC-003 (граница периода),
маппинг `Error.code` → сообщение, тесты хуков `useAuthScreen`/`useSlotsScreen` (например через
Vitest + Testing Library — не подключены).

### FE-15. Финальная проверка готовности FE

Не начато — преждевременно, пока не реализованы FE-07…FE-12.
