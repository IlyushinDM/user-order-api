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

// checkUserIDMatch compares the userID in the URL path with the authenticated userID from the context.
// Returns the authenticated userID and true if they match or context is missing, otherwise returns 0 and false.
func (h *OrderHandler) checkUserIDMatch(c *gin.Context) (uint, bool) {
	authUserID, exists := c.Get("userID")
	if !exists {
		h.log.Error("Authentication context error: userID not found")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{
			Error: "Authentication context error",
		})
		return 0, false // Authentication context missing
	}

	// Use "id" instead of "userID" to match the route parameter name in main.go
	urlUserIDStr := c.Param("id")
	urlUserID, err := strconv.ParseUint(urlUserIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("Invalid userID format in URL: '%s'", urlUserIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{
			Error: "Invalid user ID format in URL",
		})
		return 0, false // Invalid URL user ID format
	}

	if uint(urlUserID) != authUserID.(uint) {
		h.log.Warnf("Forbidden access attempt: Authenticated user %d trying to access user %d's resources", authUserID.(uint), urlUserID)
		c.JSON(http.StatusForbidden, common_handler.ErrorResponse{
			Error: "Forbidden: You can only access your own orders",
		})
		return 0, false // User ID mismatch
	}

	return authUserID.(uint), true // Match
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order for the authenticated user. The {id} in the path is validated against the authenticated user ID from the token.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "User ID" Format(uint) // Changed from userID to id
// @Param order body order_model.CreateOrderRequest true "Order data (product, quantity, price)"
// @Success 201 {object} order_model.OrderResponse "Order created successfully"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid input data or user ID format in URL"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} common_handler.ErrorResponse "Forbidden (trying to create order for another user)"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id}/orders [post] // Changed from /api/users/{userID}/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return // Response handled by checkUserIDMatch
	}

	var req order_model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("CreateOrder: Bad request format")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{
			Error:   "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	// Use authUserID obtained from the context/token, validated against URL userID
	order, err := h.orderService.CreateOrder(c.Request.Context(), authUserID, req)
	if err != nil {
		h.log.WithError(err).Errorf("CreateOrder: Failed for user %d", authUserID)
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{
			Error: "Failed to create order",
		})
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
// @Description Retrieve details of a specific order by its ID for a specific user. The {id} in the path is validated against the authenticated user ID from the token.
// @Tags Orders
// @Produce json
// @Param id path int true "User ID" Format(uint) // Changed from userID to id
// @Param orderID path int true "Order ID" Format(uint)
// @Success 200 {object} order_model.OrderResponse "Order details"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} common_handler.ErrorResponse "Forbidden (trying to access another user's order)"
// @Failure 404 {object} common_handler.ErrorResponse "Order not found or access denied"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id}/orders/{orderID} [get] // Changed from /api/users/{userID}/orders/{orderID}
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return // Response handled by checkUserIDMatch
	}

	// Read order ID using the new parameter name 'orderID'
	orderIDStr := c.Param("orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("GetOrderByID: Invalid orderID format '%s'", orderIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{
			Error: "Invalid order ID format",
		})
		return
	}

	// Pass both orderID and authUserID to the service layer for ownership check
	order, err := h.orderService.GetOrderByID(c.Request.Context(), uint(orderID), authUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// The service/repo layer returns ErrRecordNotFound if the order doesn't exist OR
			// if it exists but doesn't belong to the authenticated user.
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{
				Error: "Order not found or access denied",
			})
		} else {
			h.log.WithError(err).Errorf("GetOrderByID: Failed for order %d for user %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{
				Error: "Failed to retrieve order",
			})
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
// @Summary Get all orders for user
// @Description Retrieve paginated list of user's orders. The {id} in the path is validated against the authenticated user ID from the token.
// @Tags Orders
// @Produce json
// @Param id path int true "User ID" Format(uint) // Changed from userID to id
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} order_model.PaginatedOrdersResponse "List of orders"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid query parameters or user ID format in URL"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} common_handler.ErrorResponse "Forbidden (trying to access another user's orders)"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id}/orders [get] // Changed from /api/users/{userID}/orders
func (h *OrderHandler) GetAllOrdersByUser(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return // Response handled by checkUserIDMatch
	}

	page, limit := common_handler.GetPaginationParams(c)
	// Pass authUserID to service to get orders only for the authenticated user
	orders, total, err := h.orderService.GetAllOrdersByUser(c.Request.Context(), authUserID, page, limit)
	if err != nil {
		h.log.WithError(err).Errorf("GetAllOrdersByUser: Failed for user %d", authUserID)
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{
			Error: "Failed to retrieve orders",
		})
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
// @Summary Update an order
// @Description Update order details for a specific user's order. The {id} in the path is validated against the authenticated user ID from the token.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "User ID" Format(uint) // Changed from userID to id
// @Param orderID path int true "Order ID" Format(uint)
// @Param order body order_model.UpdateOrderRequest true "Order update data"
// @Success 200 {object} order_model.OrderResponse "Updated order"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid input data or ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} common_handler.ErrorResponse "Forbidden (trying to update another user's order)"
// @Failure 404 {object} common_handler.ErrorResponse "Order not found or access denied"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id}/orders/{orderID} [put] // Changed from /api/users/{userID}/orders/{orderID}
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return // Response handled by checkUserIDMatch
	}

	// Read order ID using the new parameter name 'orderID'
	orderIDStr := c.Param("orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("UpdateOrder: Invalid orderID format '%s'", orderIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{
			Error: "Invalid order ID format",
		})
		return
	}

	var req order_model.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warnf("UpdateOrder: Invalid request format for order %d", orderID)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{
			Error:   "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	// Pass both orderID and authUserID to service for update and ownership check
	order, err := h.orderService.UpdateOrder(c.Request.Context(), uint(orderID), authUserID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{
				Error: "Order not found or access denied",
			})
		} else {
			h.log.WithError(err).Errorf("UpdateOrder: Failed for order %d for user %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{
				Error: "Failed to update order",
			})
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
// @Description Delete an order by ID for a specific user. The {id} in the path is validated against the authenticated user ID from the token.
// @Tags Orders
// @Produce json
// @Param id path int true "User ID" Format(uint) // Changed from userID to id
// @Param orderID path int true "Order ID" Format(uint)
// @Success 204 "Order deleted"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} common_handler.ErrorResponse "Forbidden (trying to delete another user's order)"
// @Failure 404 {object} common_handler.ErrorResponse "Order not found or access denied"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id}/orders/{orderID} [delete] // Changed from /api/users/{userID}/orders/{orderID}
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	authUserID, ok := h.checkUserIDMatch(c)
	if !ok {
		return // Response handled by checkUserIDMatch
	}

	// Read order ID using the new parameter name 'orderID'
	orderIDStr := c.Param("orderID")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("DeleteOrder: Invalid orderID format '%s'", orderIDStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{
			Error: "Invalid order ID format",
		})
		return
	}

	// Pass both orderID and authUserID to service for deletion and ownership check
	if err := h.orderService.DeleteOrder(c.Request.Context(), uint(orderID), authUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{
				Error: "Order not found or access denied",
			})
		} else {
			h.log.WithError(err).Errorf("DeleteOrder: Failed for order %d for user %d", orderID, authUserID)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{
				Error: "Failed to delete order",
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
