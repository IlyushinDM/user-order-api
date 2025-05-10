package order_rep

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model" // Предполагаем наличие модели заказа
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// **Определение пользовательских ошибок репозитория заказов**
var (
	// ErrOrderNotFound возвращается, когда заказ не найден в базе данных.
	ErrOrderNotFound = errors.New("order not found")
	// ErrDatabaseError возвращается при возникновении ошибок взаимодействия с базой данных.
	ErrDatabaseError = errors.New("database error")
	// ErrNoRowsAffected возвращается, когда операция обновления или удаления не затронула ни одной записи,
	// что часто указывает на то, что запись не найдена или не было изменений.
	ErrNoRowsAffected = errors.New("no rows affected")
	// ErrOrderNotBelongsToUser возвращается, когда заказ принадлежит другому пользователю при попытке доступа/изменения.
	ErrOrderNotBelongsToUser = errors.New("order does not belong to user")
)

// OrderRepository defines the interface for interacting with order data in the database.
type OrderRepository interface {
	// Create inserts a new order into the database.
	Create(ctx context.Context, order *order_model.Order) error
	// GetByID retrieves an order by its ID and optionally by user ID for ownership check.
	GetByID(ctx context.Context, orderID uint, userID uint) (*order_model.Order, error)
	// GetAllByUser retrieves all orders for a specific user with pagination.
	// Returns a slice of orders, the total count, and an error.
	GetAllByUser(ctx context.Context, userID uint, offset, limit int) ([]order_model.Order, int64, error)
	// Update modifies an existing order in the database.
	Update(ctx context.Context, order *order_model.Order) error
	// Delete removes an order from the database by ID and user ID for ownership check.
	Delete(ctx context.Context, orderID uint, userID uint) error
}

// orderRepository implements the OrderRepository interface using GORM.
type orderRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewOrderRepository creates a new order repository instance.
// **Улучшена обработка nil зависимостей.**
func NewGormOrderRepository(db *gorm.DB, log *logrus.Logger) OrderRepository {
	if db == nil {
		logrus.Fatal("GORM DB instance is nil in NewOrderRepository")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Logrus logger instance is nil in NewOrderRepository, using default logger")
		log = defaultLog
	}
	return &orderRepository{db: db, log: log}
}

// Create inserts a new order into the database.
func (r *orderRepository) Create(ctx context.Context, order *order_model.Order) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Create")
	if order == nil {
		logger.Warn("Attempted to create a nil order")
		return fmt.Errorf("%w: cannot create nil order", ErrDatabaseError) // Ошибка входных данных, маппим на ошибку БД
	}

	// Добавление логов для отладки
	logger = logger.WithFields(logrus.Fields{
		"user_id":      order.UserID,
		"product_name": order.ProductName,
		"quantity":     order.Quantity,
		"price":        order.Price,
	})
	logger.Debug("Creating new order")

	result := r.db.WithContext(ctx).Create(order)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to create order in database")
		// Оборачиваем ошибку GORM в пользовательскую ошибку репозитория.
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.WithField("order_id", order.ID).Info("Order created successfully")
	return nil
}

// GetByID retrieves an order by its ID and user ID for ownership check.
func (r *orderRepository) GetByID(ctx context.Context, orderID uint, userID uint) (*order_model.Order, error) {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.GetByID").WithFields(logrus.Fields{"order_id": orderID, "user_id": userID})
	if orderID == 0 || userID == 0 {
		logger.Warn("Attempted to get order with zero ID or user ID")
		// Возвращаем "не найдено", т.к. нулевые ID не соответствуют реальным записям.
		return nil, ErrOrderNotFound
	}

	logger.Debug("Getting order by ID and User ID")
	var order order_model.Order
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).First(&order)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("Order not found for the given ID and User ID")
			return nil, ErrOrderNotFound // Маппинг gorm.ErrRecordNotFound на пользовательскую ошибку
		}
		logger.WithError(result.Error).Error("Failed to get order from database")
		// Оборачиваем другие ошибки GORM.
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.Info("Order retrieved successfully by ID and User ID")
	return &order, nil
}

