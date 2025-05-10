package database

import (
	"errors" // Импортируем пакет errors для создания простых ошибок, если нужно
	"fmt"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/utils/config_util"
	"github.com/IlyushinDM/user-order-api/internal/utils/logger_util"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// InitDB инициализирует подключение к базе данных через GORM и настраивает пул соединений.
// Эта функция НЕ выполняет миграции.
func InitDB(cfg *config_util.Config, log *logrus.Logger) (*gorm.DB, error) {
	// Простая проверка входных параметров
	if cfg == nil {
		return nil, errors.New("конфигурация базы данных не предоставлена (cfg is nil)")
	}
	if log == nil {
		// В данном случае логирование здесь недоступно, возвращаем чистую ошибку
		return nil, errors.New("логгер не предоставлен (log is nil)")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBPassword,
	)

	// Настройка уровня логирования GORM в зависимости от уровня Logrus
	gormLogLevel := gormlogger.Silent // gormlogger.LogLevel по умолчанию - тихий
	if log.GetLevel() >= logrus.InfoLevel {
		gormLogLevel = gormlogger.Info // Логировать запросы при уровне Info или выше
	}
	// Примечание: gormlogger.LogLevel имеет только Silent, Error, Warn, Info.
	// Наиболее подробный стандартный уровень для логирования SQL - Info.
	// Поэтому уровни Debug и Trace Logrus будут также маппиться на gormlogger.Info
	// для включения подробных логов GORM.

	gormWriter := &logger_util.LogrusGormWriter{Logger: log}

	newLogger := gormlogger.New(
		gormWriter,
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond, // Порог для "медленных" запросов
			LogLevel:                  gormLogLevel,           // Установленный уровень логирования GORM
			IgnoreRecordNotFoundError: true,                   // Игнорировать ошибки gorm.ErrRecordNotFound в логах
			ParameterizedQueries:      true,                   // Включаем логирование параметризованных запросов (лучше для отладки)
			Colorful:                  false,                  // Отключаем цветной вывод GORM, если Logrus уже цветной или для единообразия
		},
	)

	// Открываем соединение с базой данных с настроенным логированием
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {

		log.WithError(err).Errorf("ошибка подключения GORM к базе данных: %v", err)
		return nil, fmt.Errorf("ошибка подключения GORM: %w", err)
	}

	// Настройки пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		// Возвращаем ошибку вместо log.Fatalf
		log.WithError(err).Errorf("ошибка получения *sql.DB из экземпляра GORM: %v", err)
		return nil, fmt.Errorf("ошибка получения sql.DB: %w", err)
	}

	// Устанавливаем настройки пула соединений из конфигурации или дефолтов
	// Использование значений из конфига предпочтительнее хардкода
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)       // Макс. количество простаивающих соединений
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)       // Макс. количество открытых соединений
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime) // Время жизни соединения
	sqlDB.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime) // Время простоя соединения перед закрытием (добавлено)

	// Проверка "живости" соединения (ping) - опционально, но полезно при старте
	if err := sqlDB.Ping(); err != nil {
		log.WithError(err).Errorf("ошибка проверки соединения (ping) с базой данных: %v", err)
		return nil, fmt.Errorf("ошибка проверки соединения с базой данных: %w", err)
	}

	log.Info("Подключение к базе данных установлено и пул соединений настроен успешно.")
	return db, nil
}

// RunMigrations выполняет автоматическую миграцию базы данных для заданных моделей.
// Эту функцию следует вызывать отдельно после успешной инициализации DB,
// обычно только в среде разработки или при явном сценарии миграции.
func RunMigrations(db *gorm.DB, log *logrus.Logger) error {
	// Проверка на nil DB
	if db == nil {
		// В данном случае логирование может быть недоступно, возвращаем чистую ошибку
		return errors.New("экземпляр *gorm.DB не предоставлен для выполнения миграций")
	}
	if log == nil {
		// Логгер не предоставлен, возвращаем чистую ошибку
		return errors.New("логгер не предоставлен для выполнения миграций")
	}

	log.Info("Запуск автомиграций базы данных для User и Order моделей.")
	// Выполняем автомиграцию. GORM создаст таблицы, если они не существуют,
	// и добавит недостающие колонки. Он НЕ удалит колонки и НЕ изменит их тип.
	err := db.AutoMigrate(&user_model.User{}, &order_model.Order{})
	if err != nil {
		// Логируем и возвращаем ошибку миграции
		log.WithError(err).Errorf("ошибка выполнения автомиграции базы данных: %v", err)
		return fmt.Errorf("ошибка автомиграции базы данных: %w", err)
	}
	log.Info("Автомиграции завершены успешно.")
	return nil
}
