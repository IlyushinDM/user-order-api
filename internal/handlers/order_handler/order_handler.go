package order_handler

import (
	"errors"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/services/order_service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderHandler struct {
	orderService order_service.OrderService
	log          *logrus.Logger
}

func NewOrderHandler(orderService order_service.OrderService, log *logrus.Logger) *OrderHandler {
	return &OrderHandler{orderService: orderService, log: log}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order for the authenticated user.
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body order_model.CreateOrderRequest true "Order data (product, quantity, price)"
// @Success 201 {object} order_model.OrderResponse "Order created successfully"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid input data"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{userID}/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req order_model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("CreateOrder: Bad request format")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	// Get user ID from JWT token (set by middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.log.Error("CreateOrder: userID not found in context (middleware error?)")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	authUserID := userID.(uint)

	order, err := h.orderService.CreateOrder(c.Request.Context(), authUserID, req)
	if err != nil {
		// Handle potential errors like user not found if service checks for it
		// if errors.Is(err, errors.New("user not found")) { ... }
		h.log.WithError(err).Errorf("CreateOrder: Failed to create order for user %d", authUserID)
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to create order"})
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
// @Summary Get order by ID
// @Description Retrieve details of a specific order by its ID. Requires authentication. User can only retrieve their own orders.
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID" Format(uint)
// @Success 200 {object} order_model.OrderResponse "Order details"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid order ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} common_handler.ErrorResponse "Order not found or access denied"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{userID}/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("GetOrderByID: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid order ID format"})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("userID")
	if !exists {
		h.log.Error("GetOrderByID: userID not found in context")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	authUserID := userID.(uint)

	order, err := h.orderService.GetOrderByID(c.Request.Context(), uint(orderID), authUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Warnf("GetOrderByID: Order %d not found for user %d", orderID, authUserID)
			// Return 404 for both not found and permission denied to avoid leaking info
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Order not found or access denied"})
		} else {
			h.log.WithError(err).Errorf("GetOrderByID: Failed to get order %d for user %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to retrieve order"})
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
// @Summary Get all orders for the authenticated user
// @Description Retrieve a paginated list of orders belonging to the currently authenticated user.
// @Tags Orders
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} order_model.PaginatedOrdersResponse "List of user's orders"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{userID}/orders [get]
func (h *OrderHandler) GetAllOrdersByUser(c *gin.Context) {
	page, limit := common_handler.GetPaginationParams(c)

	// Get user ID from JWT token
	userID, exists := c.Get("userID")
	if !exists {
		h.log.Error("GetAllOrdersByUser: userID not found in context")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	authUserID := userID.(uint)

	orders, total, err := h.orderService.GetAllOrdersByUser(c.Request.Context(), authUserID, page, limit)
	if err != nil {
		h.log.WithError(err).Errorf("GetAllOrdersByUser: Failed to retrieve orders for user %d", authUserID)
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to retrieve orders"})
		return
	}

	orderResponses := make([]order_model.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = order_model.OrderResponse{
			ID:          order.ID,
			UserID:      order.UserID,
			ProductName: order.ProductName,
			Quantity:    order.Quantity,
			Price:       order.Price,
		}
	}

	response := order_model.PaginatedOrdersResponse{
		Page:   page,
		Limit:  limit,
		Total:  total,
		Orders: orderResponses,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateOrder godoc
// @Summary Update an order
// @Description Update details of an existing order by its ID. Requires authentication. User can only update their own orders.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID" Format(uint)
// @Param order body order_model.UpdateOrderRequest true "Order data to update"
// @Success 200 {object} order_model.OrderResponse "Order updated successfully"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid input data or order ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} common_handler.ErrorResponse "Order not found or access denied"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{userID}/orders/{id} [put]
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("UpdateOrder: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid order ID format"})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("userID")
	if !exists {
		h.log.Error("UpdateOrder: userID not found in context")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	authUserID := userID.(uint)

	var req order_model.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warnf("UpdateOrder: Bad request format for order ID %d", orderID)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	order, err := h.orderService.UpdateOrder(c.Request.Context(), uint(orderID), authUserID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "permission denied or record not found" {
			h.log.Warnf("UpdateOrder: Order %d not found or not owned by user %d", orderID, authUserID)
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Order not found or access denied"})
		} else {
			h.log.WithError(err).Errorf("UpdateOrder: Failed to update order %d for user %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to update order"})
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
// @Summary Delete an order
// @Description Delete an order by its ID. Requires authentication. User can only delete their own orders.
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID" Format(uint)
// @Success 204 "Order deleted successfully"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid order ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} common_handler.ErrorResponse "Order not found or access denied"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /orders/{id} [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("DeleteOrder: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid order ID format"})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("userID")
	if !exists {
		h.log.Error("DeleteOrder: userID not found in context")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	authUserID := userID.(uint)

	err = h.orderService.DeleteOrder(c.Request.Context(), uint(orderID), authUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "permission denied or record not found" {
			h.log.Warnf("DeleteOrder: Order %d not found or not owned by user %d", orderID, authUserID)
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Order not found or access denied"})
		} else {
			h.log.WithError(err).Errorf("DeleteOrder: Failed to delete order %d for user %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to delete order"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
