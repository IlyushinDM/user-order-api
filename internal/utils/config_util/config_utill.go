package config_util

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// loadConfig загружает переменные окружения из .env файла.
// Параметры:
//   - log *logrus.Logger: логгер для записи сообщений
func LoadConfig(log *logrus.Logger) {
	err := godotenv.Load() // Загружает .env файл из текущей директории
	if err != nil {
		log.Warn("Ошибка загрузки .env файла, используются системные переменные окружения")
		// Не завершаем работу при отсутствии .env, возможно используются системные переменные
	}
}
