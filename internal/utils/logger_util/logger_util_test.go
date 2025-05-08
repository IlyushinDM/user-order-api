package logger_util_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/utils/logger_util" // Путь к вашему пакету
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Тест для функции SetupLogger
func TestSetupLogger(t *testing.T) {
	// Сохраняем текущий stdout и восстанавливаем после теста
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Проверяем установку уровня по умолчанию (info) при отсутствии переменной окружения
	os.Unsetenv("LOG_LEVEL")
	log := logger_util.SetupLogger()
	assert.NotNil(t, log, "SetupLogger должен вернуть экземпляр логгера")
	assert.Equal(t, logrus.InfoLevel, log.GetLevel(), "Уровень логгирования должен быть Info по умолчанию")
	// Проверка формата JSON и вывода в stdout требует чтения из pipe, что сложнее.
	// Для простоты ограничимся проверкой возвращаемого типа и уровня.

	// Проверяем установку уровня из переменной окружения
	os.Setenv("LOG_LEVEL", "debug")
	log = logger_util.SetupLogger()
	assert.Equal(t, logrus.DebugLevel, log.GetLevel(), "Уровень логгирования должен быть Debug")
	os.Unsetenv("LOG_LEVEL") // Очистка переменной окружения

	// Проверяем некорректное значение переменной окружения
	os.Setenv("LOG_LEVEL", "invalid")
	log = logger_util.SetupLogger()
	assert.Equal(t, logrus.InfoLevel, log.GetLevel(), "При некорректном значении LOG_LEVEL должен использоваться уровень Info")
	os.Unsetenv("LOG_LEVEL") // Очистка переменной окружения

	w.Close() // Закрываем pipe для записи
}

// Тест для адаптера LogrusGormWriter
func TestLogrusGormWriter_Printf(t *testing.T) {
	// Создаем буфер для захвата вывода логгера
	var buf bytes.Buffer
	// Создаем мок логгер Logrus, который пишет в буфер
	mockLogger := logrus.New()
	mockLogger.SetOutput(&buf)
	mockLogger.SetFormatter(&logrus.JSONFormatter{})
	mockLogger.SetLevel(logrus.TraceLevel) // Включаем уровень Trace для проверки Printf

	// Создаем адаптер с мок логгером
	writer := &logger_util.LogrusGormWriter{Logger: mockLogger}

	// Вызываем метод Printf адаптера
	testMessage := "Тестовое сообщение с параметрами: %s %d"
	testData := []interface{}{"строка", 123}
	writer.Printf(testMessage, testData...)

	// Декодируем JSON-лог из буфера
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err, "Вывод логгера должен быть валидным JSON")

	// Проверяем, что сообщение лога соответствует ожидаемому формату Trace
	// Logrus.Tracef форматирует сообщение перед записью
	expectedMessage := "Тестовое сообщение с параметрами: строка 123"
	assert.Equal(t, "trace", logEntry["level"], "Уровень лога должен быть trace")
	assert.Equal(t, expectedMessage, logEntry["msg"], "Сообщение лога должно быть правильно отформатировано")
}
