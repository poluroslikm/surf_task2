Нужно учитывать важное состояние веток: `feature/client` сейчас не содержит `4aab8d89` и `627e18e5`; merge-base до этих коммитов. Поэтому после merge `main` в `feature/client` появятся новые требования/API, часть из них конфликтует с уже принятым тобой решением убрать push и Яндекс-интеграцию.

**Главные изменения ТЗ/дизайна**
- `4aab8d89`: большой QA rework требований/API: access+refresh token, `club_cancelled`, cancellation reason, pagination/history, дефолт 7 дней, уточнение ошибок, idempotency, a11y, availability, push как MVP.
- `627e18e5`: дизайн-ревью Figma: success после записи стал экраном, профиль редактируется инлайн, OTP segmented, фильтры получили пресеты, карточка слота получила hero/описание/share, карта получила CTA “Проложить маршрут”.
- Но после этого ты отдельно изменил scope: push убрать, Яндекс.Карты заменить моковым скриншотом + external maps handoff.

**Backlog**
**P0 — перед merge/main-sync**
1. Смёржить `main` в `feature/client` и разрешить конфликты по `01-analysis/*`, API и клиенту.
   Traceability: R/RR-D overall.
2. Актуализировать ТЗ под принятое решение: убрать push из MVP и заменить Яндекс-интеграцию на mock screenshot + external maps.
   Сейчас `main` всё ещё содержит LOGIC-007, `registerPushToken`, `reminder_hours`, `is_first_booking`, тексты про системный push и Яндекс. Это уже не соответствует твоему решению.
3. Принять финальный контракт auth: в `main` API уже `TokenPair`/refresh, в `feature/client` код ещё одиночный `token`.
   Нужно либо реализовать access+refresh, либо откатить/упростить требования. По текущему `main` требуется реализация.

**P1 — клиент/API contract**
4. Реализовать access/refresh session flow.
   `VerifyCodeResponse.tokens`, хранение `access_token + refresh_token`, `/auth/refresh`, retry исходного запроса после 401, очистка обоих токенов только если refresh не помог.
5. Обновить auth DTO/repository/session storage.
   Сейчас клиент ждёт `VerifyCodeResponseDto.token`; новый API ожидает `tokens`.
6. Обновить booking DTO/domain под новый API.
   Добавить `BookingStatus.ClubCancelled`, `cancellation_reason`, отображение причины в SCR-005/SCR-006.
7. Проверить createBooking response.
   В `main` specs ожидают `price_total`, `is_first_booking`, `reminder_hours` в контексте push. Если push убираем, `is_first_booking/reminder_hours` нужно удалить из ТЗ/API или оставить как backend-only без UI.
8. Обновить error mapping.
   Проверить `slot_started = 422`, `already_cancelled`, `club_cancelled`, `available_seats/available_rental_boards`, `idempotency_key_conflict`.

**P1 — UI/UX по design review**
9. Переделать success после записи из bottom sheet overlay в полноценный экран.
   Traceability: RR-D03 / BS-002. Кнопки: “Мои бронирования”, “Готово”.
10. Профиль: привести к инлайн-редактированию полей.
   Сейчас клиент всё ещё имеет общий режим edit/save; дизайн требует карандаш у имени/телефона и правку одного поля.
11. OTP: segmented input по длине кода.
   Коммит `b7c4a33 Fix otp length` уже частично закрыл длину; надо сверить, что UI именно segmented и длина не зашита против API.
12. Фильтры: добавить пресеты “Сегодня / Эта неделя / Выходные”.
   Сейчас есть “Сегодня / 7 дней”; не хватает “Эта неделя” и “Выходные”.
13. SCR-003 slot details: hero image/route description.
   Нужно решить источник данных: API `Route.description`/image URL или локальный placeholder. Если не MVP, зафиксировать в ТЗ.
14. Share icon на карточке прогулки.
   По решению заказчика из design-review — Phase 2/decor. В клиенте лучше не добавлять активную функцию; в ТЗ отметить как отложенную.

**P2 — спецификационные чистки**
15. Удалить LOGIC-007 и API push endpoints из актуального MVP, если решение “без push” финальное.
16. Убрать `reminder_hours` из booking success spec/API либо явно оставить неиспользуемым клиентом.
17. BS-004 переименовать тексты: не “Яндекс.Карты”, а “Открыть в картах” / “Проложить маршрут”.
18. Уточнить Runtime config: это не A/B. Оставить как build/host config для `rulesUrl`, `supportUrl`, `appVersion`; убрать упоминания remote feature flags, если где-то появятся.
19. Обновить traceability таблицы после удаления push/Yandex: FR/NFR/LOGIC/CMP должны не ссылаться на исключённые фичи.
20. После merge и правок прогнать:
   `npm --prefix 01-analysis/api run lint`
   `rtk proxy ./gradlew :shared:allTests :androidApp:build :webApp:build --no-daemon --console=plain`

**Короткий вывод**
Самые критичные доработки: синхронизация ветки с `main`, решение конфликта “новая спека требует push/Yandex, актуальный scope их убирает”, затем access/refresh tokens и `club_cancelled/cancellation_reason`. После этого уже UI-слой: success screen, inline profile, фильтры и polish по Figma.
