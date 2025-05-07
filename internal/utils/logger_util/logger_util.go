package logger_util

import (
	"os"

	"github.com/sirupsen/logrus"
)

// setupLogger настраивает логгер Logrus.
// Возвращает:
//   - *logrus.Logger: экземпляр логгера
func SetupLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{}) // Формат JSON для структурированного логгирования
	log.SetOutput(os.Stdout)                  // Логи в stdout

	levelStr := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		log.Warnf("Некорректный LOG_LEVEL '%s', используется уровень 'info'", levelStr)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	return log
}

// logrusGormWriter адаптирует логгер Logrus к интерфейсу логгера GORM
type LogrusGormWriter struct {
	Logger *logrus.Logger
}

// Printf реализует интерфейс логгера GORM
func (w *LogrusGormWriter) Printf(message string, data ...interface{}) {
	w.Logger.Tracef(message, data...)
}