// GetAllByUser retrieves all orders for a specific user with pagination.
func (r *orderRepository) GetAllByUser(ctx context.Context, userID uint, offset, limit int) ([]order_model.Order, int64, error) {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.GetAllByUser").WithField("user_id", userID).WithFields(logrus.Fields{"offset": offset, "limit": limit})
	if userID == 0 {
		logger.Warn("Attempted to get orders for zero user ID")
		return nil, 0, fmt.Errorf("%w: user ID must be positive", ErrDatabaseError) // Ошибка входных данных
	}
	// Валидация offset и limit (хотя сервис тоже может валидировать)
	if offset < 0 {
		offset = 0
		logger.Warn("Negative offset provided, defaulting to 0")
	}
	if limit <= 0 {
		limit = 10 // Значение по умолчанию, если limit некорректен или не задан
		logger.Warn("Invalid or non-positive limit provided, defaulting to 10")
	}

	logger.Debug("Getting all orders by User ID with pagination")

	var orders []order_model.Order
	var total int64

	// Get total count first
	countResult := r.db.WithContext(ctx).Model(&order_model.Order{}).Where("user_id = ?", userID).Count(&total)
	if countResult.Error != nil {
		logger.WithError(countResult.Error).Error("Failed to get total count of orders for user")
		return nil, 0, fmt.Errorf("%w: failed to count user orders", ErrDatabaseError)
	}

	// Get paginated orders
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&orders)
	if result.Error != nil {
		// Если это не ErrRecordNotFound (который для Find просто означает пустой список, а не ошибку)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.WithError(result.Error).Error("Failed to get paginated orders for user")
			return nil, 0, fmt.Errorf("%w: failed to retrieve paginated user orders", ErrDatabaseError)
		}
		// Если ErrRecordNotFound, просто возвращаем пустой список и total=0 (уже посчитан).
	}

	logger.WithField("count", len(orders)).Info("Orders retrieved successfully by User ID")
	return orders, total, nil
}

// Update modifies an existing order in the database.
// Assumes the order object passed contains the ID and updated fields.
func (r *orderRepository) Update(ctx context.Context, order *order_model.Order) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Update").WithField("order_id", order.ID).WithField("user_id", order.UserID)
	if order == nil || order.ID == 0 || order.UserID == 0 {
		logger.Warn("Attempted to update a nil order or order with zero ID/User ID")
		return fmt.Errorf("%w: invalid order object for update", ErrDatabaseError) // Ошибка входных данных
	}

	// GORM по умолчанию обновляет все поля, включая нулевые.
	// Если нужно обновить только измененные поля, используйте Select или Omit.
	// Например, result := r.db.WithContext(ctx).Model(order).Where("id = ? AND user_id = ?", order.ID, order.UserID).Updates(order_model.Order{ProductName: order.ProductName, Quantity: order.Quantity, Price: order.Price})
	// Или просто Updates(order) если модель имеет gorm:"-" для полей, которые не должны обновляться.

	logger.Debug("Updating order in database")
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", order.ID, order.UserID).Save(order) // Save обновляет всю запись, включая нулевые значения

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to update order in database")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	// Проверка rows affected полезна для операций обновления/удаления, чтобы понять, была ли запись найдена.
	if result.RowsAffected == 0 {
		// Если 0 rows affected, возможно, запись не найдена по ID или UserID.
		// Или если использовался Updates с Select, и не было изменений в выбранных полях.
		// В данном случае (используя Save с условием WHERE id AND user_id), 0 rows affected
		// скорее всего означает, что заказ не был найден для данного пользователя.
		logger.Warn("Update operation affected 0 rows, order not found or not owned by user?")
		return ErrOrderNotFound // Маппинг на ErrOrderNotFound, так как это наиболее вероятная причина
	}

	logger.Info("Order updated successfully")
	return nil
}

// Delete removes an order from the database by ID and user ID for ownership check.
func (r *orderRepository) Delete(ctx context.Context, orderID uint, userID uint) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Delete").WithFields(logrus.Fields{"order_id": orderID, "user_id": userID})
	if orderID == 0 || userID == 0 {
		logger.Warn("Attempted to delete order with zero ID or user ID")
		return fmt.Errorf("%w: invalid order ID or user ID for delete", ErrDatabaseError) // Ошибка входных данных
	}

	logger.Debug("Deleting order from database")
	// Используем Delete и проверяем user_id для гарантии владения.
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).Delete(&order_model.Order{})

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to delete order from database")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	// Проверка rows affected
	if result.RowsAffected == 0 {
		logger.Warn("Delete operation affected 0 rows, order not found or not owned by user?")
		// В зависимости от требований, можно вернуть ErrOrderNotFound или ErrOrderNotBelongsToUser.
		// ErrOrderNotFound покрывает оба случая, если точная причина не важна на этом уровне.
		return ErrOrderNotFound // Маппинг на ErrOrderNotFound
	}

	logger.Info("Order deleted successfully")
	return nil
}
