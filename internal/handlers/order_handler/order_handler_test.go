package order_handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/services/order_service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Мок ---

type mockOrderService struct {
	mock.Mock
}

func (m *mockOrderService) CreateOrder(ctx context.Context, userID uint, req order_model.CreateOrderRequest) (*order_model.Order, error) {
	args := m.Called(ctx, userID, req)
	order, _ := args.Get(0).(*order_model.Order)
	return order, args.Error(1)
}

func (m *mockOrderService) GetOrderByID(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
	args := m.Called(ctx, orderID, userID)
	order, _ := args.Get(0).(*order_model.Order)
	return order, args.Error(1)
}

func (m *mockOrderService) GetAllOrdersByUser(ctx context.Context, userID uint, page, limit int) ([]order_model.Order, int64, error) {
	args := m.Called(ctx, userID, page, limit)
	orders, _ := args.Get(0).([]order_model.Order)
	var total int64
	switch v := args.Get(1).(type) {
	case int64:
		total = v
	case int:
		total = int64(v)
	default:
		total = 0
	}
	return orders, total, args.Error(2)
}

func (m *mockOrderService) UpdateOrder(ctx context.Context, orderID, userID uint, req order_model.UpdateOrderRequest) (*order_model.Order, error) {
	args := m.Called(ctx, orderID, userID, req)
	order, _ := args.Get(0).(*order_model.Order)
	return order, args.Error(1)
}

func (m *mockOrderService) DeleteOrder(ctx context.Context, orderID, userID uint) error {
	args := m.Called(ctx, orderID, userID)
	return args.Error(0)
}

type mockCommonHandler struct {
	mock.Mock
}

func (m *mockCommonHandler) GetPaginationParams(c *gin.Context) (int, int, error) {
	args := m.Called(c)
	return args.Int(0), args.Int(1), args.Error(2)
}

// Добавлен недостающий метод GetFilteringParams
func (m *mockCommonHandler) GetFilteringParams(c *gin.Context) (map[string]any, error) {
	args := m.Called(c)
	filters, _ := args.Get(0).(map[string]any)
	return filters, args.Error(1)
}

// Ensure mockCommonHandler implements common_handler.CommonHandlerInterface
// Изменено: проверка реализации CommonHandlerInterface
var _ common_handler.CommonHandlerInterface = (*mockCommonHandler)(nil)

func addAuthUserID(c *gin.Context, userID uint) {
	c.Set("userID", userID)
}

// --- Тесты ---

func TestCreateOrder_Success(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler) // mockCommon теперь *mockCommonHandler, который реализует интерфейс
	log := logrus.New()
	// NewOrderHandler теперь ожидает common_handler.CommonHandlerInterface
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	reqBody := order_model.CreateOrderRequest{
		ProductName: "TestProduct",
		Quantity:    2,
		Price:       100,
	}
	order := &order_model.Order{
		ID:          10,
		UserID:      userID,
		ProductName: reqBody.ProductName,
		Quantity:    reqBody.Quantity,
		Price:       reqBody.Price,
	}
	mockSvc.On("CreateOrder", mock.Anything, userID, reqBody).Return(order, nil)

	// router := setupRouter(handler)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users/1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.CreateOrder(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp order_model.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, resp.ID)
	assert.Equal(t, order.UserID, resp.UserID)
	assert.Equal(t, order.ProductName, resp.ProductName)
	assert.Equal(t, order.Quantity, resp.Quantity)
	assert.Equal(t, order.Price, resp.Price)
}

