package order_rep

import (
	"context"
	"errors"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// helper to create test DB and migrate
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&order_model.Order{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newTestRepo(t *testing.T) *orderRepository {
	db := setupTestDB(t)
	log := logrus.New()
	return &orderRepository{db: db, log: log}
}

func TestCreateOrder_Success(t *testing.T) {
	repo := newTestRepo(t)
	order := &order_model.Order{
		UserID:      1,
		ProductName: "Test Product",
		Quantity:    2,
		Price:       100,
	}
	err := repo.Create(context.Background(), order)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.ID == 0 {
		t.Error("expected order ID to be set after creation")
	}
}

func TestCreateOrder_NilOrder(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.Create(context.Background(), nil)
	if err == nil || !errors.Is(err, ErrDatabaseError) {
		t.Errorf("expected ErrDatabaseError, got %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	repo := newTestRepo(t)
	order := &order_model.Order{
		UserID:      2,
		ProductName: "Test2",
		Quantity:    1,
		Price:       50,
	}
	if err := repo.Create(context.Background(), order); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	got, err := repo.GetByID(context.Background(), order.ID, order.UserID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.ID != order.ID || got.UserID != order.UserID {
		t.Errorf("unexpected order returned: %+v", got)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	_, err := repo.GetByID(context.Background(), 999, 1)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestGetByID_ZeroID(t *testing.T) {
	repo := newTestRepo(t)
	_, err := repo.GetByID(context.Background(), 0, 0)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestGetAllByUser_Success(t *testing.T) {
	repo := newTestRepo(t)
	userID := uint(10)
	for i := 0; i < 5; i++ {
		repo.Create(context.Background(), &order_model.Order{
			UserID:      userID,
			ProductName: "Bulk",
			Quantity:    1,
			Price:       10,
		})
	}
	orders, total, err := repo.GetAllByUser(context.Background(), userID, 0, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 5 || total != 5 {
		t.Errorf("expected 5 orders, got %d, total %d", len(orders), total)
	}
}

func TestGetAllByUser_Empty(t *testing.T) {
	repo := newTestRepo(t)
	orders, total, err := repo.GetAllByUser(context.Background(), 12345, 0, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 0 || total != 0 {
		t.Errorf("expected 0 orders, got %d, total %d", len(orders), total)
	}
}

func TestGetAllByUser_InvalidUserID(t *testing.T) {
	repo := newTestRepo(t)
	_, _, err := repo.GetAllByUser(context.Background(), 0, 0, 10)
	if err == nil || !errors.Is(err, ErrDatabaseError) {
		t.Errorf("expected ErrDatabaseError, got %v", err)
	}
}

func TestUpdateOrder_Success(t *testing.T) {
	repo := newTestRepo(t)
	order := &order_model.Order{
		UserID:      20,
		ProductName: "Old",
		Quantity:    1,
		Price:       5,
	}
	repo.Create(context.Background(), order)
	order.ProductName = "New"
	order.Quantity = 2
	err := repo.Update(context.Background(), order)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	got, _ := repo.GetByID(context.Background(), order.ID, order.UserID)
	if got.ProductName != "New" || got.Quantity != 2 {
		t.Errorf("update did not persist: %+v", got)
	}
}

func TestDeleteOrder_Success(t *testing.T) {
	repo := newTestRepo(t)
	order := &order_model.Order{
		UserID:      30,
		ProductName: "DeleteMe",
		Quantity:    1,
		Price:       1,
	}
	repo.Create(context.Background(), order)
	err := repo.Delete(context.Background(), order.ID, order.UserID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = repo.GetByID(context.Background(), order.ID, order.UserID)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound after delete, got %v", err)
	}
}

func TestDeleteOrder_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.Delete(context.Background(), 999, 1)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestDeleteOrder_InvalidID(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.Delete(context.Background(), 0, 0)
	if err == nil || !errors.Is(err, ErrDatabaseError) {
		t.Errorf("expected ErrDatabaseError, got %v", err)
	}
}

func TestNewGormOrderRepository_NilLogger(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormOrderRepository(db, nil)
	if repo == nil {
		t.Error("expected repo, got nil")
	}
}
