# REST API на Go

Этот репозиторий содержит пример REST API, написанный на Go. API предоставляет функциональность для управления пользователями и заказами.

## Доступ к API

API будет доступно по адресу `http://localhost:8080` (или порт, указанный в `.env`).

Документация Swagger будет доступна по адресу `http://localhost:8080/swagger/index.html`.

## Структура проекта

Код разбит на следующие слои для обеспечения модульности и тестируемости:

*   `handlers/`: Обработчики HTTP-запросов.
*   `services/`: Слой бизнес-логики.
*   `repository/`: Слой доступа к данным (работа с базой данных).
*   `middleware/`: Middleware для обработки запросов (авторизация, логирование).
*   `utils/`: Вспомогательные функции.

## Модели данных

Добавлена модель `Order` и связанные структуры запросов/ответов (`CreateOrderRequest`, `OrderResponse` и т.д.). Все модели содержат аннотации для автоматической генерации документации Swagger.

## Репозиторий

*   Созданы интерфейсы (`UserRepository`, `OrderRepository`) и их реализации для абстрагирования работы с GORM (ORM).
*   Добавлено логирование операций репозитория с использованием `logrus`.
*   Методы для работы с заказами учитывают `UserID` для обеспечения безопасности и предотвращения несанкционированного доступа.

## Сервисы

*   Добавлены `UserService` и `OrderService` для реализации бизнес-логики.
*   `UserService` выполняет такие задачи, как валидация данных пользователей, хеширование паролей и генерация JWT.
*   `OrderService` реализует логику работы с заказами, включая проверку прав доступа.

## Обработчики (Handlers)

*   Обработчики зависят от сервисов, а не напрямую от `gorm.DB`, что обеспечивает слабую связанность и улучшает тестируемость.
*   Реализована обработка ошибок, возвращаемых сервисами, с возвратом соответствующих HTTP-кодов и сообщений.
*   Интегрированы аннотации Swagger (godoc) для автоматической генерации документации.
*   Добавлен `AuthHandler` (внутри `UserHandler`) для обработки логина пользователей.
*   Добавлен `OrderHandler` для работы с заказами.
*   Общие функции (`getPaginationParams`, `getFilteringParams`, `ErrorResponse`) извлечены в файл `common.go` для переиспользования.

## Middleware

*   Добавлен `AuthMiddleware` для проверки JWT токенов в заголовке `Authorization`.
*   Добавлен `LoggerMiddleware` для логирования каждого запроса с использованием `logrus`.

## JWT (JSON Web Tokens)

*   Реализована генерация JWT токенов при логине пользователя (`LoginUser`).
*   Реализована валидация JWT токенов в `AuthMiddleware` с использованием библиотеки `golang-jwt/jwt/v5`.

## Логирование

*   `logrus` используется для структурированного логирования в формате JSON во всех слоях приложения.
*   GORM логгер также настроен на использование `logrus` для консистентного логирования операций базы данных.

## Конфигурация

*   `godotenv` используется для загрузки переменных окружения из файла `.env`.

## Swagger

*   Добавлены аннотации godoc для автоматической генерации документации Swagger.
*   `main.go` настроен для отображения UI Swagger по адресу `/swagger/index.html`.

## Docker

*   Предоставлен `Dockerfile` с многоступенчатой сборкой (multi-stage build) для уменьшения размера конечного образа Docker.
*   Предоставлен `docker-compose.yml` для легкого запуска приложения и базы данных PostgreSQL.
*   Docker Compose использует переменные окружения из файла `.env`.

## Безопасность

*   Пароли пользователей хранятся в базе данных только в виде хешей.
*   Для защиты API используются JWT токены.
*   Реализована базовая авторизация: пользователи могут изменять и удалять только свои собственные данные и заказы.

# Переменные в файле .env
Приложение требует следующие переменные окружения:
# Настройки подключения к базе данных
DB_HOST=localhost
DB_PORT=5432
DB_NAME=users_orders_db
DB_USER=postgres
DB_PASSWORD=131345

