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
*   `DB_HOST`: Хост базы данных (например, `localhost` или IP-адрес).
*   `DB_PORT`: Порт базы данных (например, `5432`).
*   `DB_USER`: Имя пользователя базы данных.
*   `DB_PASSWORD`: Пароль пользователя базы данных.
*   `DB_NAME`: Имя базы данных.
*   `PORT`: Порт, на котором будет работать приложение (например, `8080`).
*   `JWT_SECRET`: Секретный ключ для JWT (сгенерируйте случайную строку).
*   `GIN_MODE`: Управляет режимом работы веб-фреймворка Gin (debug или release).
*   `LOG_LEVEL`: Управляет уровнем логирования вашего приложения (debug, info, warn, error, fatal).
*   `JWT_EXPIRATION`: Определяет время жизни JSON Web Tokens.

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
|   |   ├── order_db/
|   |   |   ├── order_db.go
|   |   |   └── order_db_test.go
|   |   └── user_db/
|   |       ├── user_db.go
|   |       └── user_db_test.go
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

project/
├── cmd/                  # Точка входа
├── docs/                 # Документация API
├── internal/
│   ├── handlers/         # HTTP-обработчики
│   ├── models/           # Модели данных
│   ├── repository/       # Работа с БД
│   ├── services/         # Бизнес-логика
│   ├── middleware/       # Промежуточное ПО
│   └── utils/            # Вспомогательные утилиты
├── migrations/           # Скрипты миграций БД
├── configs/              # Конфигурации (опционально)
└── deployments/          # Docker/K8s-манифесты (опционально)


Вот полная, развёрнутая структура проекта с комментариями для каждого элемента в стиле Markdown:

project/
├── .dockerignore         Исключаемые при сборке Docker файлы
├── .env                  Переменные окружения (DB, JWT, порты)
├── .env.example          Шаблон для .env
├── .gitignore            Игнорируемые Git файлы
├── docker-compose.yml    Конфигурация сервисов (app + postgres)
├── Dockerfile            Сборка Go-приложения
├── go.mod                Модули Go
├── go.sum                Хеши зависимостей
└── README.md             Инструкции по запуску

cmd/
├── main.go               Инициализация:
└── main_test.go          Интеграционные тесты запуска

docs/
├── docs.go               Генерация Swagger
├── swagger.json          OpenAPI (JSON)
└── swagger.yaml          OpenAPI (YAML)

internal/handlers/
├── common_handler/       Общие эндпоинты
│   ├── common_handler.go     Healthcheck, метрики, версия API
│   └── commonhandlertest.go
├── order_handler/        Управление заказами
│   ├── order_handler.go      CRUD заказов + валидация
│   └── orderhandlertest.go
└── user_handler/         Аутентификация
    ├── user_handler.go       Регистрация/логин/обновление
    └── userhandlertest.go

internal/models/
├── order_model/          Сущность "Заказ"
│   ├── order_model.go        Поля:
│   └── ordermodeltest.go   - ID, статус, стоимость, пользователь
└── user_model/           Сущность "Пользователь"
    ├── user_model.go         Поля:
    └── usermodeltest.go    - ID, email, хеш пароля, роль

internal/repository/
├── database/             Подключение к БД
│   ├── database.go           Инициализация GORM
│   └── database_test.go
├── order_db/             Операции с заказами
│   ├── order_db.go           Get/Create/Update
│   └── orderdbtest.go
└── user_db/              Операции с пользователями
    ├── user_db.go            FindByEmail, Create
    └── userdbtest.go

internal/services/
├── order_service/        Логика заказов
│   ├── order_service.go      Проверка прав доступа
│   └── orderservicetest.go
└── user_service/         Логика пользователей
    ├── user_service.go       Генерация JWT, валидация
    └── userservicetest.go

internal/middleware/
├── auth_middleware/      Аутентификация
│   ├── auth_middleware.go    Проверка JWT
│   └── authmiddlewaretest.go
└── logger_middleware/    Логирование
    ├── logger_middleware.go  Формат логов (HTTP-запросы)
    └── loggermiddlewaretest.go

internal/utils/
├── config_util/          Загрузка конфигов
│   ├── config_util.go        Чтение .env
│   └── configutiltest.go
├── jwt_util/             Работа с JWT
│   ├── jwt_util.go           Генерация/парсинг токенов
│   └── jwtutiltest.go
├── logger_util/          Логгирование
│   ├── logger_util.go        Инициализация Logrus
│   └── loggerutiltest.go
└── password_util/        Безопасность
    ├── password_util.go      Хеширование (bcrypt)
    └── passwordutiltest.go

migrations/
├── 001userstable.up.sql    Создание users
├── 001userstable.down.sql  Удаление users
├── 002orderstable.up.sql   Создание orders
└── 002orderstable.down.sql Удаление orders

configs/                  Конфиги в YAML/JSON
deployments/              Kubernetes-манифесты
scripts/                  Скрипты для CI/CD

Эта структура соответствует лучшим практикам Go и подходит для REST API среднего размера. Для более сложных проектов можно добавить pkg/ для общего кода или api/ для gRPC.
