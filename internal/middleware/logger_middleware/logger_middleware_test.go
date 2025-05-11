package logger_middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	log_mw "github.com/IlyushinDM/user-order-api/internal/middleware/logger_middleware"

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
	r.Use(log_mw.LoggerMiddleware(log))

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

func TestLoggerMiddleware_StatusCodesAndErrors(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(log_mw.LoggerMiddleware(log))

	// 200 OK
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "ok"})
	})

	// 400 Bad Request
	r.GET("/bad", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "bad"})
	})

	// 500 Internal Server Error
	r.GET("/fail", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "fail"})
	})

	// With Gin error
	r.GET("/with-error", func(c *gin.Context) {
		c.Error(gin.Error{
			Err:  assert.AnError,
			Type: gin.ErrorTypePrivate,
		})
		c.JSON(http.StatusOK, gin.H{"msg": "error"})
	})

	tests := []struct {
		route      string
		wantStatus int
	}{
		{"/ok", http.StatusOK},
		{"/bad", http.StatusBadRequest},
		{"/fail", http.StatusInternalServerError},
		{"/with-error", http.StatusOK},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest("GET", tt.route, nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, tt.wantStatus, resp.Code, "route: %s", tt.route)
	}
}

func TestLoggerMiddleware_QueryParams(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(log_mw.LoggerMiddleware(log))

	r.GET("/query", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "ok"})
	})

	req, _ := http.NewRequest("GET", "/query?foo=bar&baz=qux", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