# Настройки пула подключений к базе данных
# Максимальное количество подключений в пуле незанятых подключений
DB_MAX_IDLE_CONNS=10
# Максимальное количество открытых подключений к базе данных
DB_MAX_OPEN_CONNS=100
# Максимальное время, в течение которого соединение может быть использовано повторно
DB_CONN_MAX_LIFETIME=1h
# Максимальное количество времени, в течение которого соединение может находиться в режиме ожидания перед закрытием
DB_CONN_MAX_IDLE_TIME=30m

# Общие настройки приложения
PORT=8080
GIN_MODE=release
LOG_LEVEL=info # Уровень логирования (например: panic, fatal, error, warn, info, debug, trace)

# Настройка JWT
JWT_SECRET=eWJyZ3R3cXN5aXZ2b3B0cndxdWVtYWFuc2Nkc2ZzZQ
# JWT_EXPIRATION: Время жизни токена (например: 24h, 3600s)
JWT_EXPIRATION=1h

# Среда приложения (prod или dev)
APP_ENV=prod

# Настройки HTTP сервера (добавлены)
# Таймаут на чтение заголовков и тела запроса (например: 5 секунд)
HTTP_READ_TIMEOUT=5
# Таймаут на запись заголовков и тела ответа (например: 10 секунд)
HTTP_WRITE_TIMEOUT=10
# Таймаут для поддержания Keep-Alive соединений (например: 60 секунд)
HTTP_IDLE_TIMEOUT=60
# Максимальный размер заголовков запроса в байтах (например: 1MB = 1024*1024)
HTTP_MAX_HEADER_BYTES=1048576

# Настройка таймаута для Graceful Shutdown (добавлено)
# Максимальное время ожидания завершения активных запросов при остановке сервера (например: 15 секунд)
SHUTDOWN_TIMEOUT=15s

Полная структура проекта:
project/
├── cmd/
│   ├── main.go          # Точка входа в приложение
|   └── main_test.go
├── docs/
|   ├── docs.go
|   ├── swagger.json
|   └── swagger.yaml
├── internal/
│   ├── core/
|   |   ├── app_core.go
|   |   └── router_core.go
│   ├── handlers/        # Обработчики HTTP-запросов
|   |   ├── common_handler/
|   |   |   ├── common_handler.go
|   |   |   └── common_handler_test.go
|   |   ├── order_handler/
|   |   |   ├── order_handler.go
|   |   |   └── order_handler_test.go
|   |   └── user_handler/
|   |       ├── user_handler.go
|   |       └── user_handler_test.go
│   ├── models/          # Модели базы данных
|   |   ├── user_model/
|   |   |   ├── user_model.go
|   |   |   └── user_model_test.go
|   |   └── order_model/
|   |       ├── order_model.go
|   |       └── order_model_test.go
│   ├── repository/      # Работа с базой данных
|   |   ├── database/
|   |   |   ├── database.go
|   |   |   └── database_test.go
|   |   ├── order_rep/
|   |   |   ├── order_rep.go
|   |   |   └── order_rep_test.go
|   |   └── user_rep/
|   |       ├── user_rep.go
|   |       └── user_rep_test.go
│   ├── services/        # Бизнес-логика
|   |   ├── order_service/
|   |   |   ├── order_service.go
|   |   |   └── order_service_test.go
|   |   └── user_service/
|   |       ├── user_service.go
|   |       └── user_service_test.go
│   ├── middleware/      # Middleware для авторизации
|   |   ├── auth_middleware/
|   |   |   ├── auth_middleware.go
|   |   |   └── auth_middleware_test.go
|   |   └── logger_middleware/
|   |       ├── logger_middleware.go
|   |       └── logger_middleware_test.go
│   └── utils/           # Вспомогательные функции
|       ├── config_util/
|       |   ├── config_util.go
|       |   └── config_util_test.go
|       ├── jwt_util/
|       |   ├── jwt_util.go
|       |   └── jwt_util_test.go
|       ├── logger_util/
|       |   ├── logger_util.go
|       |   └── logger_util_test.go
|       └── password_util/
|           ├── logger_util.go
|           └── logger_util_test.go
├── migrations/          # SQL-миграции
|   ├── 001_users_table.up.sql
|   ├── 001_users_table.down.sql
|   ├── 002_orders_table.up.sql
|   └── 002_orders_table.down.sql
├── .dockerignore
├── .gitignore
├── go.mod               # Файл зависимостей
├── go.sum
├── .env                 # Переменные окружения
├── Dockerfile           # Dockerfile для запуска приложения
├── docker-compose.yml   # Docker Compose для запуска
└── README.md

