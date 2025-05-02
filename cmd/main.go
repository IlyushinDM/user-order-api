package main

import (
	"github.com/IlyushinDM/user-order-api/internal/internal/handlers"
	"github.com/IlyushinDM/user-order-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	router := gin.Default()

	// Публичные маршруты
	router.POST("/auth/login", handlers.Login)
	router.POST("/users", handlers.CreateUser) // Без JWT-аутентификации

	// Защищенные маршруты
	authorized := router.Group("/")
	authorized.Use(middleware.JWTAuthMiddleware())
	{
		// Маршруты пользователей
		authorized.GET("/users", handlers.GetUsers)
		authorized.GET("/users/:id", handlers.GetUser)
		authorized.PUT("/users/:id", handlers.UpdateUser)
		authorized.DELETE("/users/:id", handlers.DeleteUser)

		// Маршруты заказов
		authorized.POST("/users/:user_id/orders", handlers.CreateOrder)
		authorized.GET("/users/:user_id/orders", handlers.GetUserOrders)
	}

	return router
}
