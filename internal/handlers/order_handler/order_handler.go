package order_handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/services/order_service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderHandler struct {
	orderService order_service.OrderService
	// Изменено: использование интерфейса вместо указателя на конкретную структуру
	commonHandler common_handler.CommonHandlerInterface
	log           *logrus.Logger
}

// NewOrderHandler создает новый экземпляр OrderHandler с проверкой входных параметров
func NewOrderHandler(
	orderService order_service.OrderService,
	// Изменено: использование интерфейса вместо указателя на конкретную структуру
	commonHandler common_handler.CommonHandlerInterface,
	log *logrus.Logger,
) *OrderHandler {
	if orderService == nil {
		logrus.Fatal("orderService равен nil в NewOrderHandler")
	}
	if commonHandler == nil {
		logrus.Fatal("commonHandler равен nil в NewOrderHandler")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Logger равен nil в NewOrderHandler, используется logger по умолчанию")
		log = defaultLog
	}
	return &OrderHandler{orderService, commonHandler, log}
}

// checkUserIDMatch проверяет соответствие userID из URL и аутентифицированного userID из контекста
func (h *OrderHandler) checkUserIDMatch(c *gin.Context) (uint, bool) {
	authUserID, exists := c.Get("userID")
	if !exists {
		h.log.Error("Ошибка аутентификации: userID не найден в контексте")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка аутентификации"})
		return 0, false
	}

	urlUserIDStr := c.Param("id")
	urlUserID, err := strconv.ParseUint(urlUserIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("Неверный формат userID в URL: '%s'", urlUserIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректный формат ID пользователя"})
		return 0, false
	}

	if uint(urlUserID) != authUserID.(uint) {
		h.log.Warnf("Доступ запрещен: пользователь %d пытается получить доступ к данным пользователя %d", authUserID.(uint), urlUserID)
		c.JSON(http.StatusForbidden, common_handler.ErrorResponse{Error: "Доступ запрещен"})
		return 0, false
	}

	return authUserID.(uint), true
}

// CreateOrder godoc
// @Summary Создание нового заказа
// @Description Создает новый заказ для аутентифицированного пользователя
// @Tags Заказы
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Param order body order_model.CreateOrderRequest true "Данные заказа"
// @Success 201 {object} order_model.OrderResponse "Заказ успешно создан"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректные входные данные"
// @Failure 401 {object} common_handler.ErrorResponse "Не авторизован"
// @Failure 403 {object} common_handler.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id}/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return
	}

	var req order_model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Некорректный формат запроса")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректные входные данные"})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), authUserID, req)
	if err != nil {
		h.log.WithError(err).Errorf("Ошибка при создании заказа для пользователя %d", authUserID)
		switch {
		case errors.Is(err, order_service.ErrInvalidServiceInput):
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: err.Error()})
		case errors.Is(err, order_service.ErrServiceDatabaseError):
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка при создании заказа"})
		default:
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusCreated, order_model.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		ProductName: order.ProductName,
		Quantity:    order.Quantity,
		Price:       order.Price,
	})
}

// GetOrderByID godoc
// @Summary Получение заказа по ID
// @Description Возвращает информацию о конкретном заказе пользователя
// @Tags Заказы
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Param orderID path int true "ID заказа" Format(uint)
// @Success 200 {object} order_model.OrderResponse "Информация о заказе"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректный формат ID"
// @Failure 401 {object} common_handler.ErrorResponse "Не авторизован"
// @Failure 403 {object} common_handler.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} common_handler.ErrorResponse "Заказ не найден"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id}/orders/{orderID} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return
	}

	orderIDStr := c.Param("orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("Некорректный формат orderID: '%s'", orderIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректный формат ID заказа"})
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), uint(orderID), authUserID)
	if err != nil {
		switch {
		case errors.Is(err, order_service.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Заказ не найден"})
		case errors.Is(err, order_service.ErrInvalidServiceInput):
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректный запрос"})
		case errors.Is(err, order_service.ErrServiceDatabaseError):
			h.log.WithError(err).Errorf("Ошибка БД при получении заказа %d для пользователя %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка при получении заказа"})
		default:
			h.log.WithError(err).Errorf("Ошибка при получении заказа %d для пользователя %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, order_model.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		ProductName: order.ProductName,
		Quantity:    order.Quantity,
		Price:       order.Price,
	})
}

