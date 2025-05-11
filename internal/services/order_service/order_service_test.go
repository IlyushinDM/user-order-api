package order_service

import (
	"context"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/repository/order_rep"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// --- Mock implementation for order_rep.OrderRepository ---

type mockOrderRepo struct {
	CreateFn       func(ctx context.Context, order *order_model.Order) error
	UpdateFn       func(ctx context.Context, order *order_model.Order) error
	DeleteFn       func(ctx context.Context, orderID, userID uint) error
	GetByIDFn      func(ctx context.Context, orderID, userID uint) (*order_model.Order, error)
	GetAllByUserFn func(ctx context.Context, userID uint, offset, limit int) ([]order_model.Order, int64, error)
}

func (m *mockOrderRepo) Create(ctx context.Context, order *order_model.Order) error {
	return m.CreateFn(ctx, order)
}

func (m *mockOrderRepo) Update(ctx context.Context, order *order_model.Order) error {
	return m.UpdateFn(ctx, order)
}

func (m *mockOrderRepo) Delete(ctx context.Context, orderID, userID uint) error {
	return m.DeleteFn(ctx, orderID, userID)
}

func (m *mockOrderRepo) GetByID(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
	return m.GetByIDFn(ctx, orderID, userID)
}

func (m *mockOrderRepo) GetAllByUser(ctx context.Context, userID uint, offset, limit int) ([]order_model.Order, int64, error) {
	return m.GetAllByUserFn(ctx, userID, offset, limit)
}

// --- Tests ---

func TestCreateOrder_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		CreateFn: func(ctx context.Context, order *order_model.Order) error {
			order.ID = 1
			return nil
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	req := order_model.CreateOrderRequest{
		ProductName: "TestProduct",
		Quantity:    2,
		Price:       10.5,
	}
	order, err := svc.CreateOrder(context.Background(), 42, req)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, uint(42), order.UserID)
	assert.Equal(t, "TestProduct", order.ProductName)
	assert.Equal(t, 2, order.Quantity)
	assert.Equal(t, 10.5, order.Price)
}

func TestCreateOrder_InvalidInput(t *testing.T) {
	mockRepo := &mockOrderRepo{}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	_, err := svc.CreateOrder(context.Background(), 0, order_model.CreateOrderRequest{})
	assert.ErrorIs(t, err, ErrInvalidServiceInput)

	_, err = svc.CreateOrder(context.Background(), 1, order_model.CreateOrderRequest{
		ProductName: "",
		Quantity:    0,
		Price:       0,
	})
	assert.ErrorIs(t, err, ErrInvalidServiceInput)
}

func TestCreateOrder_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		CreateFn: func(ctx context.Context, order *order_model.Order) error {
			return order_rep.ErrDatabaseError
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	req := order_model.CreateOrderRequest{
		ProductName: "Test",
		Quantity:    1,
		Price:       1,
	}
	_, err := svc.CreateOrder(context.Background(), 1, req)
	assert.ErrorIs(t, err, ErrServiceDatabaseError)
}

func TestUpdateOrder_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetByIDFn: func(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
			return &order_model.Order{
				ID:          orderID,
				UserID:      userID,
				ProductName: "Old",
				Quantity:    1,
				Price:       1,
			}, nil
		},
		UpdateFn: func(ctx context.Context, order *order_model.Order) error {
			return nil
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	req := order_model.UpdateOrderRequest{
		ProductName: "New",
		Quantity:    2,
		Price:       3,
	}
	order, err := svc.UpdateOrder(context.Background(), 1, 2, req)
	assert.NoError(t, err)
	assert.Equal(t, "New", order.ProductName)
	assert.Equal(t, 2, order.Quantity)
	assert.Equal(t, 3.0, order.Price)
}

