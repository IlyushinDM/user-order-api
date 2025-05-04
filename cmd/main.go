package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// Placeholder imports for project structure
	"github.com/IlyushinDM/user-order-api/internal/middleware"
	"github.com/IlyushinDM/user-order-api/internal/repository"
	"github.com/IlyushinDM/user-order-api/internal/services"
	"github.com/IlyushinDM/user-order-api/internal/utils"
	"github.com/IlyushinDM/user-order-apioject/internal/handlers"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize Logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("Application starting...")

	// Database connection
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE") // ???

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}

	logrus.Info("Database connection established")

	// Auto-migrate database schema (optional, typically handled by migrations)
	// err = db.AutoMigrate(&models.User{}, &models.Order{})
	// if err != nil {
	// 	logrus.Fatalf("Failed to auto-migrate database: %v", err)
	// }
	// logrus.Info("Database auto-migration completed")

	// Initialize repositories, services, and handlers
	userRepo := repository.NewUserDatabase(db)
	orderRepo := repository.NewOrderDatabase(db)

	userService := services.NewUserService(userRepo)
	orderService := services.NewOrderService(orderRepo)

	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService) // Assuming an auth handler exists
	orderHandler := handlers.NewOrderHandler(orderService)

	// Initialize Gin router
	router := gin.Default()

	// Global Middleware
	router.Use(utils.LoggerMiddleware(), gin.Recovery()) // Assuming LoggerMiddleware and Recovery are in utils

	// Public routes (e.g., authentication)
	public := router.Group("/auth")
	{
		public.POST("/login", authHandler.Login)
	}

	// Protected routes (require JWT authentication)
	api := router.Group("/api")
	api.Use(middleware.AuthRequired()) // Assuming AuthRequired is in middleware
	{
		// User routes
		api.POST("/users", userHandler.CreateUser)
		api.GET("/users", userHandler.GetUsers) // Includes pagination/filtering
		api.GET("/users/:id", userHandler.GetUserByID)
		api.PUT("/users/:id", userHandler.UpdateUser)
		api.DELETE("/users/:id", userHandler.DeleteUser)

		// Order routes
		api.POST("/users/:user_id/orders", orderHandler.CreateOrder)
		api.GET("/users/:user_id/orders", orderHandler.GetOrdersByUserID)
	}

	// Optional: Swagger documentation route
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // Requires setting up Swagger

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	logrus.Infof("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		logrus.Fatalf("Failed to run server: %v", err)
	}
}
