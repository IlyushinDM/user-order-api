package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware creates a Gin middleware for request logging using Logrus.
func LoggerMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Set logger instance in context for handlers to use
		c.Set("logger", log)

		// Process request
		c.Next()

		// Log request details after handler execution
		stop := time.Now()
		latency := stop.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		entry := log.WithFields(logrus.Fields{
			"statusCode": statusCode,
			"latency":    latency,
			"clientIP":   clientIP,
			"method":     method,
			"path":       path,
		})

		if errorMessage != "" {
			entry.Error(errorMessage)
		} else {
			entry.Info("Request completed")
		}
	}
}
