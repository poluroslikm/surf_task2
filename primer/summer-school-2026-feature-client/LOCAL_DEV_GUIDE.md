# Локальный запуск BE, DB и клиента

Краткий guide для разработки Volna: база и backend запускаются из `backend/`, клиент - из `client/`.

## 1. Требования

- Docker Compose.
- Go 1.23+.
- JDK 17+.
- Android Studio или IntelliJ IDEA для Android/Web.
- Xcode для iOS.
- Node.js/npm нужны только для OpenAPI-команд из `01-analysis/api`.

## 2. DB: PostgreSQL

1. Перейдите в backend:

```bash
cd backend
```

2. Поднимите PostgreSQL:

```bash
docker compose --profile db up -d db
```

3. Примените миграции и dev-seed:

```bash
make migrate
```

4. Если нужны готовые состояния для проверки клиентских экранов, примените mock-seed:

```bash
docker compose -f compose.yaml exec -T db psql -U volna -d volna < seed/mock_client_app_states.sql
```

Подключение по умолчанию: `postgres://volna:volna@localhost:5432/volna?sslmode=disable`.

## 3. BE: запуск API

1. Из `backend/` запустите API локально:

```bash
make run
```

2. Проверьте, что сервис жив:

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/readyz
```

3. Альтернатива: поднять API вместе с DB через Compose:

```bash
docker compose --profile app up --build
```

Основные переменные можно положить в `backend/.env`, взяв шаблон из `backend/.env.example`. Важные значения: `HTTP_ADDR`, `DATABASE_URL`, `TEST_DATABASE_URL`, `BASE_URL`.

## 4. BE: полезные команды

```bash
make test          # Go-тесты
make lint          # go vet
make generate      # генерация OpenAPI transport-кода
make lint-api      # lint OpenAPI-контрактов
```

Для k6:

```bash
make k6-seed
BASE_URL=http://127.0.0.1:8080 make k6-smoke
```

## 5. Клиент: запуск и проверка

1. Перейдите в клиент:

```bash
cd client
```

2. Быстрая проверка shared-кода:

```bash
./gradlew :shared:compileDebugKotlinAndroid
```

3. Android debug build:

```bash
./gradlew :androidApp:assembleDebug
```

4. Web dev server:

```bash
./gradlew :webApp:wasmJsBrowserDevelopmentRun
```

5. iOS: откройте `client/iosApp/iosApp.xcodeproj` в Xcode и запустите схему `iosApp`. Shared framework собирается Gradle-задачей Xcode build phase.

## 6. Клиент и локальный BE

- Shared-клиент по умолчанию ходит на `http://localhost:8080`.
- Для Web и iOS simulator локальный backend доступен как `localhost:8080`.
- Для Android emulator используйте `10.0.2.2:8080`, если запросы из эмулятора не доходят до host-машины.
- В Android уже есть debug cleartext-настройка для `localhost`, `127.0.0.1` и `10.0.2.2`.

## 7. Рекомендуемый порядок ежедневного запуска

```bash
cd backend
docker compose --profile db up -d db
make migrate
make run
```

В соседнем терминале:

```bash
cd client
./gradlew :webApp:wasmJsBrowserDevelopmentRun
```

Для Android/iOS удобнее запускать host-приложения из IDE после поднятия DB и BE.
