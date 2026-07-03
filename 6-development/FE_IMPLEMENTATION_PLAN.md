# План реализации FE (CMP-клиента) для «Шеф-стол»

> Составлен по аналогии с референсным
> `primer/summer-school-2026-feature-client/02-development/CMP_CLIENT_IMPLEMENTATION_PLAN.md`,
> но пересчитан под наш домен, наш API (`api/auth`, `api/slots`, `api/bookings`, `api/common`) и
> уже реализованный бэкенд (`backend/`, см. `6-development/BE_IMPLEMENTATION_PLAN.md`). Референс
> проектировал другой продукт (сапбординг-клуб «Волна»: телефон+OTP, группы 1–3 места, прокат досок
> с лимитом, карта маршрута, полноценный профиль) — здесь те же архитектурные приёмы (MVI, чистая
> архитектура, Decompose), но состав экранов, доменные правила и часть логик другие. Расхождения
> отмечены явно по ходу документа.

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

- [ ] [FE-00. Создать каркас KMP/CMP-проекта](#fe-00-создать-каркас-kmpcmp-проекта)
- [ ] [FE-01. Ядро: MVI-база, DI, конфиг, ошибки, логирование](#fe-01-ядро-mvi-база-di-конфиг-ошибки-логирование)
- [ ] [FE-02. Сетевой слой и DTO по контракту](#fe-02-сетевой-слой-и-dto-по-контракту)
- [ ] [FE-03. Хранилище токена и сессии](#fe-03-хранилище-токена-и-сессии)
- [ ] [FE-04. Экран авторизации SCR-001 (LOGIC-001)](#fe-04-экран-авторизации-scr-001-logic-001)
- [ ] [FE-05. Корневая навигация и таб-бар](#fe-05-корневая-навигация-и-таб-бар)
- [ ] [FE-06. Список слотов SCR-002 + фильтр дат BS-001 (LOGIC-003, LOGIC-005)](#fe-06-список-слотов-scr-002--фильтр-дат-bs-001-logic-003-logic-005)
- [ ] [FE-07. Карточка слота SCR-003](#fe-07-карточка-слота-scr-003)
- [ ] [FE-08. Оформление записи SCR-004 + подтверждение BS-002 (LOGIC-004)](#fe-08-оформление-записи-scr-004--подтверждение-bs-002-logic-004)
- [ ] [FE-09. Мои бронирования SCR-005](#fe-09-мои-бронирования-scr-005)
- [ ] [FE-10. Детали брони SCR-006 + отмена BS-003 (LOGIC-002)](#fe-10-детали-брони-scr-006--отмена-bs-003-logic-002)
- [ ] [FE-11. Оценка шефа SCR-007](#fe-11-оценка-шефа-scr-007)
- [ ] [FE-12. Аккаунт / выход BS-004](#fe-12-аккаунт--выход-bs-004)
- [ ] [FE-13. Тема оформления](#fe-13-тема-оформления)
- [ ] [FE-14. Тесты](#fe-14-тесты)
- [ ] [FE-15. Финальная проверка готовности FE](#fe-15-финальная-проверка-готовности-fe)

## Стек приложения

- Kotlin Multiplatform + Compose Multiplatform. Целевые платформы: **веб (Wasm/JS)** — основная,
  т.к. по ТЗ продукт это PWA (NFR-3), без публикации в сторах; Android/iOS — опционально, если
  решение расширить платформы будет принято отдельно (см. предыдущее обсуждение архитектуры —
  выбор CMP/MVI на Kotlin уже сделан вместо изначального PWA-на-вебе решения ТЗ, но публикация в
  сторах по-прежнему вне скоупа, NFR-3).
- Coroutines, Flow, kotlinx.serialization, kotlinx.datetime.
- Ktor Client — REST-клиент.
- Koin — DI.
- Decompose — навигация и жизненный цикл компонентов, MVI-хранилища через простые
  `StateFlow`-редьюсеры (без отдельной MVIKotlin-зависимости — для нашего объёма экранов, 7 против
  референсных с картой/профилем, хватает более лёгкого решения).
- Multiplatform Settings — хранение токена (см. `LOGIC-001` → «Хранение и использование токена» —
  `localStorage`, не `sessionStorage`).
- Kermit — логирование.
- **Нет:** MapLauncher/RouteMapPreview, moko-permissions для карт — в проекте нет карты/маршрута.
  Push-разрешение — только браузерный Web Push API (`Notification`/`PushManager`), не нативный
  SDK (NFR-3, `LOGIC-004`).

## Чистая архитектура

Как в референсе — зависимости внутрь:

```text
presentation -> domain <- data
data -> domain interfaces
platform adapters (storage, push) -> shared interfaces
```

Пакеты:

```text
com.chefstol.app.core
  config
  error
  network
  time
  storage
  mvi
  ui
  theme
  navigation
com.chefstol.app.auth
com.chefstol.app.catalog   (Program/Chef/Slot — read-only)
com.chefstol.app.booking   (Booking, rating)
```

Нет пакетов `profile`, `map`, `notifications` как отдельных доменов (см. таблицу отличий выше) —
push инкапсулирован в `LOGIC-004`-адаптер внутри `booking` (запрашивается на BS-002).

## Domain model

```kotlin
data class Client(val id: ClientId, val email: String, val createdAt: Instant)

enum class ProgramDifficulty { NOVICE, EXPERIENCED }

data class Program(
    val id: ProgramId,
    val name: String,
    val difficulty: ProgramDifficulty,
    val photoUrl: String,
    val ingredients: List<String>,
    val allergens: List<String>,
)

data class Chef(val id: ChefId, val name: String)

enum class SlotStatus { SCHEDULED, CANCELLED }

data class Slot(
    val id: SlotId,
    val program: Program,
    val chef: Chef,
    val startAt: Instant,
    val totalSeats: Int,
    val freeSeats: Int,
    val price: Int,
    val status: SlotStatus,
    val cancellationReason: String?,
)

enum class EquipmentChoice { OWN, RENTAL }
enum class BookingStatus { ACTIVE, CANCELLED, LATE_CANCEL, CANCELLED_BY_STUDIO }

data class Rating(val stars: Int, val comment: String?, val createdAt: Instant)

data class Booking(
    val id: BookingId,
    val slotId: SlotId,
    val equipmentChoice: EquipmentChoice,
    val status: BookingStatus,
    // Всегда из API — клиент не пересчитывает (FR-23, P3).
    val priceTotal: Int,
    val freeCancellationUntil: Instant,
    val rating: Rating?,
    val createdAt: Instant,
    val cancelledAt: Instant?,
    val slot: Slot,
)
```

Pure-функции (по числу наших логик — **не** переносить референсные `AvailabilityPolicy`/
`BookingPriceCalculator`, они относятся к домену, которого у нас нет):

- `CancellationPreview` (`LOGIC-002`): по `booking.freeCancellationUntil` и текущему времени —
  только **превью** для UI до подтверждения на BS-003; финальное решение (`cancelled` vs
  `late_cancel`) всегда принимает сервер, клиент это значение не хранит как источник истины.
- `SlotFilterQuery` (`LOGIC-003`): построение `date_from`/`date_to` для `listSlots` по правилам
  из LOGIC-003 (дефолт — без параметров; фильтр — только `date_to`; сброс — снова без параметров).
- `BookingGrouping`: «Предстоящие»/«Прошедшие» на SCR-005 — производится из `slot.startAt`
  относительно текущего момента, не хранимый статус (см. `bookings/api.yaml` → `listBookings`).

## MVI-стандарт

Единый паттерн состояний экрана — прямое отражение `LOGIC-005`, а не абстрактный общий стандарт
"на всякий случай":

```kotlin
sealed interface ScreenState<out T> {
    data object Loading : ScreenState<Nothing>
    data class Content<T>(val value: T, val refreshing: Boolean = false) : ScreenState<T>
    data class Empty(val reason: String) : ScreenState<Nothing>
    data class Error(val message: String) : ScreenState<Nothing>
}

enum class ActionStatus { Idle, Submitting }
```

Правила (буквально из `LOGIC-005`, не расширять сверх документированного):

- Первичная загрузка: `Loading -> Content|Empty|Error`.
- `Refreshing` (pull-to-refresh): контент/empty сохраняется, `refreshing=true`; провал — **только**
  снек, `Error` не показывается (LOGIC-005 AC-007).
- Действие (`createBooking`/`cancelBooking`/`submitRating`) не имеет своего Loading/Error —
  только `ActionStatus.Submitting` и снек по результату (LOGIC-005 «Обратная связь по действию»).
- `401` на любом запросе, кроме `login`, — централизованно очистить токен и вернуть на SCR-001
  (`LOGIC-001` → «Хранение и использование токена», AC-007). SCR-001 сам не проходит через
  `ScreenState` — он не выполняет GET при открытии (см. FE-04).

## Навигация (карта из `feature-list.md` §3)

```text
Root
  Auth (SCR-001)               — НЗ, корень при отсутствии токена
  MainTabs                      — АЗ
    Slots (SCR-002 -> SCR-003 -> SCR-004)
    MyBookings (SCR-005 -> SCR-006 -> SCR-007)
```

Bottom sheets: `BS-001` (над SCR-002), `BS-002` (после `createBooking`), `BS-003` (над SCR-006),
`BS-004` (над SCR-002 и SCR-005, через иконку аккаунта). **Нет** `BS-004 route-map` референса — в
нашем проекте `BS-004` это «Аккаунт / выход», совсем другая шторка с тем же ID из-за независимой
нумерации двух проектов.

## Data layer

Репозитории — по нашим 3 доменам API, без `ProfileRepository`/`InstructorRepository`:

- `AuthRepository`: `register`, `login`, `logout`.
- `SlotsRepository`: `listSlots`, `getSlot`.
- `BookingsRepository`: `createBooking`, `listBookings`, `getBooking`, `cancelBooking`, `submitRating`.
- `SessionRepository`: чтение/запись/очистка токена, `Flow<Boolean>` состояния авторизации.

Ktor setup — как в референсе: `ContentNegotiation` (snake_case через `@SerialName` на каждом
поле, а не глобальный naming strategy — контракт использует единообразный snake_case, но
не полагаемся на "угадывание" стратегией), `HttpTimeout`, базовый URL из конфигурации сборки,
плагин, добавляющий `Authorization: Bearer <token>` из `SessionRepository`.

Типизированные ошибки — по факту нашего `common/models.yaml → Error.code` enum (значения см. в
`backend/internal/httpapi/errors.go`, воспроизводить тот же список, не изобретать новые):

```kotlin
sealed interface AppFailure
data object Unauthorized : AppFailure
data class ApiFailure(val code: String, val message: String) : AppFailure // code — из Error.code enum
data object NetworkUnavailable : AppFailure
data object UnknownFailure : AppFailure
```

**Idempotency-Key — не реализуется в FE в этой итерации**: контракт `createBooking` его не
принимает (в отличие от референсного BE). Реализация клиента не должна ничего генерировать и
отправлять сверх задокументированного `CreateBookingRequest` (`slot_id`, `equipment_choice`) —
если дубли при retry станут проблемой, это сначала правка `api/bookings/api.yaml`
(см. `BE_IMPLEMENTATION_PLAN.md` → BE-06), а не решение только на клиенте.

---

## Декомпозиция FE

### FE-00. Создать каркас KMP/CMP-проекта

Сделать:
- `client/` с Gradle KMP-настройкой: `settings.gradle.kts`, корневой `build.gradle.kts`,
  `gradle/libs.versions.toml`, модуль `shared/` (`commonMain`, `commonTest`, `wasmJsMain` —
  веб-таргет в приоритете; `androidMain`/`iosMain` опциональны, добавляются отдельным пунктом,
  если платформы будут явно расширены).
- `webApp/` — точка входа Compose for Web/Wasm, хостит `shared`.
- Пакет `com.chefstol.app`, структура папок из раздела «Чистая архитектура» выше.

Готово, когда:
- `./gradlew :shared:compileKotlinWasmJs` (или соответствующая задача) проходит.
- Пустой экран открывается в браузере (`./gradlew :webApp:wasmJsBrowserRun`).

### FE-01. Ядро: MVI-база, DI, конфиг, ошибки, логирование

Сделать:
- `ScreenState<T>`/`ActionStatus` (см. «MVI-стандарт» выше) в `core.mvi`.
- Koin-модули: сеть, репозитории, сторы.
- `core.config`: базовый URL API из build-конфигурации (dev/prod).
- `core.error`: маппинг HTTP-ответов в `AppFailure`.
- Kermit-логирование по слоям (не логировать `password`/`token` — NFR-7).

Готово, когда:
- Юнит-тест на маппинг `Error{code,message}` → `AppFailure` проходит для всех кодов из
  `common/models.yaml` (см. FE-14).

### FE-02. Сетевой слой и DTO по контракту

Сделать:
- DTO в `commonMain` **вручную по образцу** `api/*/models.yaml` (аналогично `backend/internal/httpapi/dto` —
  генератора клиента из OpenAPI в этой итерации не подключаем, см. «Не реализовано» в отчёте).
- `AuthApi`, `SlotsApi`, `BookingsApi` — тонкие обёртки над Ktor по `operationId`
  (`register`, `login`, `logout`, `listSlots`, `getSlot`, `createBooking`, `listBookings`,
  `getBooking`, `cancelBooking`, `submitRating`).
- Единый `HttpClient` с auth-плагином и маппером ошибок из FE-01.

Готово, когда:
- Ktor `MockEngine`-тест на каждый `operationId`: успех + хотя бы один документированный код
  ошибки (см. FE-14).

### FE-03. Хранилище токена и сессии

Сделать:
- `SessionRepository` поверх Multiplatform Settings: `localStorage` в вебе (см. `LOGIC-001` →
  «Хранение и использование токена» — явно **не** `sessionStorage`).
- `Flow<Boolean>` — есть ли валидный токен (используется корневой навигацией, FE-05).
- Централизованная реакция на `401` (кроме ответа `login`): очистить токен, эмитировать событие
  возврата на SCR-001 (`LOGIC-001` AC-007).

Готово, когда:
- Тест: сохранённый токен переживает пересоздание `SessionRepository` (симуляция перезапуска).
- Тест: `401` на любом запросе, отличном от `login`, очищает токен.

### FE-04. Экран авторизации SCR-001 (LOGIC-001)

Сделать (по `5-mobile-app-spec/SCR-001-auth.md` буквально, включая нюансы):
- Форма с двумя вкладками (Вход/Регистрация), поля `email`/`password`, переключатель видимости
  пароля. **Вкладка «Вход» активна по умолчанию.**
- Клиентская валидация: email — формат (E1) по blur и при сабмите; пароль — `minLength: 8`
  **только на вкладке «Регистрация»** (E2); на вкладке «Вход» длина не проверяется клиентом —
  решает ответ сервера.
- CTA disabled, пока одно из полей пусто; в `submitting` — поля заблокированы, повторный тап
  невозможен (AC-N07).
- Обработка ответов **буквально по SCR-001/LOGIC-001**, включая специфичный нюанс: email уже
  занят — HTTP `409`, но `Error.code = bad_request` (не отдельный доменный код) — различать
  **по HTTP-статусу**, не по `code` (AC-N04, AC-003 LOGIC-001).
- `401` от `login` — сообщение «Неверный email или пароль», очистить `password`, сохранить
  `email`, фокус на пароль (AC-N03). Это внутренняя обработка формы, не общий `401`-обработчик
  из FE-03 (тот применяется ко **всем прочим** запросам, не к самому `login`).
- При успехе (`201`/`200`) — сохранить `token`+`client` через `SessionRepository`, сразу перейти
  на SCR-002, без промежуточного экрана успеха.
- Экран **не** использует `ScreenState`/`LOGIC-005` — нет GET-запроса при открытии (см. SCR-001
  «Состояния экрана»); свои локальные состояния: `Content(пустая форма)`/`Ввод`/`Submitting`/`Ошибка`.

Готово, когда:
- Все AC из `SCR-001-auth.md` (AC-001…AC-003, AC-N01…AC-N07, AC-E01…AC-E03) покрыты тестами
  стора/компонента.

### FE-05. Корневая навигация и таб-бар

Сделать:
- Decompose root-компонент: `Auth` (если нет токена) ↔ `MainTabs` (если есть), переключение по
  `Flow<Boolean>` из FE-03.
- Таб-бар авторизованной зоны: «Классы» (SCR-002) / «Мои записи» (SCR-005), виден во всех
  состояниях экрана (Loading/Content/Empty/Error) — см. SCR-002 §5.
- Иконка аккаунта в хедере SCR-002 и SCR-005 → `BS-004`.

Готово, когда:
- Навигационный тест: `logout` из `BS-004` возвращает на SCR-001 и разрушает `MainTabs`-стек
  (повторный переход вперёд/назад не показывает данные предыдущей сессии).

### FE-06. Список слотов SCR-002 + фильтр дат BS-001 (LOGIC-003, LOGIC-005)

Сделать (по `5-mobile-app-spec/SCR-002-slot-list.md` и `09_Логики/LOGIC-003`):
- Стор с `ScreenState<List<SlotUi>>`, дефолтный запрос `listSlots` **без** `date_from`/`date_to`
  при первом открытии.
- Чип активного периода — показывается, только если `dateFilter` задан и отличается от дефолта;
  ✕ на чипе → сброс без параметров дат.
- `BS-001`: поле «Показать по» — единственный ввод; «Применить» → `date_to=<дата>` **без**
  `date_from`; «Сбросить» → снова без параметров (LOGIC-003, не референсный `AND`/`OR` набор
  фильтров по типу/инструктору — их у нас просто нет).
- Карточка слота: бейдж «Мест нет» (`status=scheduled, free_seats=0`) **или** «Отменён студией»
  (`status=cancelled`, приоритет над «Мест нет» — если отменён, число мест не показывается вовсе,
  независимо от значения `free_seats`, SCR-002 §3). Оба бейджа кликабельны, ведут на SCR-003 без
  CTA записи на самой карточке.
- Loading — скелетон 4–6 карточек, не пустой экран; Empty — «Пока нет доступных классов» + кнопка
  «Выбрать другой период»/«Сбросить фильтр» в зависимости от того, активен ли дефолт; pull-to-refresh
  ошибка — снек, список не заменяется на `Error`.

Готово, когда:
- Все AC из `SCR-002-slot-list.md` покрыты (включая граничные AC-E01…AC-E04 — граница периода
  включительна, приоритет бейджа отмены над «Мест нет» независимо от `free_seats`).

### FE-07. Карточка слота SCR-003

Сделать:
- Полные данные слота (`getSlot`, перезапрос при открытии, не переиспользуется объект из списка —
  чтобы не показать устаревшую вместимость перед записью, как явно указано в референсном стеке
  решений «reload with getSlot to avoid stale availability»; тот же принцип применим и здесь).
- CTA «Записаться» — доступен только для `status=scheduled` и `free_seats>0`; для «Мест нет» и
  «Отменён студией» — не отображается / недоступен (FR-6a, FR-18).

Готово, когда:
- Переход на SCR-004 только из доступного состояния; для двух прочих состояний CTA не рендерится.

### FE-08. Оформление записи SCR-004 + подтверждение BS-002 (LOGIC-004)

Сделать:
- Форма выбора `equipment_choice` (`own`/`rental`), цена — **всегда** `slot.price`, показывается
  как есть, без вычислений (нет группы мест, нет проката с лимитом, в отличие от референса).
- `createBooking`: `slot_full` (409, из `details.available_seats` — обновить отображение), `slot_cancelled`
  (410), сетевые/5xx — снек, без перехода в `Error` экрана (LOGIC-005 «Обратная связь по действию»).
- При успехе → `BS-002`: сводка + однократный (только при первой в жизни клиента успешной записи
  на этом устройстве) запрос push-разрешения через `Notification.requestPermission()`
  (`LOGIC-004`, Web Push API — **не** нативный SDK, NFR-3). Признак `push_offered` — в
  локальном хранилище устройства, не на сервере (регистрация подписки — открытый пробел контракта,
  см. `LOGIC-004` → «API-запросы», не реализуется до отдельного решения).

Готово, когда:
- AC-001…AC-008 из `LOGIC-004` (показ блока push ровно один раз, синхронизация с
  `Notification.permission`, тихая деградация при отсутствии Push API) покрыты тестами.

### FE-09. Мои бронирования SCR-005

Сделать:
- `listBookings` — без пагинации (контракт её не имеет), группировка «Предстоящие»/«Прошедшие»
  **на клиенте** по `slot.startAt` относительно текущего момента — не по `status`.

Готово, когда:
- Тест группировки: `active`-бронь с `slot.startAt` в прошлом попадает в «Прошедшие», не в
  отдельную категорию `past` (такого статуса не существует).

### FE-10. Детали брони SCR-006 + отмена BS-003 (LOGIC-002)

Сделать:
- CTA «Отменить» доступен, только пока `slot.startAt` в будущем.
- Превью на BS-003 — по `booking.freeCancellationUntil` (уже посчитан сервером): **≥ 24ч до
  старта** → превью «бесплатно», **< 24ч** → превью предупреждения о закупленных продуктах, без
  штрафа (не 2 часа, как в референсе!). Финальный `status` (`cancelled`/`late_cancel`) — из ответа
  `cancelBooking`, клиент не решает сам.
- CTA «Оценить шефа» — виден на SCR-006, только если слот прошёл, `status=active`, `rating=null`
  → переход на SCR-007.

Готово, когда:
- Граничный тест на `freeCancellationUntil` ровно в текущий момент и по обе стороны от него.

### FE-11. Оценка шефа SCR-007

Сделать (сущности, которой нет в референсе целиком):
- Форма 1–5 звёзд + опциональный комментарий, `submitRating`.
- `422 not_ratable` / `409 already_rated` — снек, без перехода формы в `Error`.
- Редактирование ранее поставленной оценки не предусмотрено (UC-4 E2) — если `rating != null`,
  CTA на SCR-006 не показывается вовсе (см. FE-10).

Готово, когда:
- Тест: повторный вызов `submitRating` для уже оценённой записи не выполняется (CTA скрыт), а
  прямой ответ `already_rated` (если форма всё же была открыта до обновления состояния) даёт снек,
  не краш.

### FE-12. Аккаунт / выход BS-004

Сделать:
- Минимальная шторка: `logout` → очистка сессии (FE-03) → переход на SCR-001. Никакого профиля,
  смены email, удаления аккаунта — этого нет в ТЗ (см. таблицу отличий выше).

Готово, когда:
- Тест: после `logout` попытка открыть `MainTabs` напрямую (например, back-навигацией) возвращает
  на SCR-001, а не показывает закэшированные данные.

### FE-13. Тема оформления

Сделать:
- Токены темы (цвета/типографика/отступы) — **не зафиксированы Figma-макетом** (бренд не задан
  заказчиком, см. `3-design-brief/00-foundations.md` → «Объём визуальных требований»). В отличие
  от референса (там есть Figma-файл с открытым вопросом извлечения токенов), здесь с самого начала
  нет визуального источника — реализовать нейтральную тему по ASCII-wireframe из дизайн-брифов,
  зафиксировать токены как собственное решение, не как заимствование из внешнего макета.
- Все состояния из `LOGIC-005` (skeleton/empty/error/snackbar) — токенизированы, без хардкода
  цветов в экранных компонентах.

Готово, когда:
- Ни один файл экрана не содержит цветовых/типографических литералов вне `core.theme`.

### FE-14. Тесты

Сделать:
- `commonTest`: границы `LOGIC-002` (24ч ровно, ±1с, после старта), `LOGIC-003` (граница периода
  включительна), MVI-редьюсеры каждого экрана, маппинг `Error.code` → `AppFailure`.
- Ktor `MockEngine`: успех/ошибка на каждый `operationId` из `api/`.
- Компонентные тесты состояний `Loading/Content/Empty/Error/Refreshing` (LOGIC-005 AC-001…AC-010).

Готово, когда:
- `./gradlew :shared:allTests` проходит.

### FE-15. Финальная проверка готовности FE

Сделать:
- Сверить реализованные экраны с `5-mobile-app-spec/README.md` (таблица экранов) — ничего не
  реализовано сверх скоупа (нет карты/маршрута, нет профиля, нет группового бронирования).
- Прогнать `./gradlew :shared:allTests`, сборку `webApp`.
- Ручная сверка критериев приёмки каждого SCR/BS-документа со сценариями в приложении.

Готово, когда:
- Все Must-экраны (SCR-001…SCR-006, BS-001…BS-004) реализованы и покрыты тестами; SCR-007
  (Should) реализован в рамках FE-11.
