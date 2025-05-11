package core

import (
	"fmt"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/handlers/order_handler"
	"github.com/IlyushinDM/user-order-api/internal/handlers/user_handler"
	"github.com/IlyushinDM/user-order-api/internal/repository/database"
	"github.com/IlyushinDM/user-order-api/internal/repository/order_rep"
	"github.com/IlyushinDM/user-order-api/internal/repository/user_rep"
	"github.com/IlyushinDM/user-order-api/internal/services/order_service"
	"github.com/IlyushinDM/user-order-api/internal/services/user_service"
	"github.com/IlyushinDM/user-order-api/internal/utils/config_util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// NewAppWithInitDB allows injecting a custom InitDB function for testing.
func NewAppWithInitDB(logger *logrus.Logger, config *config_util.Config, initDB func(cfg *config_util.Config, log *logrus.Logger) (*gorm.DB, error)) (*App, error) {
	db, err := initDB(config, logger)
	if err != nil {
		return nil, err
	}

	commonHandler := common_handler.NewCommonHandler(logger)
	userHandler := user_handler.NewUserHandler(nil, commonHandler, logger)
	orderHandler := order_handler.NewOrderHandler(nil, commonHandler, logger)

	return &App{
		Logger:       logger,
		Config:       config,
		DB:           db,
		UserHandler:  userHandler,
		OrderHandler: orderHandler,
	}, nil
}

// App содержит основные компоненты приложения
type App struct {
	Config       *config_util.Config
	Logger       *logrus.Logger
	DB           *gorm.DB
	Router       *gin.Engine
	UserHandler  *user_handler.UserHandler
	OrderHandler *order_handler.OrderHandler
}

// NewApp создает и инициализирует новый экземпляр приложения
func NewApp(logger *logrus.Logger, config *config_util.Config) (*App, error) {
	// Подключение к базе данных
	db, err := database.InitDB(config, logger)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных '%s'", err)
	} else {
		logger.Infof("База данных успешно подключена")
	}

	// Инициализация репозиториев
	userRepo := user_rep.NewGormUserRepository(db, logger)
	orderRepo := order_rep.NewGormOrderRepository(db, logger)

	// Инициализация сервисов
	userService := user_service.NewUserService(
		userRepo,
		logger,
		config.JWTSecret,
		int(config.JWTExpiration/time.Second))
	orderService := order_service.NewOrderService(orderRepo, logger)

	// Инициализация common handler
	commonHandler := common_handler.NewCommonHandler(logger)

	// Инициализация обработчиков
	userHandler := user_handler.NewUserHandler(userService, commonHandler, logger)
	orderHandler := order_handler.NewOrderHandler(orderService, commonHandler, logger)

	app := &App{
		Config:       config,
		Logger:       logger,
		DB:           db,
		UserHandler:  userHandler,
		OrderHandler: orderHandler,
	}

	return app, nil
}
