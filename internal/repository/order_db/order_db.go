package order_db

import (
	"context"
	"errors"

	// "github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OrderRepository defines the interface for order data operations.
type OrderRepository interface {
	Create(ctx context.Context, order *order_model.Order) error
	Update(ctx context.Context, order *order_model.Order) error
	Delete(ctx context.Context, id uint, userID uint) error                        // Ensure user owns the order
	GetByID(ctx context.Context, id uint, userID uint) (*order_model.Order, error) // Ensure user owns the order
	GetAllByUser(ctx context.Context, userID uint, page, limit int) ([]order_model.Order, int64, error)
}

type GormOrderRepository struct {
	DB  *gorm.DB
	log *logrus.Logger
}

// NewGormOrderRepository creates a new order repository using GORM.
func NewGormOrderRepository(DB *gorm.DB, log *logrus.Logger) OrderRepository {
	return &GormOrderRepository{DB: DB, log: log}
}

func (r *GormOrderRepository) Create(ctx context.Context, order *order_model.Order) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Create")
	result := r.DB.WithContext(ctx).Create(order)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to create order")
		return result.Error
	}
	logger.WithField("order_id", order.ID).WithField("user_id", order.UserID).Info("Order created successfully")
	return nil
}

func (r *GormOrderRepository) Update(ctx context.Context, order *order_model.Order) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Update").WithField("order_id", order.ID)

	// Important: Ensure the update only happens if the UserID matches.
	// The service layer should fetch the order first to verify ownership before preparing the update data.
	// Here we assume `order` contains the ID and the fields to update.
	// We select specific fields to prevent accidental updates of UserID.
	result := r.DB.WithContext(ctx).Model(&order_model.Order{}).Where("id = ? AND user_id = ?", order.ID, order.UserID).
		Select("ProductName", "Quantity", "Price", "UpdatedAt"). // Specify updatable fields explicitly
		Updates(order)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to update order")
		return result.Error
	}
	if result.RowsAffected == 0 {
		// This could mean the order doesn't exist OR it belongs to another user.
		// We check existence first.
		var exists int64
		r.DB.WithContext(ctx).Model(&order_model.Order{}).Where("id = ?", order.ID).Count(&exists)
		if exists == 0 {
			logger.Warn("Order update attempted but order not found")
			return gorm.ErrRecordNotFound
		}
		// If it exists, it means the UserID didn't match (or no fields changed)
		logger.Warn("Order update attempted but no rows affected (possibly wrong user or no changes)")
		// Consider returning a specific "permission denied" error if needed, though the service layer should handle this.
		// return errors.New("record not found or permission denied")
		return gorm.ErrRecordNotFound // Simplest approach for now
	}
	logger.Info("Order updated successfully")
	return nil
}

func (r *GormOrderRepository) Delete(ctx context.Context, id uint, userID uint) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Delete").WithField("order_id", id).WithField("user_id", userID)
	// Soft delete only if the order belongs to the user
	result := r.DB.WithContext(ctx).Where("user_id = ?", userID).Delete(&order_model.Order{}, id)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to delete order")
		return result.Error
	}
	if result.RowsAffected == 0 {
		logger.Warn("Order deletion attempted but no rows affected (order not found or wrong user)")
		// Check if the order exists at all to differentiate
		var exists int64
		r.DB.WithContext(ctx).Model(&order_model.Order{}).Where("id = ?", id).Count(&exists)
		if exists == 0 {
			return gorm.ErrRecordNotFound // Order doesn't exist
		}
		// Order exists but doesn't belong to the user
		return errors.New("permission denied or record not found") // More specific error
	}
	logger.Info("Order deleted successfully (soft delete)")
	return nil
}

func (r *GormOrderRepository) GetByID(ctx context.Context, id uint, userID uint) (*order_model.Order, error) {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.GetByID").WithField("order_id", id).WithField("user_id", userID)
	var order order_model.Order
	// Find the order only if it belongs to the specified user
	result := r.DB.WithContext(ctx).Where("user_id = ?", userID).First(&order, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("Order not found for this user")
		} else {
			logger.WithError(result.Error).Error("Failed to get order by ID for user")
		}
		return nil, result.Error
	}
	logger.Info("Order retrieved successfully")
	return &order, nil
}

func (r *GormOrderRepository) GetAllByUser(ctx context.Context, userID uint, page, limit int) ([]order_model.Order, int64, error) {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.GetAllByUser").WithField("user_id", userID)
	var orders []order_model.Order
	var total int64

	query := r.DB.WithContext(ctx).Model(&order_model.Order{}).Where("user_id = ?", userID)

	// Count total records for the user
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count user orders")
		return nil, 0, err
	}

	logger.WithField("total_orders_for_user", total).Debug("Counted user orders")

	// Apply pagination
	offset := (page - 1) * limit
	result := query.Offset(offset).Limit(limit).Find(&orders)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to retrieve paginated user orders")
		return nil, 0, result.Error
	}

	logger.WithFields(logrus.Fields{
		"page":            page,
		"limit":           limit,
		"offset":          offset,
		"retrieved_count": len(orders),
		"total_count":     total,
	}).Info("User orders retrieved successfully")

	return orders, total, nil
}