................
# User Order API

## Обзор Проекта

API для управления пользователями и их заказами. Предоставляет набор конечных точек для выполнения операций, связанных с пользователями и заказами. Построен на Go с использованием фреймворка Gin и поддерживает конфигурирование через переменные окружения.

## Структура Проекта

Проект организован в соответствии с рекомендациями Go Standard Project Layout:

project/
├── cmd/                 # Точки входа для запуска приложений
│   ├── main.go          # Основная точка входа (HTTP сервер)
│   └── main_test.go
├── docs/                # Файлы документации API (Swagger)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/            # Приватный код приложения (не предназначен для использования другими проектами)
│   ├── core/            # Основная инициализация приложения и роутинг
│   │   ├── app_core.go
│   │   └── router_core.go
│   ├── handlers/        # Обработчики HTTP-запросов
│   │   ├── common_handler/
│   │   ├── order_handler/
│   │   └── user_handler/
│   ├── models/          # Структуры данных, представляющие сущности БД
│   │   ├── order_model/
│   │   └── user_model/
│   ├── repository/      # Логика взаимодействия с базой данных
│   │   ├── database/
│   │   ├── order_rep/
│   │   └── user_rep/
│   ├── services/        # Бизнес-логика приложения
│   │   ├── order_service/
│   │   └── user_service/
│   ├── middleware/      # HTTP Middleware
│   │   ├── auth_middleware/
│   │   └── logger_middleware/
│   └── utils/           # Вспомогательные утилиты и хелперы
│       ├── config_util/ # Утилита для загрузки конфигурации
│       ├── jwt_util/    # Утилита для работы с JWT
│       ├── logger_util/ # Утилита для логирования
│       └── password_util/# Утилита для работы с паролями
├── migrations/          # Скрипты миграции базы данных (SQL)
│   ├── 001_users_table.up.sql
│   ├── 001_users_table.down.sql
│   ├── 002_orders_table.up.sql
│   └── 002_orders_table.down.sql
├── .dockerignore        # Исключения для Docker
├── .gitignore           # Исключения для Git
├── go.mod               # Модуль Go и зависимости
├── go.sum               # Контрольные суммы зависимостей
├── .env                 # Переменные окружения для локальной разработки/настройки
├── Dockerfile           # Определение Docker-образа
├── docker-compose.yml   # Определение сервисов для Docker Compose
└── README.md            # Этот файл

## Начало Работы

### Предварительные Требования

*   Go (версия 1.18 или выше)
*   Docker и Docker Compose (рекомендуется для локальной базы данных и запуска в контейнерах)
*   PostgreSQL (если не используете Docker Compose)

### Настройка

