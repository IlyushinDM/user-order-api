package core

import (
	auth_mw "github.com/IlyushinDM/user-order-api/internal/middleware/auth_middleware"
	log_mw "github.com/IlyushinDM/user-order-api/internal/middleware/logger_middleware"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// setupRouter настраивает маршруты для Gin
func SetupRouter(app *App) *gin.Engine {
	// Устанавливаем режим Gin из конфигурации
	gin.SetMode(app.Config.GinMode)

	router := gin.New()

	// Подключение middleware
	router.Use(gin.Recovery())
	// Передаем экземпляр логгера в middleware
	router.Use(log_mw.LoggerMiddleware(app.Logger))

	// Маршрут для документации Swagger
	// @BasePath /
	// @schemes http https
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Публичные маршруты (не требуют аутентификации)
	// Маршрут для входа пользователя
	router.POST("/auth/login", app.UserHandler.LoginUser)
	// Маршрут для создания пользователя
	router.POST("/api/users", app.UserHandler.CreateUser)

	// Защищенные маршруты API (требуют аутентификации)
	api := router.Group("/api")
	// Передаем JWT Secret из конфигурации в AuthMiddleware
	api.Use(auth_mw.AuthMiddleware(app.Logger, app.Config.JWTSecret))
	{
		// Маршруты для работы с пользователями
		userRoutes := api.Group("/users")
		{
			userRoutes.GET("", app.UserHandler.GetAllUsers)
			userRoutes.GET("/:id", app.UserHandler.GetUserByID)
			userRoutes.PUT("/:id", app.UserHandler.UpdateUser)
			userRoutes.DELETE("/:id", app.UserHandler.DeleteUser)

			// Маршруты для работы с заказами конкретного пользователя
			userRoutes.POST("/:id/orders", app.OrderHandler.CreateOrder)
			userRoutes.GET("/:id/orders", app.OrderHandler.GetAllOrdersByUser)
			userRoutes.GET("/:id/orders/:orderID", app.OrderHandler.GetOrderByID)
			userRoutes.PUT("/:id/orders/:orderID", app.OrderHandler.UpdateOrder)
			userRoutes.DELETE("/:id/orders/:orderID", app.OrderHandler.DeleteOrder)
		}
	}
	return router
}