func TestCreateOrder_BadRequest_BindJSON(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	// router := setupRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users/1/orders", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, uint(1))

	handler.CreateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_Forbidden_UserIDMismatch(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users/2/orders", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	addAuthUserID(c, uint(1))

	handler.CreateOrder(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetOrderByID_Success(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	orderID := uint(10)
	order := &order_model.Order{
		ID:          orderID,
		UserID:      userID,
		ProductName: "TestProduct",
		Quantity:    2,
		Price:       100,
	}
	mockSvc.On("GetOrderByID", mock.Anything, orderID, userID).Return(order, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders/10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, userID)

	handler.GetOrderByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp order_model.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, resp.ID)
}

func TestGetOrderByID_NotFound(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	orderID := uint(10)
	mockSvc.On("GetOrderByID", mock.Anything, orderID, userID).Return(nil, order_service.ErrOrderNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders/10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, userID)

	handler.GetOrderByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAllOrdersByUser_Success(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	page, limit := 1, 10
	orders := []order_model.Order{
		{ID: 1, UserID: userID, ProductName: "A", Quantity: 1, Price: 10},
		{ID: 2, UserID: userID, ProductName: "B", Quantity: 2, Price: 20},
	}
	total := int64(2)

	mockCommon.On("GetPaginationParams", mock.Anything).Return(page, limit, nil)
	// При необходимости, настройте ожидание для GetFilteringParams, даже если он не используется в GetAllOrdersByUser напрямую
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)
	mockSvc.On("GetAllOrdersByUser", mock.Anything, userID, page, limit).Return(orders, total, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders?page=1&limit=10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.GetAllOrdersByUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp order_model.PaginatedOrdersResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, total, resp.Total)
	assert.Len(t, resp.Orders, 2)
}

func TestGetAllOrdersByUser_PaginationError(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	mockCommon.On("GetPaginationParams", mock.Anything).Return(0, 0, errors.New("bad pagination"))
	// При необходимости, настройте ожидание для GetFilteringParams
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders?page=abc", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.GetAllOrdersByUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateOrder_Success(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	orderID := uint(10)
	reqBody := order_model.UpdateOrderRequest{
		ProductName: "Updated",
		Quantity:    3,
		Price:       200,
	}
	order := &order_model.Order{
		ID:          orderID,
		UserID:      userID,
		ProductName: reqBody.ProductName,
		Quantity:    reqBody.Quantity,
		Price:       reqBody.Price,
	}
	mockSvc.On("UpdateOrder", mock.Anything, orderID, userID, reqBody).Return(order, nil)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/users/1/orders/10", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, userID)

	handler.UpdateOrder(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp order_model.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, resp.ID)
	assert.Equal(t, order.ProductName, resp.ProductName)
}

// Дополнительный тест для ErrNoUpdateFields в UpdateOrder
func TestUpdateOrder_NoUpdateFields(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	orderID := uint(10)
	reqBody := order_model.UpdateOrderRequest{
		ProductName: "NoChange",
	}
	existingOrder := &order_model.Order{
		ID:          orderID,
		UserID:      userID,
		ProductName: "NoChange", // Предполагаем, что текущее имя продукта такое же
		Quantity:    5,
		Price:       250,
	}

	mockSvc.On("UpdateOrder", mock.Anything, orderID, userID, reqBody).Return(nil, order_service.ErrNoUpdateFields)
	// Мок сервиса для GetOrderByID, так как хендлер вызывает его при ErrNoUpdateFields
	mockSvc.On("GetOrderByID", mock.Anything, orderID, userID).Return(existingOrder, nil)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/users/1/orders/10", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, userID)

	handler.UpdateOrder(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp order_model.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, existingOrder.ID, resp.ID)
	assert.Equal(t, existingOrder.ProductName, resp.ProductName)
}

func TestDeleteOrder_Success(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	orderID := uint(10)
	mockSvc.On("DeleteOrder", mock.Anything, orderID, userID).Return(nil)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/users/1/orders/10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, userID)

	handler.DeleteOrder(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteOrder_NotFound(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	orderID := uint(10)
	mockSvc.On("DeleteOrder", mock.Anything, orderID, userID).Return(order_service.ErrOrderNotFound)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/users/1/orders/10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, userID)

	handler.DeleteOrder(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCheckUserIDMatch_AuthMissing(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	uid, ok := handler.checkUserIDMatch(c)
	assert.False(t, ok)
	assert.Equal(t, uint(0), uid)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCheckUserIDMatch_BadUserIDFormat(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/abc/orders", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	addAuthUserID(c, uint(1))
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	uid, ok := handler.checkUserIDMatch(c)
	assert.False(t, ok)
	assert.Equal(t, uint(0), uid)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCheckUserIDMatch_Forbidden(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/2/orders", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	addAuthUserID(c, uint(1))
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	uid, ok := handler.checkUserIDMatch(c)
	assert.False(t, ok)
	assert.Equal(t, uint(0), uid)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- Additional error cases for coverage ---

func TestCreateOrder_ServiceError(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	reqBody := order_model.CreateOrderRequest{
		ProductName: "TestProduct",
		Quantity:    2,
		Price:       100,
	}
	mockSvc.On("CreateOrder", mock.Anything, userID, reqBody).Return(nil, order_service.ErrInvalidServiceInput)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users/1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.CreateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetOrderByID_BadOrderIDFormat(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders/abc", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "abc"},
	}
	addAuthUserID(c, userID)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	handler.GetOrderByID(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateOrder_BadOrderIDFormat(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/users/1/orders/abc", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "abc"},
	}
	addAuthUserID(c, userID)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	handler.UpdateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateOrder_BadRequest_BindJSON(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/users/1/orders/10", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, userID)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	handler.UpdateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteOrder_BadOrderIDFormat(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/users/1/orders/abc", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "orderID", Value: "abc"},
	}
	addAuthUserID(c, userID)
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	handler.DeleteOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteOrder_Forbidden_UserIDMismatch(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/users/2/orders/10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "2"},
		{Key: "orderID", Value: "10"},
	}
	addAuthUserID(c, uint(1))
	// При необходимости, настройте ожидание для GetPaginationParams и GetFilteringParams
	mockCommon.On("GetPaginationParams", mock.Anything).Return(1, 10, nil)
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	handler.DeleteOrder(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAllOrdersByUser_ServiceError(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	page, limit := 1, 10
	mockCommon.On("GetPaginationParams", mock.Anything).Return(page, limit, nil)
	// При необходимости, настройте ожидание для GetFilteringParams
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)
	mockSvc.On("GetAllOrdersByUser", mock.Anything, userID, page, limit).Return(nil, int64(0), order_service.ErrInvalidServiceInput)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders?page=1&limit=10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.GetAllOrdersByUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllOrdersByUser_Forbidden_UserIDMismatch(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	page, limit := 1, 10
	mockCommon.On("GetPaginationParams", mock.Anything).Return(page, limit, nil)
	// При необходимости, настройте ожидание для GetFilteringParams
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/2/orders?page=1&limit=10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	addAuthUserID(c, userID)

	handler.GetAllOrdersByUser(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAllOrdersByUser_BadRequest(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	page, limit := 0, 0
	mockCommon.On("GetPaginationParams", mock.Anything).Return(page, limit, errors.New("bad pagination"))
	// При необходимости, настройте ожидание для GetFilteringParams
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders?page=abc", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.GetAllOrdersByUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllOrdersByUser_EmptyOrders(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	page, limit := 1, 10
	orders := []order_model.Order{}
	total := int64(0)

	mockCommon.On("GetPaginationParams", mock.Anything).Return(page, limit, nil)
	// При необходимости, настройте ожидание для GetFilteringParams
	mockCommon.On("GetFilteringParams", mock.Anything).Return(nil, nil)
	mockSvc.On("GetAllOrdersByUser", mock.Anything, userID, page, limit).Return(orders, total, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders?page=1&limit=10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.GetAllOrdersByUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp order_model.PaginatedOrdersResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, total, resp.Total)
	assert.Len(t, resp.Orders, 0)
}

func TestGetAllOrdersByUser_EmptyOrdersWithFilters(t *testing.T) {
	mockSvc := new(mockOrderService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewOrderHandler(mockSvc, mockCommon, log)

	userID := uint(1)
	page, limit := 1, 10
	orders := []order_model.Order{}
	total := int64(0)

	mockCommon.On("GetPaginationParams", mock.Anything).Return(page, limit, nil)
	// Настройка ожидания для GetFilteringParams с пустыми фильтрами
	mockCommon.On("GetFilteringParams", mock.Anything).Return(map[string]any{}, nil)
	mockSvc.On("GetAllOrdersByUser", mock.Anything, userID, page, limit).Return(orders, total, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1/orders?page=1&limit=10", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	addAuthUserID(c, userID)

	handler.GetAllOrdersByUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp order_model.PaginatedOrdersResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, total, resp.Total)
	assert.Len(t, resp.Orders, 0)
}