1.  **Клонируйте репозиторий:**

        git clone <URL вашего репозитория>
    cd project
    ```

2.  **Настройте переменные окружения:**
    Создайте файл `.env` в корне проекта. Скопируйте содержимое из примера ниже и заполните необходимые данные для подключения к вашей базе данных и других настроек.

    ```dotenv
    # Настройки подключения к базе данных
    DB_HOST=localhost
    DB_PORT=5432
    DB_NAME=users_orders_db
    DB_USER=postgres
    DB_PASSWORD=131345

    # Настройки пула подключений к базе данных
    DB_MAX_IDLE_CONNS=10
    DB_MAX_OPEN_CONNS=100
    DB_CONN_MAX_LIFETIME=1h
    DB_CONN_MAX_IDLE_TIME=30m

    # Общие настройки приложения
    PORT=8080
    GIN_MODE=release # production или debug
    LOG_LEVEL=info # Уровень логирования (panic, fatal, error, warn, info, debug, trace)

    # Настройка JWT
    JWT_SECRET=eWJyZ3R3cXN5aXZ2b3B0cndxdWVtYWFuc2Nkc2ZzZQ # Секретный ключ для подписи JWT
    JWT_EXPIRATION=1h # Время жизни токена (например: 24h, 3600s)

    # Среда приложения (prod или dev)
    APP_ENV=prod

    # Настройки HTTP сервера
    HTTP_READ_TIMEOUT=5 # Таймаут на чтение запроса в секундах
    HTTP_WRITE_TIMEOUT=10 # Таймаут на запись ответа в секундах
    HTTP_IDLE_TIMEOUT=60 # Таймаут для поддержания Keep-Alive соединений в секундах
    HTTP_MAX_HEADER_BYTES=1048576 # Максимальный размер заголовков запроса в байтах (1MB)

    # Настройка таймаута для Graceful Shutdown
    SHUTDOWN_TIMEOUT=15s # Максимальное время ожидания завершения активных запросов при остановке сервера
    ```

3.  **Настройка базы данных:**
    *   **Используя Docker Compose (рекомендуется):**
        Запустите контейнер базы данных и выполните миграции:
        ```bash
        docker-compose up -d db
        # Подождите немного, пока база данных запустится
        # Выполните миграции (необходимо установить инструмент для миграций, например, migrate)
        # migrate -path migrations -database "postgresql://DB_USER:DB_PASSWORD@DB_HOST:DB_PORT/DB_NAME?sslmode=disable" up
        # Замените DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME на значения из вашего .env
        ```
        *Примечание:* Для выполнения миграций может потребоваться установка дополнительного инструмента, например, `migrate` от golang-migrate/migrate. Инструкции по его установке и использованию можно найти в официальной документации инструмента.

    *   **Вручную:**
        Установите и настройте PostgreSQL локально. Создайте базу данных с именем, указанным в `.env`. Выполните SQL-скрипты из директории `migrations/` в правильном порядке (`.up.sql`).

### Запуск Приложения

*   **Используя Go:**
    Убедитесь, что у вас настроен файл `.env` и доступна база данных.
    ```bash
    go run cmd/main.go
    ```

*   **Используя Docker Compose:**
    Убедитесь, что у вас настроен файл `.env` (используя переменные окружения для Docker Compose).
    ```bash
    docker-compose up --build
    ```

## Конфигурация

Приложение использует переменные окружения для конфигурации. Основные переменные описаны в секции [.env](#настройте-переменные-окружения). Утилита `internal/utils/config_util` загружает и парсит эти переменные при запуске.

## Документация API

Документация API доступна в формате Swagger.
При локальном запуске сервера документация доступна по адресу: `http://localhost:<PORT>/swagger/index.html`.
(Замените `<PORT>` на порт, указанный в вашем файле `.env`).

Файлы Swagger (`swagger.json`, `swagger.yaml`) генерируются автоматически с использованием комментариев в коде (`// @...`). Для генерации документации может потребоваться установка и запуск инструмента Swag:
bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/main.go

## Тестирование

В проекте предусмотрены модульные тесты для различных компонентов, расположенные в файлах с суффиксом `_test.go`.
Для запуска всех тестов выполните команду в корне проекта:
bash
go test ./...

Для запуска тестов в конкретной директории:
bash
go test ./internal/services/user_service

## Зависимости

Управление зависимостями осуществляется с помощью Go Modules. Зависимости перечислены в файле `go.mod`.
Для загрузки или обновления зависимостей используйте стандартные команды Go:
bash
go mod tidy # Добавляет недостающие и удаляет неиспользуемые зависимости
go get <пакет> # Добавляет или обновляет конкретный пакет
