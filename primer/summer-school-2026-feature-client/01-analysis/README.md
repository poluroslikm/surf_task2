# День 1 — Анализ · «Сапбординг-клуб»

Артефакты аналитика по сквозному проекту летней школы. Структура повторяет
классический процесс работы аналитика: от входа заказчика до ТЗ, которое
передаётся в **День 2** разработчику.

## Маршрут по этапам

| Этап | Папка | Что внутри |
| :-- | :-- | :-- |
| **Вход** | [0-customer-brief/](0-customer-brief/) | [customer-brief.md](0-customer-brief/customer-brief.md) — сырой бриф заказчика (заполнен) |
| **1. Выявление требований** | [1-elicitation/](1-elicitation/) | [customer-questions.md](1-elicitation/customer-questions.md), [domain-description.md](1-elicitation/domain-description.md) |
| **2. Описание требований** | [2-requirements/](2-requirements/) | [business](2-requirements/business-requirements.md) · [functional](2-requirements/functional-requirements.md) · [non-functional](2-requirements/non-functional-requirements.md) · [user-stories](2-requirements/user-stories.md) · [use-cases](2-requirements/use-cases.md) |
| **Бриф для дизайна** | [3-design-brief/](3-design-brief/) | [design-brief.md](3-design-brief/design-brief.md) — требования для UI/UX дизайнера |
| **3. Проектирование** | [4-design/](4-design/) | [data-model.md](4-design/data-model.md), [api-sequence.md](4-design/api-sequence.md) |
| **4. ТЗ** | [5-mobile-app-spec/](5-mobile-app-spec/) | [README.md](5-mobile-app-spec/README.md) — шаблон продумаем совместно |
| **API (OpenAPI)** | [api/](api/) | [redocly.yaml](api/redocly.yaml) — многофайловый OpenAPI (домены: auth, slots, bookings, profile, instructors) |

## Дополнительно (подготовка к лекции)

- [prompts/](prompts/) — [хорошие](prompts/good-prompts.md) и [плохие](prompts/bad-prompts.md) промпты для демо.
- [checklists/](checklists/) — [чек-лист цифровой гигиены](checklists/digital-hygiene-checklist.md) перед передачей в разработку.

## Статус

Бриф заказчика заполнен. Остальные файлы — **пустые шаблоны**, заполняются по ходу
подготовки и на лекции. Шаблон ТЗ (`5-mobile-app-spec/`) ещё предстоит выстроить совместно.

**API:** спецификация переведена на многофайловый формат OpenAPI (Redocly) —
реестр доменов в [api/redocly.yaml](api/redocly.yaml); устаревший единый
`api/openapi.yaml` больше не используется. Контракты доработаны по QA-ревью.

> **Передача в День 2:** итоговые требования + модель данных + API-спецификация + ТЗ.
