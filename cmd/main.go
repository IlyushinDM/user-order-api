package main

import (
	"github.com/IlyushinDM/user-order-api/internal/handlers"
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

// import (
// 	"log"
// 	"os"

// 	"your_module_path/internal/handlers" // Replace with your module path
// 	"your_module_path/internal/repository" // Replace with your module path
// 	"your_module_path/internal/services" // Replace with your module path
// 	"your_module_path/internal/middleware" // Replace with your module path
// 	"your_module_path/internal/models" // Replace with your module path

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// 	"github.com/joho/godotenv" // Optional: for loading .env file
// )

// func main() {
// 	// Load environment variables from .env file (optional)
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("No .env file found, reading from environment")
// 	}

// 	// Database connection
// 	dsn := os.Getenv("DATABASE_URL") // Get connection string from environment variable
// 	if dsn == "" {
// 		log.Fatal("DATABASE_URL environment variable not set")
// 	}

// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}

// 	// Auto-migrate the schema (for development/simple projects, migrations preferred for production)
// 	// In a real project with a migrations folder, you'd run migration scripts here or separately.
// 	err = db.AutoMigrate(&models.User{}, &models.Order{})
// 	if err != nil {
// 		log.Fatalf("Failed to auto-migrate database schema: %v", err)
// 	}

// 	// Initialize repositories
// 	userRepo := repository.NewUserRepository(db)
// 	orderRepo := repository.NewOrderRepository(db) // Assuming you create OrderRepository

// 	// Initialize services
// 	userService := services.NewUserService(userRepo)
// 	orderService := services.NewOrderService(orderRepo) // Assuming you create OrderService

// 	// Initialize handlers
// 	userHandler := handlers.NewUserHandler(userService)
// 	orderHandler := handlers.NewOrderHandler(orderService) // Assuming you create OrderHandler
// 	authHandler := handlers.NewAuthHandler(userService) // Assuming you create AuthHandler

// 	// Setup Gin router
// 	router := gin.Default()

// 	// Public routes
// 	router.POST("/auth/login", authHandler.Login) // Assuming Login handler exists
// 	router.POST("/users", userHandler.CreateUser) // User creation might be public or require admin auth

// 	// Authenticated routes
// 	authenticated := router.Group("/")
// 	authenticated.Use(middleware.JWTAuthMiddleware()) // Assuming JWTAuthMiddleware exists
// 	{
// 		authenticated.GET("/users", userHandler.ListUsers)
// 		authenticated.GET("/users/:id", userHandler.GetUserByID)
// 		authenticated.PUT("/users/:id", userHandler.UpdateUser)
// 		authenticated.DELETE("/users/:id", userHandler.DeleteUser)
// 		authenticated.POST("/users/:user_id/orders", orderHandler.CreateOrder)
// 		authenticated.GET("/users/:user_id/orders", orderHandler.ListOrders)
// 	}

// 	// Run the server
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080" // Default port if not specified
// 	}
// 	log.Printf("Server starting on port %s", port)
// 	if err := router.Run(":" + port); err != nil {
// 		log.Fatalf("Server failed to start: %v", err)
// 	}
// }
