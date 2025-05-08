package main

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/IlyushinDM/user-order-api/internal/handlers/order_handler"
	"github.com/IlyushinDM/user-order-api/internal/handlers/user_handler"
	auth_mw "github.com/IlyushinDM/user-order-api/internal/middleware/auth_middleware"
	log_mw "github.com/IlyushinDM/user-order-api/internal/middleware/logger_middleware"
	"github.com/IlyushinDM/user-order-api/internal/repository/database"
	"github.com/IlyushinDM/user-order-api/internal/repository/order_db"
	"github.com/IlyushinDM/user-order-api/internal/repository/user_db"
	"github.com/IlyushinDM/user-order-api/internal/services/order_service"
	"github.com/IlyushinDM/user-order-api/internal/services/user_service"
	conf_u "github.com/IlyushinDM/user-order-api/internal/utils/config_util"
	log_u "github.com/IlyushinDM/user-order-api/internal/utils/logger_util"

	_ "github.com/IlyushinDM/user-order-api/docs"
)

// Объявления Swagger
// @title User Order API
// @version 1.0
// @description Пример сервера для управления пользователями и их заказами.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in заголовок
// @name Авторизация
// @description Введите "Bearer" с пробелом и JWT токеном. Пример: "Bearer {token}"
func main() {
	// --- Конфигурация и логирование ---
	log := log_u.SetupLogger()
	conf_u.LoadConfig(log)

	// --- Подключение к базе данных ---
	db, err := database.InitDB(log)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// --- Инъекция зависимостей ---
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpStr := os.Getenv("JWT_EXPIRATION")
	jwtExp, err := strconv.Atoi(jwtExpStr)
	if err != nil || jwtExp <= 0 {
		jwtExp = 3600 // Дефолтное значение 3600 секунд (1 час)
		log.Warnf("Некорректное JWT_EXPIRATION, используется значение по умолчанию: %d секунд", jwtExp)
	}

	// Инициализация репозиториев
	userRepo := user_db.NewGormUserRepository(db, log)
	orderRepo := order_db.NewGormOrderRepository(db, log)

	// Инициализация сервисов
	userService := user_service.NewUserService(userRepo, log, jwtSecret, jwtExp)
	orderService := order_service.NewOrderService(orderRepo, userRepo, log)

	// Инициализация обработчиков
	userHandler := user_handler.NewUserHandler(userService, log)
	orderHandler := order_handler.NewOrderHandler(orderService, log)

	// --- Настройка роутера Gin ---
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// --- Middleware ---
	router.Use(gin.Recovery())
	router.Use(log_mw.LoggerMiddleware(log))

	// --- Маршруты ---

	// Документация Swagger (публичная)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Маршруты аутентификации (публичные)
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", userHandler.LoginUser)
		authRoutes.POST("/register", userHandler.CreateUser)
	}

	// API маршруты (защищенные JWT)
	api := router.Group("/api")
	api.Use(auth_mw.AuthMiddleware(log))
	{
		// Маршруты пользователей
		userRoutes := api.Group("/users")
		{
			userRoutes.GET("", userHandler.GetAllUsers)

			userRoutes.GET("/:id", userHandler.GetUserByID)
			userRoutes.PUT("/:id", userHandler.UpdateUser)
			userRoutes.DELETE("/:id", userHandler.DeleteUser)

			// Маршруты заказов для конкретного пользователя
			userRoutes.POST("/:id/orders", orderHandler.CreateOrder)
			userRoutes.GET("/:id/orders", orderHandler.GetAllOrdersByUser)
			userRoutes.GET("/:id/orders/:orderID", orderHandler.GetOrderByID)
			userRoutes.PUT("/:id/orders/:orderID", orderHandler.UpdateOrder)
			userRoutes.DELETE("/:id/orders/:orderID", orderHandler.DeleteOrder)
		}
	}

	// --- Запуск сервера ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Infof("Сервер запускается на порту %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
