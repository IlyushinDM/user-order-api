package config_util

// import (
// 	"bytes"
// 	"os"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// )

// // testLogger создает логгер, который пишет в буфер, чтобы мы могли проверить его вывод.
// func testLogger(buf *bytes.Buffer) *logrus.Logger {
// 	log := logrus.New()
// 	log.SetOutput(buf)
// 	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableColors: true}) // Простой формат для тестов
// 	log.SetLevel(logrus.DebugLevel)                                                      // Логируем все для тестов
// 	return log
// }

// // Helper для очистки переменных окружения, установленных в тесте.
// // Используйте t.Setenv (Go 1.17+) для автоматической очистки.
// // Если Go < 1.17, нужно делать это вручную:
// func unsetEnvVars(vars ...string) {
// 	for _, v := range vars {
// 		os.Unsetenv(v)
// 	}
// }

// func TestLoadConfig_Defaults(t *testing.T) {
// 	// Убедимся, что переменные не установлены (особенно важно, если тесты запускаются параллельно или в CI)
// 	// Используем t.Setenv для автоматической очистки после теста (Go 1.17+)
// 	// Если Go < 1.17, используйте os.Unsetenv и восстанавливайте значения.
// 	if os.Getenv("GO_VERSION") < "1.17" { // Примерная проверка, лучше использовать build tags
// 		unsetEnvVars("JWT_SECRET", "JWT_EXPIRATION", "PORT", "GIN_MODE")
// 	} else {
// 		t.Setenv("JWT_SECRET", "")
// 		t.Setenv("JWT_EXPIRATION", "")
// 		t.Setenv("PORT", "")
// 		t.Setenv("GIN_MODE", "")
// 	}

// 	var logBuf bytes.Buffer
// 	logger := testLogger(&logBuf)

// 	cfg, err := LoadConfig(logger)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, cfg)

// 	// Проверяем значения по умолчанию
// 	assert.Equal(t, "your-very-secret-key-for-dev-only", cfg.JWTSecret, "Default JWTSecret mismatch")
// 	assert.Equal(t, 3600, cfg.JWTExpiration, "Default JWTExpiration mismatch")
// 	assert.Equal(t, "8080", cfg.Port, "Default Port mismatch")
// 	assert.Equal(t, gin.DebugMode, cfg.GinMode, "Default GinMode mismatch")

// 	// Проверяем логирование предупреждений/информации о значениях по умолчанию
// 	logOutput := logBuf.String()
// 	assert.Contains(t, logOutput, "JWT_SECRET не установлен", "Log for default JWTSecret missing")
// 	assert.Contains(t, logOutput, "JWT_EXPIRATION не установлено", "Log for default JWTExpiration missing")
// 	assert.Contains(t, logOutput, "PORT не установлен", "Log for default Port missing")
// 	assert.Contains(t, logOutput, "GIN_MODE не установлен", "Log for default GinMode missing")
// }

// func TestLoadConfig_FromEnvironment(t *testing.T) {
// 	// Сохраняем оригинальные значения, чтобы восстановить их после теста
// 	// Это более надежно, если t.Setenv недоступен или для сложных сценариев.
// 	originalJwtSecret, JwtSecret := os.LookupEnv("JWT_SECRET")
// 	originalJwtExp, JwtExp := os.LookupEnv("JWT_EXPIRATION")
// 	originalPort, Port := os.LookupEnv("PORT")
// 	originalGinMode, GinMode := os.LookupEnv("GIN_MODE")

// 	defer func() { // Восстанавливаем переменные окружения
// 		if JwtSecret {
// 			os.Setenv("JWT_SECRET", originalJwtSecret)
// 		} else {
// 			os.Unsetenv("JWT_SECRET")
// 		}
// 		if JwtExp {
// 			os.Setenv("JWT_EXPIRATION", originalJwtExp)
// 		} else {
// 			os.Unsetenv("JWT_EXPIRATION")
// 		}
// 		if Port {
// 			os.Setenv("PORT", originalPort)
// 		} else {
// 			os.Unsetenv("PORT")
// 		}
// 		if GinMode {
// 			os.Setenv("GIN_MODE", originalGinMode)
// 		} else {
// 			os.Unsetenv("GIN_MODE")
// 		}
// 	}()

// 	// Устанавливаем тестовые значения
// 	os.Setenv("JWT_SECRET", "my-test-secret")
// 	os.Setenv("JWT_EXPIRATION", "1800")
// 	os.Setenv("PORT", "9090")
// 	os.Setenv("GIN_MODE", gin.ReleaseMode)

// 	var logBuf bytes.Buffer
// 	logger := testLogger(&logBuf) // Используем пустой буфер для этого теста

// 	cfg, err := LoadConfig(logger)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, cfg)

// 	assert.Equal(t, "my-test-secret", cfg.JWTSecret)
// 	assert.Equal(t, 1800, cfg.JWTExpiration)
// 	assert.Equal(t, "9090", cfg.Port)
// 	assert.Equal(t, gin.ReleaseMode, cfg.GinMode)

// 	// Убедимся, что предупреждений о значениях по умолчанию нет
// 	logOutput := logBuf.String()
// 	assert.NotContains(t, logOutput, "не установлен")
// 	assert.NotContains(t, logOutput, "используется значение по умолчанию")
// }

