package services

import (
	"context"
	"errors"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockOrderRepository is a mock implementation of OrderRepository.
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	if args.Error(0) != nil {
		return args.Error(0)
	}
	// If no error, set the ID (as if the DB did)
	order.ID = uint(1)
	return nil
}

func (m *MockOrderRepository) Update(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, orderID uint, userID uint) error {
	args := m.Called(ctx, orderID, userID)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	args := m.Called(ctx, orderID, userID)
	order := args.Get(0).(*models.Order)
	return order, args.Error(1)
}

func (m *MockOrderRepository) GetAllByUser(ctx context.Context, userID uint, page, limit int) ([]models.Order, int64, error) {
	args := m.Called(ctx, userID, page, limit)
	orders := args.Get(0).([]models.Order)
	total := args.Get(1).(int64)
	return orders, total, args.Error(2)
}

// MockUserRepository is a mock implementation of UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	user := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	user := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	args := m.Called(ctx, page, limit, filters)
	users := args.Get(0).([]models.User)
	total := args.Get(1).(int64)
	return users, total, args.Error(2)
}

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// Test case 1: Successful order creation
	t.Run("Success", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository) // You might not need this if you remove the user existence check
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		userID := uint(1)
		req := models.CreateOrderRequest{
			ProductName: "Test Product",
			Quantity:    2,
			Price:       25.00,
		}

		// Define what the mock should return
		mockOrderRepo.On("Create", ctx, mock.MatchedBy(func(order *models.Order) bool {
			return order.ProductName == req.ProductName && order.Quantity == req.Quantity && order.Price == req.Price && order.UserID == userID
		})).Return(nil)

		order, err := orderService.CreateOrder(ctx, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, req.ProductName, order.ProductName)
		assert.Equal(t, req.Quantity, order.Quantity)
		assert.Equal(t, req.Price, order.Price)
		assert.Equal(t, userID, order.UserID)

		mockOrderRepo.AssertExpectations(t)
	})

	// Test case 2: Order creation fails in repository
	t.Run("RepoFailure", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository) // You might not need this if you remove the user existence check
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		userID := uint(1)
		req := models.CreateOrderRequest{
			ProductName: "Test Product",
			Quantity:    2,
			Price:       25.00,
		}

		// Define what the mock should return (an error)
		mockOrderRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

		order, err := orderService.CreateOrder(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, order)
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderService_UpdateOrder(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		orderID := uint(1)
		userID := uint(1)
		req := models.UpdateOrderRequest{
			ProductName: "Updated Product",
			Quantity:    3,
			Price:       30.00,
		}

		existingOrder := &models.Order{
			ID:          orderID,
			UserID:      userID,
			ProductName: "Original Product",
			Quantity:    2,
			Price:       25.00,
		}

		// Mock the GetByID to return the existing order
		mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(existingOrder, nil)

		// Mock the Update to return no error
		mockOrderRepo.On("Update", ctx, mock.MatchedBy(func(order *models.Order) bool {
			return order.ID == orderID && order.ProductName == req.ProductName && order.Quantity == req.Quantity && order.Price == req.Price && order.UserID == userID
		})).Return(nil)

		updatedOrder, err := orderService.UpdateOrder(ctx, orderID, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, updatedOrder)
		assert.Equal(t, req.ProductName, updatedOrder.ProductName)
		assert.Equal(t, req.Quantity, updatedOrder.Quantity)
		assert.Equal(t, req.Price, updatedOrder.Price)

		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("OrderNotFound", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		orderID := uint(1)
		userID := uint(1)
		req := models.UpdateOrderRequest{
			ProductName: "Updated Product",
			Quantity:    3,
			Price:       30.00,
		}

		mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(nil, gorm.ErrRecordNotFound)

		updatedOrder, err := orderService.UpdateOrder(ctx, orderID, userID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedOrder)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("UpdateRepoFailure", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		orderID := uint(1)
		userID := uint(1)
		req := models.UpdateOrderRequest{
			ProductName: "Updated Product",
			Quantity:    3,
			Price:       30.00,
		}

		existingOrder := &models.Order{
			ID:          orderID,
			UserID:      userID,
			ProductName: "Original Product",
			Quantity:    2,
			Price:       25.00,
		}

		mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(existingOrder, nil)
		mockOrderRepo.On("Update", ctx, mock.Anything).Return(errors.New("database error"))

		updatedOrder, err := orderService.UpdateOrder(ctx, orderID, userID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedOrder)

		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderService_DeleteOrder(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		orderID := uint(1)
		userID := uint(1)

		mockOrderRepo.On("Delete", ctx, orderID, userID).Return(nil)

		err := orderService.DeleteOrder(ctx, orderID, userID)

		assert.NoError(t, err)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("OrderNotFound", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		orderID := uint(1)
		userID := uint(1)

		mockOrderRepo.On("Delete", ctx, orderID, userID).Return(gorm.ErrRecordNotFound)

		err := orderService.DeleteOrder(ctx, orderID, userID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetOrderByID(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		orderID := uint(1)
		userID := uint(1)

		expectedOrder := &models.Order{
			ID:          orderID,
			UserID:      userID,
			ProductName: "Test Product",
			Quantity:    2,
			Price:       25.00,
		}

		mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(expectedOrder, nil)

		order, err := orderService.GetOrderByID(ctx, orderID, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrder, order)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("OrderNotFound", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		orderID := uint(1)
		userID := uint(1)

		mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(nil, gorm.ErrRecordNotFound)

		order, err := orderService.GetOrderByID(ctx, orderID, userID)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetAllOrdersByUser(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		userID := uint(1)
		page := 1
		limit := 10

		expectedOrders := []models.Order{
			{ID: 1, UserID: userID, ProductName: "Product 1", Quantity: 1, Price: 10.00},
			{ID: 2, UserID: userID, ProductName: "Product 2", Quantity: 2, Price: 20.00},
		}
		var total int64 = 2

		mockOrderRepo.On("GetAllByUser", ctx, userID, page, limit).Return(expectedOrders, total, nil)

		orders, totalOrders, err := orderService.GetAllOrdersByUser(ctx, userID, page, limit)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, orders)
		assert.Equal(t, total, totalOrders)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("RepoFailure", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockUserRepo := new(MockUserRepository)
		orderService := NewOrderService(mockOrderRepo, mockUserRepo, logger)

		userID := uint(1)
		page := 1
		limit := 10

		mockOrderRepo.On("GetAllByUser", ctx, userID, page, limit).Return(nil, int64(0), errors.New("database error"))

		orders, totalOrders, err := orderService.GetAllOrdersByUser(ctx, userID, page, limit)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, int64(0), totalOrders)
		mockOrderRepo.AssertExpectations(t)
	})
}
