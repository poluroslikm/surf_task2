# Шеф-стол — клиент (React + TypeScript + Vite)

Реализация по `../6-development/FE_IMPLEMENTATION_PLAN.md`, контракту `../api/{auth,slots,bookings,push,common}`
и бэкенду `../backend/`.

> **Смена стека.** Изначально клиент писался на Kotlin Compose Multiplatform (wasmJs). После
> воспроизводимого краша компилятора Kotlin/Wasm на Windows — подтверждённые открытые баги
> JetBrains **KT-79120** и **KT-79144** («Symbol for Any not found», «It happens only... on Js and
> WasmJs (JVM and Native are fine)»), фикс не помог даже на Kotlin 2.2.20/2.2.21 — стек сменили на
> обычный веб (React/TS/Vite). Заодно это устраняет расхождение с ТЗ: NFR-3 всегда требовал именно
> PWA, а не Kotlin/CMP.

## Реализовано в этой итерации

- **FE-00** — каркас через `npm create vite@latest -- --template react-ts` + `vite-plugin-pwa`
  (манифест + service worker — реально закрывает NFR-3, PWA).
- **FE-02** (без кодогенерации) — DTO в `src/api/types.ts` вручную по `api/auth/models.yaml`,
  `api/slots/models.yaml`, `api/bookings/models.yaml`; `authApi.ts`/`slotsApi.ts`/`bookingsApi.ts`
  — тонкие обёртки по `operationId`.
- **FE-03** — `session.ts` поверх `localStorage` (LOGIC-001 — не `sessionStorage`), подписка на
  изменение сессии, централизованная очистка токена на `401` (кроме самого `login`).
- **FE-04** — `AuthScreen` целиком: вкладки Вход/Регистрация, валидация email/пароля по blur,
  спец-случай «409 при регистрации = `code: bad_request`, различать по HTTP-статусу», очистка
  только пароля + фокус на поле при `401` на `login`, disabled/submitting-кнопка.
- **FE-05** — `react-router-dom`, полный маршрутный контракт (см. комментарий в `App.tsx`),
  `TabBar` с реальной навигацией «Классы»/«Мои записи».
- **FE-06** (без BS-001 UI) — `SlotsScreen`: `ScreenState` loading/content/empty/error, бейдж
  «Отменён студией» с приоритетом над «Мест нет», тег сложности, оба варианта Empty, кнопка
  «Обновить» вместо pull-to-refresh.
- **FE-07** — `SlotCardScreen` (`/slots/:slotId`): полная карточка слота (фото, состав, аллергены,
  шеф, вместимость, цена), всегда переисполняет `getSlot`, CTA «Записаться» только при
  `status='scheduled' && free_seats>0`.
- **FE-08** — `BookingScreen` (`/slots/:slotId/book`) + `BookingSuccessSheet` (BS-002): выбор
  экипировки, `createBooking` со сценарием ошибок из `api-sequence.md` (409/410 → снек + возврат
  на список, 400 — текст сервера), одноразовый пуш-промпт (LOGIC-004).
- **FE-09** — `MyBookingsScreen` (`/bookings`): группировка «Предстоящие»/«Прошедшие» строго по
  таблице `SCR-005-my-bookings.md` (не по одному лишь `status`), сортировка внутри групп.
- **FE-10** — `BookingDetailsScreen` (`/bookings/:bookingId`) + `CancelConfirmSheet` (BS-003):
  превью 24ч-правила по серверному `free_cancellation_until`, финальный статус/снек — только из
  ответа `cancelBooking`.
- **FE-11** — `RatingScreen` (`/bookings/:bookingId/rate`): контекст через `navigate(state)` от
  SCR-006 + резервный `getBooking` на прямом заходе, guard-состояния «Недоступно»/«Уже оценено».
- **FE-12** — `AccountSheet` (BS-004): два шага (email → подтверждение выхода), без
  backdrop/swipe-закрытия, `logout` → `session.clear()`.

FE-07..FE-12 реализованы двумя параллельными саб-агентами в изолированных git worktree (аналог
подхода, использованного для бэкендового BE-06..09), результат построчно проверен и слит вручную.