// func TestLoadConfig_InvalidJwtExpiration(t *testing.T) {
// 	if os.Getenv("GO_VERSION") < "1.17" {
// 		originalJwtExp, has := os.LookupEnv("JWT_EXPIRATION")
// 		defer func() {
// 			if has {
// 				os.Setenv("JWT_EXPIRATION", originalJwtExp)
// 			} else {
// 				os.Unsetenv("JWT_EXPIRATION")
// 			}
// 		}()
// 		os.Setenv("JWT_EXPIRATION", "not-a-number")
// 	} else {
// 		t.Setenv("JWT_EXPIRATION", "not-a-number")
// 		t.Setenv("JWT_SECRET", "temp-secret") // Устанавливаем, чтобы не было лога о JWT_SECRET
// 		t.Setenv("PORT", "7070")              // Устанавливаем, чтобы не было лога о PORT
// 		t.Setenv("GIN_MODE", gin.TestMode)    // Устанавливаем, чтобы не было лога о GIN_MODE
// 	}

// 	var logBuf bytes.Buffer
// 	logger := testLogger(&logBuf)

// 	cfg, err := LoadConfig(logger)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, cfg)

// 	// Должно использоваться значение по умолчанию
// 	assert.Equal(t, 3600, cfg.JWTExpiration)

// 	// Проверяем лог
// 	logOutput := logBuf.String()
// 	assert.Contains(t, logOutput, "Некорректное JWT_EXPIRATION ('not-a-number')")
// 	assert.Contains(t, logOutput, "используется значение по умолчанию: 3600 секунд")
// }

// func TestLoadConfig_ZeroJwtExpiration(t *testing.T) {
// 	if os.Getenv("GO_VERSION") < "1.17" {
// 		originalJwtExp, has := os.LookupEnv("JWT_EXPIRATION")
// 		defer func() {
// 			if has {
// 				os.Setenv("JWT_EXPIRATION", originalJwtExp)
// 			} else {
// 				os.Unsetenv("JWT_EXPIRATION")
// 			}
// 		}()
// 		os.Setenv("JWT_EXPIRATION", "0")
// 	} else {
// 		t.Setenv("JWT_EXPIRATION", "0")
// 		t.Setenv("JWT_SECRET", "temp-secret")
// 		t.Setenv("PORT", "7070")
// 		t.Setenv("GIN_MODE", gin.TestMode)
// 	}

// 	var logBuf bytes.Buffer
// 	logger := testLogger(&logBuf)

// 	cfg, err := LoadConfig(logger)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, cfg)

// 	assert.Equal(t, 3600, cfg.JWTExpiration) // Должно использоваться значение по умолчанию

// 	logOutput := logBuf.String()
// 	assert.Contains(t, logOutput, "Некорректное JWT_EXPIRATION ('0')")
// }

// func TestLoadConfig_InvalidGinMode(t *testing.T) {
// 	if os.Getenv("GO_VERSION") < "1.17" {
// 		originalGinMode, has := os.LookupEnv("GIN_MODE")
// 		defer func() {
// 			if has {
// 				os.Setenv("GIN_MODE", originalGinMode)
// 			} else {
// 				os.Unsetenv("GIN_MODE")
// 			}
// 		}()
// 		os.Setenv("GIN_MODE", "invalid-mode")
// 	} else {
// 		t.Setenv("GIN_MODE", "invalid-mode")
// 		t.Setenv("JWT_SECRET", "temp-secret")
// 		t.Setenv("JWT_EXPIRATION", "300")
// 		t.Setenv("PORT", "7070")
// 	}

// 	var logBuf bytes.Buffer
// 	logger := testLogger(&logBuf)

// 	cfg, err := LoadConfig(logger)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, cfg)

// 	assert.Equal(t, gin.DebugMode, cfg.GinMode) // Должно использоваться значение по умолчанию

// 	logOutput := logBuf.String()
// 	assert.Contains(t, logOutput, "Некорректный GIN_MODE ('invalid-mode')")
// 	assert.Contains(t, logOutput, "используется значение по умолчанию: debug")
// }

// func TestLoadConfig_EmptyPort(t *testing.T) {
// 	// Этот тест дублирует часть TestLoadConfig_Defaults, но фокусируется только на PORT
// 	if os.Getenv("GO_VERSION") < "1.17" {
// 		originalPort, has := os.LookupEnv("PORT")
// 		defer func() {
// 			if has {
// 				os.Setenv("PORT", originalPort)
// 			} else {
// 				os.Unsetenv("PORT")
// 			}
// 		}()
// 		os.Unsetenv("PORT") // Убеждаемся, что не установлено
// 	} else {
// 		t.Setenv("PORT", "")
// 		t.Setenv("JWT_SECRET", "temp-secret")
// 		t.Setenv("JWT_EXPIRATION", "300")
// 		t.Setenv("GIN_MODE", gin.TestMode)
// 	}

// 	var logBuf bytes.Buffer
// 	logger := testLogger(&logBuf)

// 	cfg, err := LoadConfig(logger)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, cfg)

// 	assert.Equal(t, "8080", cfg.Port)

// 	logOutput := logBuf.String()
// 	assert.Contains(t, logOutput, "PORT не установлен, используется значение по умолчанию: 8080")
// }
