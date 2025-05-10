package common_handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert" // Рекомендуется использовать testify для удобных утверждений

	// Импортируем тестируемый пакет
	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	// Замените "your_module_path" на фактический путь к вашему модулю.
)

// setupTestGinContext создает mock Gin контекст для тестирования
// Принимает URL строку для создания http.Request
func setupTestGinContext(w *httptest.ResponseRecorder, reqURL string) (*gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode) // Устанавливаем тестовый режим для Gin
	c, router := gin.CreateTestContext(w)

	// Создаем полный объект http.Request с нужным URL
	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)
	c.Request = req

	return c, router
}

// TestNewCommonHandler тестирует функцию создания экземпляра CommonHandler
func TestNewCommonHandler(t *testing.T) {
	// Тестовый случай: логгер предоставлен
	t.Run("С логгером", func(t *testing.T) {
		mockLogger := logrus.New()
		handler := common_handler.NewCommonHandler(mockLogger)

		// Проверяем, что handler не nil
		assert.NotNil(t, handler, "NewCommonHandler не должен возвращать nil при предоставлении логгера")
		// Проверяем, что используемый логгер соответствует предоставленному
		// Прямой доступ к приватному полю затруднен, но можно проверить косвенно или не проверять совсем, если зависимость введена правильно
	})

	// Тестовый случай: логгер nil
	t.Run("Логгер nil", func(t *testing.T) {
		// В этом случае должен быть создан логгер по умолчанию внутри функции
		handler := common_handler.NewCommonHandler(nil)

		// Проверяем, что handler не nil
		assert.NotNil(t, handler, "NewCommonHandler не должен возвращать nil при логгере nil")
		// Проверяем, что создан логгер (косвенно)
		// Прямая проверка типа или значения логгера по умолчанию затруднена извне пакета
	})
}

// TestGetPaginationParams тестирует извлечение параметров пагинации
func TestGetPaginationParams(t *testing.T) {
	// Создаем экземпляр CommonHandler для тестов
	mockLogger := logrus.New()
	handler := common_handler.NewCommonHandler(mockLogger)

	// Тестовый случай: параметры по умолчанию
	t.Run("Параметры по умолчанию", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем корневой URL для параметров по умолчанию
		c, _ := setupTestGinContext(w, "/")

		page, limit, err := handler.GetPaginationParams(c)

		// Проверяем отсутствие ошибки и ожидаемые значения по умолчанию
		assert.NoError(t, err, "Не должно быть ошибки для параметров по умолчанию")
		assert.Equal(t, 1, page, "Значение страницы по умолчанию должно быть 1")
		assert.Equal(t, 10, limit, "Значение лимита по умолчанию должно быть 10")
	})

	// Тестовый случай: допустимые пользовательские параметры
	t.Run("Допустимые пользовательские параметры", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с допустимыми параметрами
		c, _ := setupTestGinContext(w, "/?page=5&limit=25")

		page, limit, err := handler.GetPaginationParams(c)

		// Проверяем отсутствие ошибки и ожидаемые значения
		assert.NoError(t, err, "Не должно быть ошибки для допустимых параметров")
		assert.Equal(t, 5, page, "Должно быть прочитано пользовательское значение страницы")
		assert.Equal(t, 25, limit, "Должно быть прочитано пользовательское значение лимита")
	})

	// Тестовый случай: недопустимый формат параметра страницы
	t.Run("Недопустимый формат страницы", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым форматом страницы
		c, _ := setupTestGinContext(w, "/?page=abc&limit=10")

		page, limit, err := handler.GetPaginationParams(c)

		// Проверяем наличие ошибки и что страница/лимит равны 0 при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого формата страницы")
		assert.Contains(t, err.Error(), "недопустимый параметр страницы", "Сообщение об ошибке должно указывать на параметр страницы")
		assert.Equal(t, 0, page, "Страница должна быть 0 при ошибке")
		assert.Equal(t, 0, limit, "Лимит должен быть 0 при ошибке")
	})

	// Тестовый случай: недопустимое значение параметра страницы (меньше 1)
	t.Run("Недопустимое значение страницы (меньше 1)", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым значением страницы
		c, _ := setupTestGinContext(w, "/?page=0&limit=10")

		page, limit, err := handler.GetPaginationParams(c)

		// Проверяем наличие ошибки и что страница/лимит равны 0 при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого значения страницы")
		assert.Contains(t, err.Error(), "недопустимый параметр страницы", "Сообщение об ошибке должно указывать на параметр страницы")
		assert.Equal(t, 0, page, "Страница должна быть 0 при ошибке")
		assert.Equal(t, 0, limit, "Лимит должен быть 0 при ошибке")
	})

	// Тестовый случай: недопустимый формат параметра лимита
	t.Run("Недопустимый формат лимита", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым форматом лимита
		c, _ := setupTestGinContext(w, "/?page=1&limit=xyz")

		page, limit, err := handler.GetPaginationParams(c)

		// Проверяем наличие ошибки и что страница/лимит равны 0 при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого формата лимита")
		assert.Contains(t, err.Error(), "недопустимый параметр limit", "Сообщение об ошибке должно указывать на параметр лимита")
		assert.Equal(t, 0, page, "Страница должна быть 0 при ошибке")
		assert.Equal(t, 0, limit, "Лимит должен быть 0 при ошибке")
	})

	// Тестовый случай: недопустимое значение параметра лимита (меньше 1)
	t.Run("Недопустимое значение лимита (меньше 1)", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым значением лимита
		c, _ := setupTestGinContext(w, "/?page=1&limit=0")

		page, limit, err := handler.GetPaginationParams(c)

		// Проверяем наличие ошибки и что страница/лимит равны 0 при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого значения лимита")
		assert.Contains(t, err.Error(), "недопустимый параметр limit", "Сообщение об ошибке должно указывать на параметр лимита")
		assert.Equal(t, 0, page, "Страница должна быть 0 при ошибке")
		assert.Equal(t, 0, limit, "Лимит должен быть 0 при ошибке")
	})

	// Тестовый случай: значение лимита превышает максимальное (100)
	t.Run("Лимит превышает максимальное", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с лимитом больше 100
		c, _ := setupTestGinContext(w, "/?page=1&limit=150")

		page, limit, err := handler.GetPaginationParams(c)

		// Проверяем отсутствие ошибки и что лимит ограничен 100
		assert.NoError(t, err, "Не должно быть ошибки при лимите больше 100, он должен быть ограничен")
		assert.Equal(t, 1, page, "Значение страницы должно быть корректным")
		assert.Equal(t, 100, limit, "Лимит должен быть ограничен значением 100")
	})
}

