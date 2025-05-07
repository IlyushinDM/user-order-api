package database

import (
	"fmt"
	"os"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/utils/logger_util"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// initDB инициализирует подключение к базе данных через GORM.
// Параметры:
//   - log *logrus.Logger: логгер для записи сообщений
//
// Возвращает:
//   - *gorm.DB: подключение к БД
//   - error: ошибка, если подключение не удалось
func InitDB(log *logrus.Logger) (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=UTC",
		dbHost, dbPort, dbUser, dbName, dbPassword,
	)

	// Настройка уровня логирования GORM в зависимости от уровня Logrus
	gormLogLevel := gormlogger.Silent
	if log.GetLevel() >= logrus.InfoLevel {
		gormLogLevel = gormlogger.Info
	}
	if log.GetLevel() >= logrus.WarnLevel {
		gormLogLevel = gormlogger.Warn
	}

	gormWriter := &logger_util.LogrusGormWriter{Logger: log}

	newLogger := gormlogger.New(
		gormWriter,
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond, // Порог для медленных запросов
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Errorf("Ошибка подключения GORM: %v", err)
		return nil, err
	}

	// Автомиграции (только для разработки)
	log.Info("Запуск автомиграций базы данных.")
	err = db.AutoMigrate(&user_model.User{}, &order_model.Order{})
	if err != nil {
		log.Fatalf("Ошибка автомиграции базы данных: %v", err)
	}
	log.Info("Автомиграции завершены.")

	// Настройки пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Ошибка получения sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info("Подключение к базе данных установлено.")
	return db, nil
}
