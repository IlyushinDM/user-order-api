package logger_middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware создает Gin middleware для логирования запросов, используя предоставленный экземпляр логгера.
func LoggerMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Засекаем время начала обработки запроса
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Устанавливаем предоставленный логгер в контекст запроса.
		// Это позволяет обработчикам запросов получить доступ к тому же экземпляру логгера.
		c.Set("logger", log)

		// Обрабатываем запрос, передавая управление следующему middleware или обработчику
		c.Next()

		// Логируем детали запроса после выполнения всех обработчиков в цепочке.
		stop := time.Now()              // Время завершения обработки
		latency := stop.Sub(start)      // Вычисляем задержку (время обработки)
		clientIP := c.ClientIP()        // IP-адрес клиента
		method := c.Request.Method      // Метод HTTP запроса (GET, POST и т.д.)
		statusCode := c.Writer.Status() // Статусный код ответа
		errorMessages := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Восстанавливаем полный путь запроса с параметрами, если они были
		if raw != "" {
			path = path + "?" + raw
		}

		// Создаем log entry с дополнительными полями, специфичными для HTTP запроса.
		entry := log.WithFields(logrus.Fields{
			"statusCode": statusCode, // Статусный код ответа
			"latency":    latency,    // Задержка обработки запроса
			"clientIP":   clientIP,   // IP-адрес клиента
			"method":     method,     // Метод HTTP запроса
			"path":       path,       // Полный путь запроса
		})

		// Определяем уровень логирования на основе статусного кода и наличия ошибок.
		if len(c.Errors) > 0 {
			entry.Error(errorMessages)
		} else if statusCode >= 500 {
			entry.Error("Ошибка сервера")
		} else if statusCode >= 400 {
			entry.Warn("Ошибка клиента")
		} else {
			entry.Info("Запрос успешно выполнен")
		}
	}
}
