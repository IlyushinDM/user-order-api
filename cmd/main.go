package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/IlyushinDM/user-order-api/docs" // Swagger docs generated path
	"github.com/IlyushinDM/user-order-api/internal/handlers"
	"github.com/IlyushinDM/user-order-api/internal/middleware"
	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/repository"
	"github.com/IlyushinDM/user-order-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// @title User Order API
// @version 1.0
// @description This is a sample server for managing users and their orders.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer {token}"
func main() {
	// --- Configuration & Logging ---
	log := setupLogger()
	loadConfig(log) // Load .env file

	// --- Database Connection ---
	db, err := initDB(log)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// --- Dependency Injection ---
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpStr := os.Getenv("JWT_EXPIRATION") // Default to 1 hour
	jwtExp, err := strconv.Atoi(jwtExpStr)
	if err != nil || jwtExp <= 0 {
		jwtExp = 3600 // Default to 3600 seconds (1 hour)
		log.Warnf("Invalid or missing JWT_EXPIRATION, defaulting to %d seconds", jwtExp)
	}

	// Repositories
	userRepo := repository.NewGormUserRepository(db, log)
	orderRepo := repository.NewGormOrderRepository(db, log)

	// Services
	userService := services.NewUserService(userRepo, log, jwtSecret, jwtExp)
	orderService := services.NewOrderService(orderRepo, userRepo, log) // Pass userRepo if needed by order service

	// Handlers
	userHandler := handlers.NewUserHandler(userService, log)
	orderHandler := handlers.NewOrderHandler(orderService, log)

	// --- Gin Router Setup ---
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New() // Use gin.New() instead of gin.Default() for custom middleware setup

	// --- Middleware ---
	router.Use(gin.Recovery())                   // Recover from panics
	router.Use(middleware.LoggerMiddleware(log)) // Log requests using Logrus

	// --- Routes ---
	// Public routes (Swagger, Auth)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", userHandler.LoginUser)
		// Potentially add /register here if CreateUser shouldn't be protected
	}

	// API v1 routes (protected by JWT)
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(log)) // Apply JWT authentication to all /api/v1 routes
	{
		// User routes (already protected by group middleware)
		userRoutes := api.Group("/users")
		{
			userRoutes.POST("", userHandler.CreateUser) // Keep POST /users public or move under /auth/register? Decision: Keep here for now, relies on service logic. Or move register to /auth
			userRoutes.GET("", userHandler.GetAllUsers)
			userRoutes.GET("/:id", userHandler.GetUserByID)
			userRoutes.PUT("/:id", userHandler.UpdateUser)
			userRoutes.DELETE("/:id", userHandler.DeleteUser)
		}

		// Order routes (already protected by group middleware)
		orderRoutes := api.Group("/orders")
		{
			orderRoutes.POST("", orderHandler.CreateOrder)
			orderRoutes.GET("", orderHandler.GetAllOrdersByUser) // Gets orders for the authenticated user
			orderRoutes.GET("/:id", orderHandler.GetOrderByID)   // Gets a specific order for the authenticated user
			orderRoutes.PUT("/:id", orderHandler.UpdateOrder)
			orderRoutes.DELETE("/:id", orderHandler.DeleteOrder)
		}
	}

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Infof("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupLogger configures the Logrus logger.
func setupLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{}) // Use JSON format for structured logging
	// log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true}) // Or Text format

	log.SetOutput(os.Stdout) // Log to standard output

	levelStr := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		log.Warnf("Invalid LOG_LEVEL '%s', defaulting to 'info'", levelStr)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	return log
}

// loadConfig loads environment variables from .env file.
func loadConfig(log *logrus.Logger) {
	err := godotenv.Load() // Loads .env file from current directory by default
	if err != nil {
		log.Warn("Error loading .env file, using system environment variables")
		// Don't fail if .env is not present, might be using system env vars
	}
}

// --- Logger Adapter ---
// logrusGormWriter adapts logrus logger to GORM's logger interface
type logrusGormWriter struct {
	logger *logrus.Logger
}

// Printf implements gorm logger interface
func (w *logrusGormWriter) Printf(message string, data ...interface{}) {
	// You might want to adjust the log level based on the message content
	// or GORM's context, but for simplicity, using Info or Debug level
	w.logger.Tracef(message, data...) // Use Tracef or Debugf for GORM logs
}

// --- End Logger Adapter ---

// initDB initializes the database connection using GORM.
func initDB(log *logrus.Logger) (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=UTC", // Added TimeZone
		dbHost, dbPort, dbUser, dbName, dbPassword,
	)

	// Configure GORM logger level based on Logrus level
	gormLogLevel := gormlogger.Silent
	if log.GetLevel() >= logrus.InfoLevel {
		gormLogLevel = gormlogger.Info
	}
	if log.GetLevel() >= logrus.WarnLevel {
		gormLogLevel = gormlogger.Warn // GORM's Warn logs errors too
	}
	// if log.GetLevel() >= logrus.ErrorLevel { // Map logrus Error to GORM Error
	// 	gormLogLevel = gormlogger.Error
	// }

	// Create the custom writer adapter
	gormWriter := &logrusGormWriter{logger: log}

	// Configure the GORM logger using the custom writer
	newLogger := gormlogger.New(
		gormWriter, // Use the custom writer adapter HERE
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond, // Adjust slow query threshold
			LogLevel:                  gormLogLevel,           // Set log level based on Logrus level
			IgnoreRecordNotFoundError: true,                   // Don't log ErrRecordNotFound as errors
			ParameterizedQueries:      false,                  // Log SQL queries with parameters (set to true for production if desired)
			Colorful:                  false,                  // Disable colors for JSON output compatibility
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, // Use the configured GORM logger
	})
	if err != nil {
		log.Errorf("Failed GORM connection: %v", err)
		return nil, err
	}

	// Auto migrate the schema (Simple approach for development)
	// For production, use a dedicated migration tool (e.g., goose, sql-migrate)
	log.Info("Running database auto-migrations...")
	err = db.AutoMigrate(&models.User{}, &models.Order{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	log.Info("Database auto-migration completed.")

	// Optional: Connection Pooling settings
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info("Database connection established successfully.")
	return db, nil
}

// Make sure other files (user_model.go, user_handler.go, etc.) remain the same
// as in the previous response unless other changes are needed.
