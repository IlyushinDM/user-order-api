package config_util

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Установка требуемых переменных env для конфигурации
func setRequiredEnv() func() {
	envs := map[string]string{
		"DB_HOST":        "localhost",
		"DB_PORT":        "5432",
		"DB_USER":        "user",
		"DB_NAME":        "dbname",
		"DB_PASSWORD":    "password",
		"JWT_SECRET":     "secret",
		"JWT_EXPIRATION": "1h",
	}
	originals := make(map[string]string)
	for k, v := range envs {
		originals[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	return func() {
		for k, v := range originals {
			os.Setenv(k, v)
		}
	}
}

func TestLoadConfig_SuccessFromEnv(t *testing.T) {
	cleanup := setRequiredEnv()
	defer cleanup()

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	cfg, err := LoadConfig(log)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "user", cfg.DBUser)
	assert.Equal(t, "dbname", cfg.DBName)
	assert.Equal(t, "password", cfg.DBPassword)
	assert.Equal(t, "secret", cfg.JWTSecret)
	assert.Equal(t, time.Hour, cfg.JWTExpiration)
	// Check defaults
	assert.Equal(t, "release", cfg.AppEnv)
	assert.Equal(t, "release", cfg.GinMode)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, 10, cfg.DBMaxIdleConns)
	assert.Equal(t, 100, cfg.DBMaxOpenConns)
	assert.Equal(t, time.Hour, cfg.DBConnMaxLifetime)
	assert.Equal(t, 30*time.Minute, cfg.DBConnMaxIdleTime)
	assert.Equal(t, 5, cfg.ReadTimeout)
	assert.Equal(t, 10, cfg.WriteTimeout)
	assert.Equal(t, 60, cfg.IdleTimeout)
	assert.Equal(t, 1048576, cfg.MaxHeaderBytes)
	assert.Equal(t, 15*time.Second, cfg.ShutdownTimeout)
}

func TestLoadConfig_MissingRequiredEnv(t *testing.T) {
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_EXPIRATION")

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	cfg, err := LoadConfig(log)
	assert.Nil(t, cfg)
	assert.Error(t, err)
}

func TestLoadConfig_NilLogger(t *testing.T) {
	cleanup := setRequiredEnv()
	defer cleanup()

	cfg, err := LoadConfig(nil)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "логгер не предоставлен")
}

func TestLoadConfig_OverridesDefaults(t *testing.T) {
	cleanup := setRequiredEnv()
	defer cleanup()
	os.Setenv("APP_ENV", "dev")
	os.Setenv("GIN_MODE", "debug")
	os.Setenv("PORT", "1234")
	os.Setenv("DB_MAX_IDLE_CONNS", "20")
	os.Setenv("DB_MAX_OPEN_CONNS", "200")
	os.Setenv("DB_CONN_MAX_LIFETIME", "2h")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1h")
	os.Setenv("HTTP_READ_TIMEOUT", "15")
	os.Setenv("HTTP_WRITE_TIMEOUT", "20")
	os.Setenv("HTTP_IDLE_TIMEOUT", "120")
	os.Setenv("HTTP_MAX_HEADER_BYTES", "2048")
	os.Setenv("SHUTDOWN_TIMEOUT", "30s")

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	cfg, err := LoadConfig(log)
	assert.NoError(t, err)
	assert.Equal(t, "dev", cfg.AppEnv)
	assert.Equal(t, "debug", cfg.GinMode)
	assert.Equal(t, "1234", cfg.Port)
	assert.Equal(t, 20, cfg.DBMaxIdleConns)
	assert.Equal(t, 200, cfg.DBMaxOpenConns)
	assert.Equal(t, 2*time.Hour, cfg.DBConnMaxLifetime)
	assert.Equal(t, time.Hour, cfg.DBConnMaxIdleTime)
	assert.Equal(t, 15, cfg.ReadTimeout)
	assert.Equal(t, 20, cfg.WriteTimeout)
	assert.Equal(t, 120, cfg.IdleTimeout)
	assert.Equal(t, 2048, cfg.MaxHeaderBytes)
	assert.Equal(t, 30*time.Second, cfg.ShutdownTimeout)
}