- **FEAT-06/07 (push)** — `handleEnableReminders()` в `BookingSuccessSheet.tsx` теперь реально
  вызывает `PushManager.subscribe()` и регистрирует подписку через `pushApi.register()`
  (`POST /push/subscriptions`); кастомный service worker (`src/sw.ts`, `vite-plugin-pwa` в режиме
  `injectManifest`) показывает системное уведомление на `push` и фокусирует/открывает вкладку на
  `notificationclick`. `VAPID_PUBLIC_KEY` в `core/config.ts` захардкожен и должен совпадать с
  приватным ключом на бэкенде (см. `backend/.env.example`).
- **FEAT-08** — `DateFilterSheet` (BS-001): грабер, свайп/бэкдроп/✕ закрывают без применения,
  «Сбросить»/«Применить» вызывают `resetDateFilter`/`applyDateTo`; чип периода в `SlotsScreen`
  теперь форматирует дату (`toLocaleDateString('ru-RU')`), иконка фильтра показывает точку при
  активном периоде.
- **FEAT-09** — поиск по шефу/программе на `SlotsScreen`: чисто клиентский фильтр уже загруженного
  списка (без нового запроса), «Ничего не найдено» при пустом результате поиска.

FEAT-06..09 реализованы двумя параллельными саб-агентами (push и фильтр+поиск — отдельно, чтобы не
конфликтовать на `SlotsScreen.tsx`) плюс FEAT-02..05 на бэкенде третьим — все три построчно
проверены и слиты вручную (см. `../6-development/FEATURES_IMPLEMENTATION_PLAN.md`).

## Не реализовано

- **FE-13 → FE-15** — вынесенная тема (сейчас CSS-токены захардкожены в `index.css`), тесты
  (Vitest/Testing Library не подключены), финальная сверка готовности — не начаты.

## Проверено вживую

В отличие от Kotlin-версии — здесь реально прогнано после каждого мержа:

- **Важно:** корневой `tsconfig.json` — solution-style (`files: []` + `references`), поэтому
  голый `npx tsc --noEmit` — тихий no-op, ничего не проверяет. Реальную проверку типов делает
  `npx tsc -b --noEmit` (или сам `npm run build`, у которого сборочный шаг — `tsc -b && vite
  build`). Обнаружено агентом при реализации FEAT-06/07 — подтверждено намеренной инъекцией
  ошибки типа.
- `npx tsc -b --noEmit` и `npm run build` — чисто, включая объединённую сборку после мержа трёх
  параллельных саб-агентов (push backend+frontend, фильтр+поиск); 54 модуля трансформировано.
- `npm run build` с `injectManifest`-стратегией `vite-plugin-pwa`: `dist/sw.js` — реально
  собранный кастомный service worker (не auto-generated заглушка), проверено `grep` на
  `push`/`notificationclick`/`showNotification` в собранном файле.
- Полный цикл (регистрация → список → карточка слота → запись → мои бронирования → детали →
  отмена/оценка) ранее уже был опробован пользователем вживую в браузере на реальном бэкенде до
  добавления FE-07..12/FEAT-06..09 — сами новые экраны верифицированы сборкой/типами, но ещё не
  прокликаны руками в браузере после этого мержа (см. риск ниже).

**Риск, требующий ручной проверки:** билд и типы гарантируют компилируемость, но не гарантируют
визуальную/поведенческую корректность (напр. верно ли выглядит бейдж, не съезжает ли вёрстка
шторки, реально ли браузер показывает push-уведомление на реальном устройстве). Прежде чем
считать FE-07..12/FEAT-06..09 полностью готовыми, стоит открыть `npm run dev` в браузере и пройти
полный сценарий записи/отмены и разрешения push вручную.

## Запуск локально

```
npm install
npm run dev      # открыть адрес из вывода в браузере
# или
npm run build && npm run preview
```

## Переменные конфигурации

`src/core/config.ts` → `API_BASE_URL` — захардкожен на `http://localhost:8080` (адрес
`backend/compose.yaml` для локальной разработки). Публичный адрес деплоя не определён — заменить
перед любой сборкой, отличной от локальной.

`src/core/config.ts` → `VAPID_PUBLIC_KEY` — тоже захардкожен (VAPID-ключи не секрет для клиента по
построению), **должен буквально совпадать** с `VAPID_PRIVATE_KEY`/`VAPID_PUBLIC_KEY` на бэкенде
(`backend/.env.example`) — это одна и та же пара ключей, не независимые значения.
