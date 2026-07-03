# Шеф-стол — клиент (Kotlin Compose Multiplatform)

Реализация по `../6-development/FE_IMPLEMENTATION_PLAN.md`, контракту `../api/{auth,slots,common}`
и бэкенду `../backend/`. Единственный подключённый таргет — **wasmJs (веб)**; Android/iOS не
настроены (см. план — веб приоритетнее, остальное отложено).

## Реализовано в этой итерации

- **FE-00** — каркас: `settings.gradle.kts`, `shared/` (commonMain + wasmJsMain), `webApp/`.
- **FE-01** (частично) — `ScreenState`/`ActionStatus` (прямое отражение LOGIC-005), `AppFailure`
  + `ErrorCodes` (список кодов зеркалит `backend/internal/httpapi/errors.go`).
- **FE-02** — DTO вручную по `api/auth/models.yaml` и `api/slots/models.yaml`, `AuthApi`/`SlotsApi`
  поверх Ktor по `operationId`.
- **FE-03** — `SessionStorage`/`BrowserSessionStorage` поверх `localStorage` (не `sessionStorage`,
  см. `LOGIC-001` → «Хранение и использование токена»); `Flow<Boolean>` для реактивного переключения
  корневого экрана; централизованная очистка токена на `401` (кроме самого `login`).
- **FE-04** — экран `SCR-001` целиком: вкладки Вход/Регистрация, валидация email/пароля по blur,
  спец-случай «409 при регистрации = `code: bad_request`, различать по HTTP-статусу», очистка
  только пароля при `401` на `login`, disabled/submitting-состояния кнопки.
- **FE-06** (частично) — `SCR-002`: `ScreenState` Loading/Content/Empty/Error, бейдж «Отменён
  студией» с приоритетом над «Мест нет», кнопка «Сбросить фильтр» в Empty. Стор поддерживает
  `ApplyDateFilter`/`ResetDateFilter` по правилу LOGIC-003 (клиент **никогда** не передаёт
  `date_from`, только опционально `date_to`).

## Не реализовано

- **FE-05** — полноценная корневая навигация с таб-баром: сейчас `App.kt` — простое
  `if (isAuthenticated)` переключение Auth↔Slots, без Decompose и без вкладки «Мои записи»
  (её экрана ещё нет).
- **BS-001** — сама шторка выбора даты (UI) не реализована, только контракт стора
  (`ApplyDateFilter`/`ResetDateFilter`); чип активного периода на SCR-002 тоже не отрисован.
- **FE-07 → FE-15** — карточка слота, оформление записи, подтверждение, мои бронирования, детали
  брони, отмена, оценка шефа, аккаунт/выход, тема, тесты, финальная сверка — не начаты.
- Swipe-to-refresh — заменён обычной кнопкой «Обновить»/«Обновление…» в `SlotsScreen`: не был
  уверен в точном API pull-to-refresh для текущей версии Compose Multiplatform без возможности
  собрать проект и свериться.
- Bearer-токен добавляется вручную в каждом авторизованном запросе (`AuthApi.logout`,
  `SlotsApi.*`) вместо общего Ktor-плагина — приемлемо для 4 эндпоинтов этой итерации, но при
  добавлении `bookings` (FE-08+) стоит вынести в общий плагин/interceptor, чтобы не дублировать.
- **Ничего не собрано и не запускалось.**

## Известные ограничения окружения, в котором писался код

В этой среде есть только **Java 25** (`java -version` отработал), но нет `kotlin`, `gradle` (даже
через `gradlew` не запускалось — целиком неизвестно, установится ли Kotlin/Compose-тулчейн через
сеть в этой песочнице), поэтому весь Kotlin-код в `client/` написан вручную и **ни разу не
компилировался**. Версии в `gradle/libs.versions.toml` (Kotlin 2.0.20, Compose Multiplatform 1.7.0,
Ktor 2.3.12, Koin 3.5.6) подобраны по памяти как совместимый набор на момент написания — сверьте
их актуальность перед первой сборкой. Особые риски для первой сборки:

- Обёртка `Card(onClick = ...)` (Material3) и параметр `supportingText` у `OutlinedTextField` —
  сигнатуры не сверены с реальной версией `compose.material3` из `libs.versions.toml`.
- `io.ktor.client.engine.js.Js` — используется как engine и для JS, и для wasmJs таргетов;
  стоит проверить, что версия `ktor-client-js` в токе поддерживает `wasmJs` (в некоторых версиях
  Ktor это разделено на отдельный артефакт).
- `kotlinx-browser` (`kotlinx.browser.localStorage`) — актуальная координата/версия артефакта не
  проверена live.

Перед тем как доверять этому коду:

```
./gradlew :shared:compileKotlinWasmJs
./gradlew :webApp:wasmJsBrowserDevelopmentRun
```

## Переменные конфигурации

`core/config/AppConfig.API_BASE_URL` — захардкожен на `http://localhost:8080` (адрес
`backend/compose.yaml` для локальной разработки). Публичный адрес деплоя не определён — заменить
перед любой сборкой, отличной от локальной.
