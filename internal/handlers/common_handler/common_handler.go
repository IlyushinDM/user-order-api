package common_handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorResponse определяет стандартную структуру ответа с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// CommonHandlerInterface определяет интерфейс для общих вспомогательных функций,
// используемых другими модулями, например, для пагинации и фильтрации.
type CommonHandlerInterface interface {
	GetPaginationParams(c *gin.Context) (page, limit int, err error)
	GetFilteringParams(c *gin.Context) (map[string]any, error)
}

// CommonHandler структура для общих вспомогательных функций, управляющая зависимостями
type CommonHandler struct {
	log *logrus.Logger
}

// NewCommonHandler создает новый экземпляр CommonHandler
func NewCommonHandler(log *logrus.Logger) *CommonHandler {
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Экземпляр регистратора Logrus равен nil в NewCommonHandler, используется регистратор по умолчанию")
		log = defaultLog
	}
	return &CommonHandler{log: log}
}

// GetPaginationParams извлекает параметры пагинации из запроса.
func (h *CommonHandler) GetPaginationParams(c *gin.Context) (page, limit int, err error) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, parseErr := strconv.Atoi(pageStr)
	if parseErr != nil || page < 1 {
		h.log.WithContext(c.Request.Context()).Warnf(
			"Неверный формат параметра страницы: %s, значение по умолчанию равно 1", pageStr)
		return 0, 0, errors.New(
			"недопустимый параметр страницы: должен быть положительным целым числом")
	}

	limit, parseErr = strconv.Atoi(limitStr)
	if parseErr != nil || limit < 1 {
		h.log.WithContext(c.Request.Context()).Warnf(
			"Недопустимый формат параметра ограничения: %s, значение по умолчанию равно 10", limitStr)
		return 0, 0, errors.New(
			"недопустимый параметр limit: должен быть положительным целым числом")
	}

	// Ограничение максимального значения limit для защиты от перегрузки
	if limit > 100 {
		h.log.WithContext(c.Request.Context()).Warnf(
			"Слишком высокий предельный параметр: %d, значение не более 100", limit)
		limit = 100
	}
	return page, limit, nil
}

// GetFilteringParams извлекает параметры фильтрации из запроса.
func (h *CommonHandler) GetFilteringParams(c *gin.Context) (map[string]any, error) {
	filters := make(map[string]any)
	log := h.log.WithContext(c.Request.Context())

	// Обработка минимального возраста
	if minAgeStr := c.Query("min_age"); minAgeStr != "" {
		minAge, err := strconv.Atoi(minAgeStr)
		if err == nil && minAge > 0 {
			filters["min_age"] = minAge
		} else {
			log.Warnf("Некорректный параметр min_age: %s", minAgeStr)
			return nil, errors.New(
				"недопустимый параметр min_age: должен быть положительным целым числом")
		}
	}

	// Обработка максимального возраста
	if maxAgeStr := c.Query("max_age"); maxAgeStr != "" {
		maxAge, err := strconv.Atoi(maxAgeStr)
		if err == nil && maxAge > 0 {
			filters["max_age"] = maxAge
		} else {
			log.Warnf("Некорректный параметр max_age: %s", maxAgeStr)
			return nil, errors.New(
				"недопустимый параметр max_age: должен быть положительным целым числом")
		}
	}

	// Обработка фильтра по имени
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	return filters, nil
}
