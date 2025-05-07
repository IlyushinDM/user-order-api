package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	// Setup
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel) // Ensure debug level to see all logs

	// Create a buffer to capture log output
	//var logBuffer bytes.Buffer
	//log.SetOutput(&logBuffer)

	// Gin setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LoggerMiddleware(log))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/test_error", func(c *gin.Context) {
		c.Error(errors.New("test error")).SetType(gin.ErrorTypePrivate)
		c.String(http.StatusInternalServerError, "Error")
	})

	t.Run("Successful Request Logging", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Check log output (this might require more sophisticated parsing)
		// Example: assert.Contains(t, logBuffer.String(), "Request completed")
	})

	t.Run("Request Logging with Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test_error", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// Check log output for error message
		// Example: assert.Contains(t, logBuffer.String(), "test error")
	})

	t.Run("Logger instance in context", func(t *testing.T) {
		router.GET("/context_test", func(c *gin.Context) {
			loggerInterface, exists := c.Get("logger")
			assert.True(t, exists, "Logger should exist in context")

			logger, ok := loggerInterface.(*logrus.Logger)
			assert.True(t, ok, "Logger in context should be of type *logrus.Logger")

			assert.NotNil(t, logger, "Logger from context should not be nil")
			c.String(http.StatusOK, "OK")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/context_test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Request with query parameters", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test?param1=value1&param2=value2", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		//Add assertions to check log output
		//assert.Contains(t, logBuffer.String(), "/test?param1=value1&param2=value2")
	})
}
