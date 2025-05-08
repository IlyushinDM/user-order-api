package config_util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// isRunningInDocker проверяет, запущено ли приложение внутри Docker-контейнера.
func isRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// GetProjectRoot возвращает абсолютный путь к корню проекта.
// Определяет путь, исходя из расположения текущего файла в структуре проекта.
func GetProjectRoot() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("не удалось получить путь к текущему файлу")
	}

	// Поднимаемся на 3 уровня вверх от internal/utils/config_util/config_util.go
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentFile))))
	return projectRoot, nil
}

// LoadConfig загружает переменные окружения из файла .env, расположенного в корне проекта.
// Параметры:
//   - log *logrus.Logger: логгер для записи сообщений
func LoadConfig(log *logrus.Logger) {
	// Пропускаем загрузку .env, если запущено в Docker
	if isRunningInDocker() {
		log.Info("Приложение запущено в Docker - проверка файла .env пропущена")
		return
	}

	// 1. Получаем путь к корню проекта
	rootPath, err := GetProjectRoot()
	if err != nil {
		log.Warnf("Не удалось определить корень проекта: %v", err)
		return
	}

	// 2. Формируем полный путь к .env
	envPath := filepath.Join(rootPath, ".env")

	// 3. Проверяем существование файла
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Warnf("Файл .env не найден по пути: %s", envPath)
		return
	}

	// 4. Загружаем переменные
	if err := godotenv.Load(envPath); err != nil {
		log.Warnf("Ошибка при загрузке файла .env: %v", err)
	}
}
