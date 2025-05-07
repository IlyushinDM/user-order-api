package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/IlyushinDM/user-order-api/internal/models"
)

// MockOrderService is a mock implementation of OrderService
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID uint, req models.CreateOrderRequest) (*models.Order, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, orderID, userID uint) (*models.Order, error) {
	args := m.Called(ctx, orderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) GetAllOrdersByUser(ctx context.Context, userID uint, page, limit int) ([]models.Order, int64, error) {
	args := m.Called(ctx, userID, page, limit)
	return args.Get(0).([]models.Order), int64(args.Int(1)), args.Error(2)
}

func (m *MockOrderService) UpdateOrder(ctx context.Context, orderID, userID uint, req models.UpdateOrderRequest) (*models.Order, error) {
	args := m.Called(ctx, orderID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) DeleteOrder(ctx context.Context, orderID, userID uint) error {
	args := m.Called(ctx, orderID, userID)
	return args.Error(0)
}

func setupOrderHandlerTest(t *testing.T) (*gin.Context, *httptest.ResponseRecorder, *MockOrderService, *OrderHandler) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("logger", &logrus.Logger{}) // Set a logger

	mockOrderService := new(MockOrderService)
	orderHandler := NewOrderHandler(mockOrderService, &logrus.Logger{})

	return c, w, mockOrderService, orderHandler
}

func TestCreateOrder(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Mock request body
	reqBody := models.CreateOrderRequest{
		ProductName: "Test Product",
		Quantity:    1,
		Price:       99.99,
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", uint(1)) // Simulate authenticated user

	// Mock service response
	mockOrderService.On("CreateOrder", mock.Anything, uint(1), reqBody).Return(&models.Order{
		ID:          1,
		UserID:      1,
		ProductName: "Test Product",
		Quantity:    1,
		Price:       99.99,
	}, nil)

	// Call the handler
	orderHandler.CreateOrder(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"product_name":"Test Product"`)
	mockOrderService.AssertExpectations(t)
}

func TestGetOrderByID(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Set up context
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/orders/1", nil)
	c.Set("userID", uint(1))

	// Mock service response
	mockOrderService.On("GetOrderByID", mock.Anything, uint(1), uint(1)).Return(&models.Order{
		ID:          1,
		UserID:      1,
		ProductName: "Test Product",
		Quantity:    1,
		Price:       99.99,
	}, nil)

	// Call the handler
	orderHandler.GetOrderByID(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"product_name":"Test Product"`)
	mockOrderService.AssertExpectations(t)
}

func TestGetAllOrdersByUser(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Set up context
	c.Request, _ = http.NewRequest(http.MethodGet, "/orders?page=1&limit=10", nil)
	c.Set("userID", uint(1))

	// Mock service response
	mockOrderService.On("GetAllOrdersByUser", mock.Anything, uint(1), 1, 10).Return([]models.Order{
		{ID: 1, UserID: 1, ProductName: "Test Product", Quantity: 1, Price: 99.99},
	}, int64(1), nil)

	// Call the handler
	orderHandler.GetAllOrdersByUser(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"product_name":"Test Product"`)
	mockOrderService.AssertExpectations(t)
}

func TestUpdateOrder(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Set up context
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	reqBody := models.UpdateOrderRequest{
		ProductName: "Updated Product",
		Quantity:    2,
		Price:       199.98,
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPut, "/orders/1", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", uint(1))

	// Mock service response
	mockOrderService.On("UpdateOrder", mock.Anything, uint(1), uint(1), reqBody).Return(&models.Order{
		ID:          1,
		UserID:      1,
		ProductName: "Updated Product",
		Quantity:    2,
		Price:       199.98,
	}, nil)

	// Call the handler
	orderHandler.UpdateOrder(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"product_name":"Updated Product"`)
	mockOrderService.AssertExpectations(t)
}

func TestDeleteOrder(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Set up context
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/orders/1", nil)
	c.Set("userID", uint(1))

	// Mock service response
	mockOrderService.On("DeleteOrder", mock.Anything, uint(1), uint(1)).Return(nil)

	// Call the handler
	orderHandler.DeleteOrder(c)

	// Assertions
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockOrderService.AssertExpectations(t)
}

func TestCreateOrder_InvalidInput(t *testing.T) {
	c, w, _, orderHandler := setupOrderHandlerTest(t)

	// Mock request body with invalid data (missing required field)
	reqBody := map[string]interface{}{
		"quantity": 1,
		"price":    99.99,
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", uint(1)) // Simulate authenticated user

	// Call the handler
	orderHandler.CreateOrder(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid input data")
}

func TestGetOrderByID_InvalidIDFormat(t *testing.T) {
	c, w, _, orderHandler := setupOrderHandlerTest(t)

	// Set up context with invalid order ID format
	c.Params = []gin.Param{{Key: "id", Value: "invalid_id"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/orders/invalid_id", nil)
	c.Set("userID", uint(1))

	// Call the handler
	orderHandler.GetOrderByID(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid order ID format")
}

func TestGetOrderByID_OrderNotFound(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Set up context
	orderID := uint(999) // Non-existent order ID
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(int(orderID))}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/orders/"+strconv.Itoa(int(orderID)), nil)
	c.Set("userID", uint(1))

	// Mock service to return error indicating order not found
	mockOrderService.On("GetOrderByID", mock.Anything, orderID, uint(1)).Return(nil, gorm.ErrRecordNotFound)

	// Call the handler
	orderHandler.GetOrderByID(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Order not found or access denied")
}

func TestUpdateOrder_Unauthorized(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Set up context
	orderID := uint(1)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(int(orderID))}}
	reqBody := models.UpdateOrderRequest{
		ProductName: "Updated Product",
		Quantity:    2,
		Price:       199.98,
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPut, "/orders/"+strconv.Itoa(int(orderID)), bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", uint(1)) // Simulate authenticated user

	// Mock service to return permission denied error
	mockOrderService.On("UpdateOrder", mock.Anything, orderID, uint(1), reqBody).Return(nil, errors.New("permission denied or record not found"))

	// Call the handler
	orderHandler.UpdateOrder(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 to avoid leaking information
	assert.Contains(t, w.Body.String(), "Order not found or access denied")
	mockOrderService.AssertExpectations(t)
}

func TestDeleteOrder_OrderNotFound(t *testing.T) {
	c, w, mockOrderService, orderHandler := setupOrderHandlerTest(t)

	// Set up context
	orderID := uint(999) // Non-existent order ID
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(int(orderID))}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/orders/"+strconv.Itoa(int(orderID)), nil)
	c.Set("userID", uint(1))

	// Mock service to return error indicating order not found
	mockOrderService.On("DeleteOrder", mock.Anything, orderID, uint(1)).Return(gorm.ErrRecordNotFound)

	// Call the handler
	orderHandler.DeleteOrder(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Order not found or access denied")
}
