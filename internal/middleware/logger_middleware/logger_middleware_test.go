package logger_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	// Создаем новый логгер для тестов
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel) // Отключаем вывод логов во время тестов

	// Инициализируем Gin в тестовом режиме
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggerMiddleware(log))

	// Добавляем тестовый маршрут для проверки установки логгера в контекст
	r.GET("/test-context", func(c *gin.Context) {
		logger, exists := c.Get("logger")
		assert.True(t, exists)
		assert.NotNil(t, logger)

		// Приведение типа должно быть внутри условия
		if contextLogger, ok := logger.(*logrus.Logger); ok {
			assert.Equal(t, log, contextLogger)
		} else {
			t.Error("Logger в контексте имеет неправильный тип")
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Стандартный тестовый маршрут
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Проверка основного функционала middleware
	t.Run("Логирование успешного запроса", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}