// GetAllOrdersByUser godoc
// @Summary Получение всех заказов пользователя
// @Description Возвращает список заказов пользователя с пагинацией
// @Tags Заказы
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(10) minimum(1) maximum(100)
// @Success 200 {object} order_model.PaginatedOrdersResponse "Список заказов"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректные параметры"
// @Failure 401 {object} common_handler.ErrorResponse "Не авторизован"
// @Failure 403 {object} common_handler.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id}/orders [get]
func (h *OrderHandler) GetAllOrdersByUser(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return
	}

	page, limit, err := h.commonHandler.GetPaginationParams(c)
	if err != nil {
		h.log.WithError(err).Warnf("Некорректные параметры пагинации для пользователя %d", authUserID)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Неверные параметры пагинации"})
		return
	}

	orders, total, err := h.orderService.GetAllOrdersByUser(c.Request.Context(), authUserID, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, order_service.ErrInvalidServiceInput):
			h.log.WithError(err).Warnf("Ошибка валидации сервиса для пользователя %d", authUserID)
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Неверные параметры запроса"})
		case errors.Is(err, order_service.ErrServiceDatabaseError):
			h.log.WithError(err).Errorf("Ошибка БД при получении заказов пользователя %d", authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{
				Error: "Ошибка при получении списка заказов",
			})
		default:
			h.log.WithError(err).Errorf("Ошибка при получении заказов пользователя %d", authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Внутренняя ошибка сервера"})
		}
		return
	}

	response := order_model.PaginatedOrdersResponse{
		Page:   page,
		Limit:  limit,
		Total:  total,
		Orders: make([]order_model.OrderResponse, len(orders)),
	}

	for i, order := range orders {
		response.Orders[i] = order_model.OrderResponse{
			ID:          order.ID,
			UserID:      order.UserID,
			ProductName: order.ProductName,
			Quantity:    order.Quantity,
			Price:       order.Price,
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateOrder godoc
// @Summary Обновление заказа
// @Description Обновляет информацию о заказе пользователя
// @Tags Заказы
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Param orderID path int true "ID заказа" Format(uint)
// @Param order body order_model.UpdateOrderRequest true "Данные для обновления"
// @Success 200 {object} order_model.OrderResponse "Обновленный заказ"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректные данные"
// @Failure 401 {object} common_handler.ErrorResponse "Не авторизован"
// @Failure 403 {object} common_handler.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} common_handler.ErrorResponse "Заказ не найден"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id}/orders/{orderID} [put]
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return
	}

	orderIDStr := c.Param("orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("Некорректный формат orderID: '%s'", orderIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректный формат ID заказа"})
		return
	}

	var req order_model.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warnf("Некорректный формат запроса для заказа %d", orderID)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректные данные для обновления"})
		return
	}

	order, err := h.orderService.UpdateOrder(c.Request.Context(), uint(orderID), authUserID, req)
	if err != nil {
		switch {
		case errors.Is(err, order_service.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Заказ не найден"})
		case errors.Is(err, order_service.ErrNoUpdateFields):
			h.log.WithField("order_id", orderID).Info("Получен запрос на обновление без изменений")
			// Вернуть существующий заказ, если нет изменений
			existingOrder, getErr := h.orderService.GetOrderByID(c.Request.Context(), uint(orderID), authUserID)
			if getErr != nil {
				h.log.WithError(getErr).Errorf("Ошибка при получении существующего заказа %d после ErrNoUpdateFields", orderID)
				c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка при получении заказа"})
				return
			}
			c.JSON(http.StatusOK, order_model.OrderResponse{
				ID:          existingOrder.ID,
				UserID:      existingOrder.UserID,
				ProductName: existingOrder.ProductName,
				Quantity:    existingOrder.Quantity,
				Price:       existingOrder.Price,
			})
			return
		case errors.Is(err, order_service.ErrInvalidServiceInput):
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: err.Error()})
		case errors.Is(err, order_service.ErrServiceDatabaseError):
			h.log.WithError(err).Errorf("Ошибка БД при обновлении заказа %d для пользователя %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка при обновлении заказа"})
		default:
			h.log.WithError(err).Errorf("Ошибка при обновлении заказа %d для пользователя %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, order_model.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		ProductName: order.ProductName,
		Quantity:    order.Quantity,
		Price:       order.Price,
	})
}

// DeleteOrder godoc
// @Summary Удаление заказа
// @Description Удаляет заказ пользователя по ID
// @Tags Заказы
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Param orderID path int true "ID заказа" Format(uint)
// @Success 204 "Заказ удален"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректный формат ID"
// @Failure 401 {object} common_handler.ErrorResponse "Не авторизован"
// @Failure 403 {object} common_handler.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} common_handler.ErrorResponse "Заказ не найден"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id}/orders/{orderID} [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return
	}

	orderIDStr := c.Param("orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("Некорректный формат orderID: '%s'", orderIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректный формат ID заказа"})
		return
	}

	if err := h.orderService.DeleteOrder(c.Request.Context(), uint(orderID), authUserID); err != nil {
		switch {
		case errors.Is(err, order_service.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Заказ не найден"})
		case errors.Is(err, order_service.ErrInvalidServiceInput):
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Некорректные данные запроса"})
		case errors.Is(err, order_service.ErrServiceDatabaseError):
			h.log.WithError(err).Errorf("Ошибка БД при удалении заказа %d для пользователя %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка при удалении заказа"})
		default:
			h.log.WithError(err).Errorf("Ошибка при удалении заказа %d для пользователя %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Внутренняя ошибка сервера"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
