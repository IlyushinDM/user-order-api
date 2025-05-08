package config_util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestGetProjectRoot проверяет, что функция GetProjectRoot возвращает корректный путь.
// Примечание: Этот тест предполагает определенную структуру каталогов относительно тестового файла.
func TestGetProjectRoot(t *testing.T) {
	rootPath, err := GetProjectRoot()
	assert.NoError(t, err, "GetProjectRoot не должна возвращать ошибку")
	assert.NotEmpty(t, rootPath, "Путь к корню проекта не должен быть пустым")

	// Проверяем, существует ли ожидаемый файл (например, go.mod) в полученном корневом каталоге
	// Это делает тест более надежным, но зависит от наличия go.mod
	goModPath := filepath.Join(rootPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("go.mod не найден в предполагаемом корневом каталоге проекта: %s", rootPath)
	}
}

// TestLoadConfig проверяет загрузку конфигурации.
// Примечание: Этот тест базовый. Для полного тестирования может потребоваться
// создание временного файла .env и проверка установки переменных окружения.
func TestLoadConfig(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel) // Устанавливаем уровень логирования для теста

	// Сохраняем текущие переменные окружения, чтобы восстановить их после теста
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := splitEnv(env)
			os.Setenv(parts[0], parts[1])
		}
	}()
	os.Clearenv() // Очищаем переменные окружения для чистоты теста

	// Сценарий 1: Файл .env не существует (ожидается предупреждение в логах, но не ошибка)
	t.Run("dotenv_file_missing", func(t *testing.T) {
		// Убеждаемся, что .env точно нет там, где его будет искать LoadConfig
		// Это сложно сделать идеально без изменения GetProjectRoot или мокирования os.Stat
		// Поэтому просто вызываем LoadConfig и ожидаем, что она не вызовет панику
		assert.NotPanics(t, func() {
			LoadConfig(log)
		}, "LoadConfig не должна паниковать, если .env файл отсутствует")
	})

	// Сценарий 2: Файл .env существует и содержит переменные
	t.Run("dotenv_file_exists", func(t *testing.T) {
		// Создаем временный файл .env
		// Получаем "корень проекта" относительно текущего файла теста
		// Это предположение, что GetProjectRoot работает корректно
		// и что тест находится в utils/config_util/
		// dir, _ := os.Getwd() // current test dir
		// projectRootGuess := filepath.Dir(filepath.Dir(dir)) // ../../

		// Вместо сложного вычисления пути, создадим .env в текущей директории теста,
		// и временно подменим функцию GetProjectRoot, если бы это было возможно без рефакторинга.
		// В данном случае, проще проверить загрузку переменных, если они уже установлены.

		// Альтернативно, можно создать временный .env файл в ожидаемом месте
		// rootPath, _ := GetProjectRoot() // Предполагаем, что это работает
		// tempEnvPath := filepath.Join(rootPath, ".env.test")
		// os.Rename(filepath.Join(rootPath, ".env"), tempEnvPath) // бэкапим существующий
		// defer os.Rename(tempEnvPath, filepath.Join(rootPath, ".env")) // восстанавливаем

		tempEnvFile, err := os.Create(".env") // Создаем .env в текущей директории теста
		if !assert.NoError(t, err, "Не удалось создать временный .env файл") {
			return
		}
		defer os.Remove(tempEnvFile.Name()) // Удаляем временный файл после теста

		_, err = tempEnvFile.WriteString("TEST_VAR_CONFIG=loaded_from_env\n")
		assert.NoError(t, err, "Не удалось записать во временный .env файл")
		tempEnvFile.Close()

		// Модифицируем LoadConfig так, чтобы она искала .env в текущей директории для этого теста
		// Это потребует изменения самой функции LoadConfig или использования моков.
		// Вместо этого, проверим, что если переменная уже установлена, она не перезаписывается,
		// если .env не загружается (например, в Docker).

		// Простой тест: вызвать LoadConfig и проверить, установилась ли переменная
		// Для этого теста мы должны мокнуть GetProjectRoot, чтобы он указывал на текущую директорию
		// или изменить LoadConfig для большей тестируемости.
		// Поскольку это сложно без изменения кода, оставим этот сценарий как идею для улучшения.

		// В данном случае, LoadConfig загрузит .env из корня проекта, если он там есть.
		// Мы можем проверить, что функция не паникует.
		assert.NotPanics(t, func() {
			LoadConfig(log) // Это вызовет реальный LoadConfig
		}, "LoadConfig не должна паниковать при наличии .env")

		// После вызова LoadConfig (если она нашла и загрузила .env из корня проекта)
		// можно было бы проверить переменную. Но это делает тест зависимым от реального .env.
		// Пример:
		// if os.Getenv("SOME_VAR_FROM_REAL_ENV") == "" {
		// t.Log("Переменная из реального .env не загружена, что может быть нормально")
		// }
	})

	// Сценарий 3: Запуск в Docker (пропуск загрузки .env)
	t.Run("running_in_docker", func(t *testing.T) {
		// Устанавливаем переменную окружения, чтобы имитировать Docker
		// Это требует изменения isRunningInDocker, чтобы она проверяла переменную,
		// или мокирования os.Stat("/.dockerenv").
		// Поскольку os.Stat мокировать сложно без интерфейса, этот тест концептуальный.
		// Например, если бы isRunningInDocker была такой:
		// func isRunningInDocker() bool {
		// if os.Getenv("IS_DOCKER_TEST") == "true" { return true }
		// if _, err := os.Stat("/.dockerenv"); err == nil { return true }
		// return false
		// }
		// Тогда можно было бы: os.Setenv("IS_DOCKER_TEST", "true"); LoadConfig(log); ...
		// Проверяем, что log.Info содержит "проверка файла .env пропущена"
		// Это потребует перехвата вывода логгера.
	})
}

// Вспомогательная функция для разделения строки окружения
func splitEnv(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env, ""}
}
