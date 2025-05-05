package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OrderService defines the interface for order business logic.
type OrderService interface {
	CreateOrder(ctx context.Context, userID uint, req models.CreateOrderRequest) (*models.Order, error)
	UpdateOrder(ctx context.Context, orderID uint, userID uint, req models.UpdateOrderRequest) (*models.Order, error)
	DeleteOrder(ctx context.Context, orderID uint, userID uint) error
	GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	GetAllOrdersByUser(ctx context.Context, userID uint, page, limit int) ([]models.Order, int64, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
	userRepo  repository.UserRepository // Optional: Might need to check user existence
	log       *logrus.Logger
}

// NewOrderService creates a new order service.
func NewOrderService(orderRepo repository.OrderRepository, userRepo repository.UserRepository, log *logrus.Logger) OrderService {
	return &orderService{orderRepo: orderRepo, userRepo: userRepo, log: log}
}

func (s *orderService) CreateOrder(ctx context.Context, userID uint, req models.CreateOrderRequest) (*models.Order, error) {
	logger := s.log.WithContext(ctx).WithField("method", "OrderService.CreateOrder").WithField("user_id", userID)

	// Optional: Check if user exists (if not implicitly handled by DB foreign key constraints)
	// _, err := s.userRepo.GetByID(ctx, userID)
	// if err != nil {
	//  if errors.Is(err, gorm.ErrRecordNotFound) {
	//      logger.Warn("Attempt to create order for non-existent user")
	//      return nil, errors.New("user not found")
	//  }
	//  logger.WithError(err).Error("Failed to check user existence")
	//  return nil, fmt.Errorf("database error checking user: %w", err)
	// }

	order := &models.Order{
		UserID:      userID, // Set the user ID from the authenticated user
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		logger.WithError(err).Error("Failed to create order in repository")
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	logger.WithField("order_id", order.ID).Info("Order created successfully")
	return order, nil
}

func (s *orderService) UpdateOrder(ctx context.Context, orderID uint, userID uint, req models.UpdateOrderRequest) (*models.Order, error) {
	logger := s.log.WithContext(ctx).WithField("method", "OrderService.UpdateOrder").WithField("order_id", orderID).WithField("user_id", userID)

	// First, get the order to ensure it exists and belongs to the user
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Update failed: Order not found or doesn't belong to user")
			return nil, err // Return specific error
		}
		logger.WithError(err).Error("Failed to get order for update")
		return nil, fmt.Errorf("database error finding order: %w", err)
	}

	// Apply updates from the request
	updated := false
	if req.ProductName != "" && req.ProductName != order.ProductName {
		order.ProductName = req.ProductName
		updated = true
		logger.Debug("Updating order product name")
	}
	if req.Quantity > 0 && req.Quantity != order.Quantity {
		order.Quantity = req.Quantity
		updated = true
		logger.Debug("Updating order quantity")
	}
	if req.Price > 0 && req.Price != order.Price { // Be careful with float comparison, maybe use a tolerance
		order.Price = req.Price
		updated = true
		logger.Debug("Updating order price")
	}

	if !updated {
		logger.Info("No fields to update for order")
		return order, nil // Return current order data if no changes
	}

	// The order object now contains the ID and updated fields.
	// The repository's Update method should handle saving these changes.
	// We pass the full order object including the UserID for the repository to double-check ownership during the update transaction.
	if err := s.orderRepo.Update(ctx, order); err != nil {
		// The repo's Update should return ErrRecordNotFound if the ownership check fails during the UPDATE query itself.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Update failed in repository: Order not found or permission denied during update")
		} else {
			logger.WithError(err).Error("Failed to update order in repository")
		}
		return nil, fmt.Errorf("failed to save updated order: %w", err)
	}

	logger.Info("Order updated successfully")
	// Fetch the potentially updated order again if repository.Update doesn't return it
	// updatedOrder, err := s.orderRepo.GetByID(ctx, orderID, userID) ...
	return order, nil // Return the modified order object
}

func (s *orderService) DeleteOrder(ctx context.Context, orderID uint, userID uint) error {
	logger := s.log.WithContext(ctx).WithField("method", "OrderService.DeleteOrder").WithField("order_id", orderID).WithField("user_id", userID)

	// The repository delete includes the userID check
	err := s.orderRepo.Delete(ctx, orderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Deletion failed: Order not found")
		} else if err.Error() == "permission denied or record not found" { // Check specific error from repo
			logger.Warn("Deletion failed: Order not found or permission denied")
		} else {
			logger.WithError(err).Error("Failed to delete order in repository")
		}
		return err // Propagate error
	}

	logger.Info("Order deleted successfully")
	return nil
}

func (s *orderService) GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	logger := s.log.WithContext(ctx).WithField("method", "OrderService.GetOrderByID").WithField("order_id", orderID).WithField("user_id", userID)

	// Repository GetByID includes the userID check
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Order not found for user")
		} else {
			logger.WithError(err).Error("Failed to get order from repository")
		}
		return nil, err
	}

	logger.Info("Order retrieved successfully")
	return order, nil
}

func (s *orderService) GetAllOrdersByUser(ctx context.Context, userID uint, page, limit int) ([]models.Order, int64, error) {
	logger := s.log.WithContext(ctx).WithField("method", "OrderService.GetAllOrdersByUser").WithField("user_id", userID)

	orders, total, err := s.orderRepo.GetAllByUser(ctx, userID, page, limit)
	if err != nil {
		logger.WithError(err).Error("Failed to get orders for user from repository")
		return nil, 0, err
	}

	logger.WithFields(logrus.Fields{"count": len(orders), "total": total}).Info("Retrieved orders for user successfully")
	return orders, total, nil
}
