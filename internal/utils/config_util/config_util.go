package config_util

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

// Config содержит все настройки конфигурации приложения.
// Значения могут быть загружены из переменных окружения или .env файла.
type Config struct {
	AppEnv  string `env:"APP_ENV" env-default:"release"`
	GinMode string `env:"GIN_MODE" env-default:"release"`
	Port    string `env:"PORT" env-default:"8080"`

	// Настройки базы данных
	DBHost            string        `env:"DB_HOST" env-required:"true"`
	DBPort            string        `env:"DB_PORT" env-required:"true"`
	DBUser            string        `env:"DB_USER" env-required:"true"`
	DBName            string        `env:"DB_NAME" env-required:"true"`
	DBPassword        string        `env:"DB_PASSWORD" env-required:"true"`
	DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" env-default:"10"`
	DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" env-default:"100"`
	DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" env-default:"1h"`
	DBConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" env-default:"30m"`

	// Настройки JWT
	JWTSecret     string        `env:"JWT_SECRET" env-required:"true"`
	JWTExpiration time.Duration `env:"JWT_EXPIRATION" env-required:"true"` // cleanenv can parse "1h" directly into time.Duration

	// Настройки HTTP сервера
	// Note: Original struct used int, .env values are numbers.
	// cleanenv will parse these numbers as int.
	ReadTimeout    int `env:"HTTP_READ_TIMEOUT" env-default:"5"`
	WriteTimeout   int `env:"HTTP_WRITE_TIMEOUT" env-default:"10"`
	IdleTimeout    int `env:"HTTP_IDLE_TIMEOUT" env-default:"60"`
	MaxHeaderBytes int `env:"HTTP_MAX_HEADER_BYTES" env-default:"1048576"` // 1MB

	// Таймаут для graceful shutdown
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"15s"` // cleanenv parses "15s" into time.Duration
}

// LoadConfig загружает конфигурацию приложения.
// Сначала пытается прочитать из .env файла, затем переопределяет переменными окружения.
// Если .env файл не найден, загружает только из переменных окружения.
// Проверяет обязательные поля и применяет дефолты в обоих случаях.
// Возвращает указатель на структуру Config или ошибку.
func LoadConfig(log *logrus.Logger) (*Config, error) {
	if log == nil {
		return nil, errors.New("логгер не предоставлен для загрузки конфигурации")
	}

	var cfg Config

	// Попытка загрузки из .env файла и переменных окружения
	err := cleanenv.ReadConfig(".env", &cfg)
	// Проверка ошибки загрузки
	if err != nil {
		// Если ошибка - файл не найден (.env), то это не критично.
		// В этом случае, пытаемся загрузить только из переменных окружения.
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("Файл .env не найден. Попытка загрузки конфигурации только из переменных окружения.")

			// Явно загружаем из переменных окружения.
			// ReadEnv также применяет дефолты и проверяет обязательные поля.
			err = cleanenv.ReadEnv(&cfg)
			if err != nil {
				// Если загрузка из переменных окружения не удалась (например, отсутствуют обязательные переменные),
				// это критическая ошибка.
				log.WithError(err).Errorf("Критическая ошибка загрузки конфигурации из переменных окружения")
				return nil, fmt.Errorf("не удалось загрузить конфигурацию из переменных окружения: %w", err)
			}
			// Если ReadEnv успешно завершился, err теперь nil, и мы продолжим.

		} else {
			// Если ошибка не является ошибкой "файл не найден", это критическая ошибка
			// загрузки (например, ошибка парсинга .env, или отсутствующие обязательные поля
			// даже после попытки cleanenv прочитать из окружения в первом вызове).
			log.WithError(err).Errorf("Критическая ошибка загрузки конфигурации из .env или переменных окружения (не связана с отсутствием файла)")
			return nil, fmt.Errorf("не удалось загрузить конфигурацию: %w", err)
		}
	}

	// Если мы дошли сюда без возврата ошибки, значит, конфигурация успешно загружена
	// либо из .env + env, либо только из env. Обязательные поля были проверены.
	log.Info("Конфигурация успешно загружена")

	// Логирование загруженных значений для отладки
	log.Debugf("APP_ENV: %s", cfg.AppEnv)
	log.Infof("DB_HOST: %s, DB_PORT: %s, DB_NAME: %s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Debugf("DB_MAX_IDLE_CONNS: %d, DB_MAX_OPEN_CONNS: %d", cfg.DBMaxIdleConns, cfg.DBMaxOpenConns)
	log.Debugf("DB_CONN_MAX_LIFETIME: %s, DB_CONN_MAX_IDLE_TIME: %s", cfg.DBConnMaxLifetime, cfg.DBConnMaxIdleTime)
	log.Debugf("JWT_EXPIRATION: %s", cfg.JWTExpiration)
	log.Debugf("HTTP_READ_TIMEOUT: %d, HTTP_WRITE_TIMEOUT: %d, HTTP_IDLE_TIMEOUT: %d, HTTP_MAX_HEADER_BYTES: %d",
		cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout, cfg.MaxHeaderBytes)
	log.Debugf("SHUTDOWN_TIMEOUT: %s", cfg.ShutdownTimeout)

	return &cfg, nil
}