// TestGetFilteringParams тестирует извлечение параметров фильтрации
func TestGetFilteringParams(t *testing.T) {
	// Создаем экземпляр CommonHandler для тестов
	mockLogger := logrus.New()
	handler := common_handler.NewCommonHandler(mockLogger)

	// Тестовый случай: нет параметров фильтрации
	t.Run("Нет параметров фильтрации", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем корневой URL без параметров
		c, _ := setupTestGinContext(w, "/")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем отсутствие ошибки и пустую карту фильтров
		assert.NoError(t, err, "Не должно быть ошибки без параметров фильтрации")
		assert.NotNil(t, filters, "Карта фильтров должна быть инициализирована")
		assert.Empty(t, filters, "Карта фильтров должна быть пустой")
	})

	// Тестовый случай: допустимый параметр min_age
	t.Run("Допустимый min_age", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с допустимым min_age
		c, _ := setupTestGinContext(w, "/?min_age=18")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем отсутствие ошибки и наличие min_age в фильтрах
		assert.NoError(t, err, "Не должно быть ошибки для допустимого min_age")
		assert.NotNil(t, filters, "Карта фильтров должна быть инициализирована")
		assert.Contains(t, filters, "min_age", "Должен быть добавлен фильтр min_age")
		assert.Equal(t, 18, filters["min_age"], "Значение min_age должно быть корректным")
	})

	// Тестовый случай: недопустимый формат параметра min_age
	t.Run("Недопустимый формат min_age", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым форматом min_age
		c, _ := setupTestGinContext(w, "/?min_age=abc")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем наличие ошибки и что фильтры nil при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого формата min_age")
		assert.Contains(t, err.Error(), "недопустимый параметр min_age", "Сообщение об ошибке должно указывать на параметр min_age")
		assert.Nil(t, filters, "Фильтры должны быть nil при ошибке парсинга min_age")
	})

	// Тестовый случай: недопустимое значение параметра min_age (меньше или равно 0)
	t.Run("Недопустимое значение min_age (<= 0)", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым значением min_age
		c, _ := setupTestGinContext(w, "/?min_age=0")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем наличие ошибки и что фильтры nil при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого значения min_age")
		assert.Contains(t, err.Error(), "недопустимый параметр min_age", "Сообщение об ошибке должно указывать на параметр min_age")
		assert.Nil(t, filters, "Фильтры должны быть nil при ошибке валидации min_age")
	})

	// Тестовый случай: допустимый параметр max_age
	t.Run("Допустимый max_age", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с допустимым max_age
		c, _ := setupTestGinContext(w, "/?max_age=65")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем отсутствие ошибки и наличие max_age в фильтрах
		assert.NoError(t, err, "Не должно быть ошибки для допустимого max_age")
		assert.NotNil(t, filters, "Карта фильтров должна быть инициализирована")
		assert.Contains(t, filters, "max_age", "Должен быть добавлен фильтр max_age")
		assert.Equal(t, 65, filters["max_age"], "Значение max_age должно быть корректным")
	})

	// Тестовый случай: недопустимый формат параметра max_age
	t.Run("Недопустимый формат max_age", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым форматом max_age
		c, _ := setupTestGinContext(w, "/?max_age=xyz")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем наличие ошибки и что фильтры nil при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого формата max_age")
		assert.Contains(t, err.Error(), "недопустимый параметр max_age", "Сообщение об ошибке должно указывать на параметр max_age")
		assert.Nil(t, filters, "Фильтры должны быть nil при ошибке парсинга max_age")
	})

	// Тестовый случай: недопустимое значение параметра max_age (меньше или равно 0)
	t.Run("Недопустимое значение max_age (<= 0)", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с недопустимым значением max_age
		c, _ := setupTestGinContext(w, "/?max_age=0")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем наличие ошибки и что фильтры nil при ошибке
		assert.Error(t, err, "Должна быть ошибка для недопустимого значения max_age")
		assert.Contains(t, err.Error(), "недопустимый параметр max_age", "Сообщение об ошибке должно указывать на параметр max_age")
		assert.Nil(t, filters, "Фильтры должны быть nil при ошибке валидации max_age")
	})

	// Тестовый случай: допустимый параметр name
	t.Run("Допустимый name", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с допустимым name
		c, _ := setupTestGinContext(w, "/?name=John+Doe") // Используем + для пробела в URL

		filters, err := handler.GetFilteringParams(c)

		// Проверяем отсутствие ошибки и наличие name в фильтрах
		assert.NoError(t, err, "Не должно быть ошибки для допустимого name")
		assert.NotNil(t, filters, "Карта фильтров должна быть инициализирована")
		assert.Contains(t, filters, "name", "Должен быть добавлен фильтр name")
		assert.Equal(t, "John Doe", filters["name"], "Значение name должно быть корректным")
	})

	// Тестовый случай: несколько допустимых параметров фильтрации
	t.Run("Несколько допустимых параметров", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с несколькими допустимыми параметрами
		c, _ := setupTestGinContext(w, "/?min_age=20&max_age=50&name=Jane")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем отсутствие ошибки и наличие всех параметров в фильтрах
		assert.NoError(t, err, "Не должно быть ошибки для нескольких допустимых параметров")
		assert.NotNil(t, filters, "Карта фильтров должна быть инициализирована")
		assert.Len(t, filters, 3, "Должно быть 3 фильтра")
		assert.Contains(t, filters, "min_age", "Должен быть добавлен фильтр min_age")
		assert.Equal(t, 20, filters["min_age"], "Значение min_age должно быть корректным")
		assert.Contains(t, filters, "max_age", "Должен быть добавлен фильтр max_age")
		assert.Equal(t, 50, filters["max_age"], "Значение max_age должно быть корректным")
		assert.Contains(t, filters, "name", "Должен быть добавлен фильтр name")
		assert.Equal(t, "Jane", filters["name"], "Значение name должно быть корректным")
	})

	// Тестовый случай: комбинация допустимых и недопустимых параметров фильтрации (min_age некорректен)
	t.Run("Комбинация допустимых и недопустимых параметров (min_age некорректен)", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с некорректным min_age
		c, _ := setupTestGinContext(w, "/?min_age=abc&max_age=50&name=Jane")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем наличие ошибки и что фильтры nil при ошибке парсинга
		assert.Error(t, err, "Должна быть ошибка из-за некорректного min_age")
		assert.Contains(t, err.Error(), "недопустимый параметр min_age", "Сообщение об ошибке должно указывать на параметр min_age")
		assert.Nil(t, filters, "Фильтры должны быть nil при ошибке парсинга min_age")
	})

	// Тестовый случай: комбинация допустимых и недопустимых параметров фильтрации (max_age некорректен)
	t.Run("Комбинация допустимых и недопустимых параметров (max_age некорректен)", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Передаем URL с некорректным max_age
		c, _ := setupTestGinContext(w, "/?min_age=20&max_age=xyz&name=Jane")

		filters, err := handler.GetFilteringParams(c)

		// Проверяем наличие ошибки и что фильтры nil при ошибке парсинга
		assert.Error(t, err, "Должна быть ошибка из-за некорректного max_age")
		assert.Contains(t, err.Error(), "недопустимый параметр max_age", "Сообщение об ошибке должно указывать на параметр max_age")
		assert.Nil(t, filters, "Фильтры должны быть nil при ошибке парсинга max_age")
	})
}