func TestUpdateOrder_NoFieldsToUpdate(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetByIDFn: func(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
			return &order_model.Order{
				ID:          orderID,
				UserID:      userID,
				ProductName: "Same",
				Quantity:    1,
				Price:       1,
			}, nil
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	req := order_model.UpdateOrderRequest{
		ProductName: "Same",
		Quantity:    1,
		Price:       1,
	}
	order, err := svc.UpdateOrder(context.Background(), 1, 2, req)
	assert.ErrorIs(t, err, ErrNoUpdateFields)
	assert.NotNil(t, order)
}

func TestUpdateOrder_RepoNotFound(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetByIDFn: func(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
			return nil, order_rep.ErrOrderNotFound
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	req := order_model.UpdateOrderRequest{
		ProductName: "New",
	}
	_, err := svc.UpdateOrder(context.Background(), 1, 2, req)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestUpdateOrder_RepoUpdateError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetByIDFn: func(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
			return &order_model.Order{
				ID:          orderID,
				UserID:      userID,
				ProductName: "Old",
				Quantity:    1,
				Price:       1,
			}, nil
		},
		UpdateFn: func(ctx context.Context, order *order_model.Order) error {
			return order_rep.ErrDatabaseError
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	req := order_model.UpdateOrderRequest{
		ProductName: "New",
	}
	_, err := svc.UpdateOrder(context.Background(), 1, 2, req)
	assert.ErrorIs(t, err, ErrServiceDatabaseError)
}

func TestDeleteOrder_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		DeleteFn: func(ctx context.Context, orderID, userID uint) error {
			return nil
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	err := svc.DeleteOrder(context.Background(), 1, 2)
	assert.NoError(t, err)
}

func TestDeleteOrder_RepoNotFound(t *testing.T) {
	mockRepo := &mockOrderRepo{
		DeleteFn: func(ctx context.Context, orderID, userID uint) error {
			return order_rep.ErrOrderNotFound
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	err := svc.DeleteOrder(context.Background(), 1, 2)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestDeleteOrder_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		DeleteFn: func(ctx context.Context, orderID, userID uint) error {
			return order_rep.ErrDatabaseError
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	err := svc.DeleteOrder(context.Background(), 1, 2)
	assert.ErrorIs(t, err, ErrServiceDatabaseError)
}

func TestGetOrderByID_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetByIDFn: func(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
			return &order_model.Order{
				ID:          orderID,
				UserID:      userID,
				ProductName: "Test",
				Quantity:    1,
				Price:       1,
			}, nil
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	order, err := svc.GetOrderByID(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, uint(1), order.ID)
	assert.Equal(t, uint(2), order.UserID)
}

func TestGetOrderByID_RepoNotFound(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetByIDFn: func(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
			return nil, order_rep.ErrOrderNotFound
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	_, err := svc.GetOrderByID(context.Background(), 1, 2)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestGetOrderByID_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetByIDFn: func(ctx context.Context, orderID, userID uint) (*order_model.Order, error) {
			return nil, order_rep.ErrDatabaseError
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	_, err := svc.GetOrderByID(context.Background(), 1, 2)
	assert.ErrorIs(t, err, ErrServiceDatabaseError)
}

func TestGetAllOrdersByUser_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetAllByUserFn: func(ctx context.Context, userID uint, offset, limit int) ([]order_model.Order, int64, error) {
			return []order_model.Order{
				{ID: 1, UserID: userID, ProductName: "A", Quantity: 1, Price: 1},
				{ID: 2, UserID: userID, ProductName: "B", Quantity: 2, Price: 2},
			}, 2, nil
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	orders, total, err := svc.GetAllOrdersByUser(context.Background(), 2, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, int64(2), total)
}

func TestGetAllOrdersByUser_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		GetAllByUserFn: func(ctx context.Context, userID uint, offset, limit int) ([]order_model.Order, int64, error) {
			return nil, 0, order_rep.ErrDatabaseError
		},
	}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	_, _, err := svc.GetAllOrdersByUser(context.Background(), 2, 1, 10)
	assert.ErrorIs(t, err, ErrServiceDatabaseError)
}

func TestGetAllOrdersByUser_InvalidInput(t *testing.T) {
	mockRepo := &mockOrderRepo{}
	log := logrus.New()
	svc := NewOrderService(mockRepo, log)

	_, _, err := svc.GetAllOrdersByUser(context.Background(), 0, 1, 10)
	assert.ErrorIs(t, err, ErrInvalidServiceInput)
}
