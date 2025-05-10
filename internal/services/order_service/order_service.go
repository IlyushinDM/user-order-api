package order_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/repository/order_rep" // Используем рефакторингованный репозиторий

	// user_rep больше не нужен напрямую, так как репозиторий заказов отвечает за проверку user_id
	// "github.com/IlyushinDM/user-order-api/internal/repository/user_rep"
	"github.com/sirupsen/logrus"
	// gorm.io/gorm больше не нужен напрямую здесь, ошибки GORM обрабатываются на уровне репозитория
)

// **Определение ошибок сервисного слоя для заказов**
var (
	// ErrOrderNotFound возвращается, когда заказ не найден (либо не существует, либо не принадлежит пользователю).
	ErrOrderNotFound = errors.New("order not found")
	// ErrInvalidServiceInput возвращается, если входные данные для метода сервиса недопустимы.
	ErrInvalidServiceInput = errors.New("invalid service input")
	// ErrServiceDatabaseError возвращается при ошибках, возникших при взаимодействии с репозиторием.
	ErrServiceDatabaseError = errors.New("service database error")
	// ErrNoUpdateFields возвращается, если при обновлении не были предоставлены поля для изменения.
	ErrNoUpdateFields = errors.New("no fields to update")
)

// OrderService defines the interface for order business logic.
type OrderService interface {
	CreateOrder(ctx context.Context, userID uint,
		req order_model.CreateOrderRequest) (*order_model.Order, error)
	UpdateOrder(ctx context.Context, orderID uint,
		userID uint, req order_model.UpdateOrderRequest) (*order_model.Order, error)
	DeleteOrder(ctx context.Context, orderID uint, userID uint) error
	GetOrderByID(ctx context.Context, orderID uint, userID uint) (*order_model.Order, error)
	GetAllOrdersByUser(ctx context.Context,
		userID uint, page, limit int) ([]order_model.Order, int64, error)
}

type orderService struct {
	orderRepo order_rep.OrderRepository
	// userRepo Dependency removed as ownership check is now primarily in order_rep
	log *logrus.Logger
}

// NewOrderService creates a new order service.
// **Улучшена обработка nil зависимостей.**
func NewOrderService(
	orderRepo order_rep.OrderRepository,
	// userRepo user_rep.UserRepository, // Removed
	log *logrus.Logger,
) OrderService {
	if orderRepo == nil {
		logrus.Fatal("OrderRepository instance is nil in NewOrderService")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Logrus logger instance is nil in NewOrderService, using default logger")
		log = defaultLog
	}
	return &orderService{orderRepo: orderRepo, log: log} // Removed userRepo from struct init
}

func (s *orderService) CreateOrder(ctx context.Context,
	userID uint,
	req order_model.CreateOrderRequest,
) (*order_model.Order, error) {
	logger := s.log.WithContext(ctx).WithField(
		"method",
		"OrderService.CreateOrder").WithField("user_id", userID)

	// **Базовая валидация входных данных сервиса**
	if userID == 0 {
		logger.Warn("Attempted to create order with zero user ID")
		return nil, fmt.Errorf("%w: user ID must be positive", ErrInvalidServiceInput)
	}
	if req.ProductName == "" || req.Quantity <= 0 || req.Price <= 0 {
		logger.WithFields(logrus.Fields{
			"product_name": req.ProductName,
			"quantity":     req.Quantity,
			"price":        req.Price,
		}).Warn("Invalid input for order creation")
		return nil, fmt.Errorf("%w: product name, quantity, and price are required and must be positive", ErrInvalidServiceInput)
	}

	order := &order_model.Order{
		UserID:      userID, // Set the user ID from the authenticated user
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}

	// Call repository method
	if err := s.orderRepo.Create(ctx, order); err != nil {
		logger.WithError(err).Error("Failed to create order in repository")
		// **Маппинг ошибок репозитория на ошибки сервиса**
		switch {
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: failed to save order to database", ErrServiceDatabaseError)
		default:
			// Неизвестная ошибка репозитория
			return nil, fmt.Errorf("%w: failed to create order", ErrServiceDatabaseError) // Оборачиваем в общую ошибку БД сервиса
		}
	}

	logger.WithField("order_id", order.ID).Info("Order created successfully")
	return order, nil
}

func (s *orderService) UpdateOrder(
	ctx context.Context,
	orderID uint,
	userID uint,
	req order_model.UpdateOrderRequest,
) (*order_model.Order, error) {
	logger := s.log.WithContext(ctx).WithField(
		"method",
		"OrderService.UpdateOrder").WithField("order_id", orderID).WithField("user_id", userID)

	// **Базовая валидация входных данных сервиса**
	if orderID == 0 || userID == 0 {
		logger.Warn("Attempted to update order with zero order ID or user ID")
		return nil, fmt.Errorf("%w: order ID and user ID must be positive", ErrInvalidServiceInput)
	}

	// Первым шагом получаем заказ для проверки существования и текущих данных
	// Репозиторий GetByID включает проверку user_id
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		logger.WithError(err).Error("Failed to get order for update from repository")
		// **Маппинг ошибок репозитория**
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound):
			logger.Warn("Update failed: Order not found for the given ID and User ID")
			return nil, ErrOrderNotFound // Маппинг на сервисную ошибку "не найдено"
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: database error finding order for update", ErrServiceDatabaseError)
		default:
			return nil, fmt.Errorf("%w: failed to find order for update", ErrServiceDatabaseError)
		}
	}

	// Apply updates from the request (only if value is provided in request)
	updated := false
	if req.ProductName != "" && req.ProductName != order.ProductName {
		order.ProductName = req.ProductName
		updated = true
		logger.Debug("Updating order product name")
	}
	// Проверка > 0 для quantity и price, чтобы обновить только если предоставлено валидное значение
	if req.Quantity > 0 && req.Quantity != order.Quantity {
		order.Quantity = req.Quantity
		updated = true
		logger.Debug("Updating order quantity")
	}
	// Используем небольшую дельту для сравнения float, или сравниваем только если req.Price > 0
	// Для простоты, сравниваем только если req.Price > 0
	if req.Price > 0 && req.Price != order.Price {
		order.Price = req.Price
		updated = true
		logger.Debug("Updating order price")
	}

	if !updated {
		logger.Info("No fields to update for order")
		// **Возвращаем ErrNoUpdateFields, чтобы вызывающий код мог обработать этот сценарий**
		return order, ErrNoUpdateFields // Возвращаем текущий объект заказа и ошибку
	}

	// The order object now contains the ID, UserID, and updated fields.
	// The repository's Update method should handle saving these changes and re-checking ownership.
	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.WithError(err).Error("Failed to update order in repository")
		// **Маппинг ошибок репозитория**
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound): // Репозиторий вернул OrderNotFound (например, если запись исчезла между Get и Update)
			logger.Warn("Update failed in repository: Order not found during save")
			return nil, ErrOrderNotFound // Маппинг на сервисную ошибку "не найдено"
		case errors.Is(err, order_rep.ErrNoRowsAffected): // Если репозиторий вернул NoRowsAffected (менее вероятно после GetByID)
			logger.Warn("Update failed in repository: No rows affected during save")
			return nil, ErrOrderNotFound // Считаем, что это тоже означает, что заказ не был найден
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: database error saving updated order", ErrServiceDatabaseError)
		default:
			return nil, fmt.Errorf("%w: failed to save updated order", ErrServiceDatabaseError)
		}
	}

	logger.Info("Order updated successfully")
	// В идеале, получить обновленный объект из БД после Save,
	// но для простоты возвращаем модифицированный в памяти объект, если репозиторий не возвращает его.
	// Если репозиторий.Update возвращает *order_model.Order, можно вернуть его.
	// order, err = s.orderRepo.GetByID(ctx, orderID, userID) // Опционально, чтобы быть уверенным в актуальности
	// if err != nil { ... handle error ... }
	return order, nil // Возвращаем модифицированный объект
}

func (s *orderService) DeleteOrder(ctx context.Context, orderID uint, userID uint) error {
	logger := s.log.WithContext(ctx).WithField(
		"method",
		"OrderService.DeleteOrder").WithField("order_id", orderID).WithField("user_id", userID)

	// **Базовая валидация входных данных сервиса**
	if orderID == 0 || userID == 0 {
		logger.Warn("Attempted to delete order with zero order ID or user ID")
		return fmt.Errorf("%w: order ID and user ID must be positive", ErrInvalidServiceInput)
	}

	// The repository delete includes the userID check
	err := s.orderRepo.Delete(ctx, orderID, userID)
	if err != nil {
		logger.WithError(err).Error("Failed to delete order in repository")
		// **Маппинг ошибок репозитория**
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound): // Репозиторий вернул OrderNotFound (включая проверку user_id)
			logger.Warn("Deletion failed: Order not found for the given ID and User ID")
			return ErrOrderNotFound // Маппинг на сервисную ошибку "не найдено"
		case errors.Is(err, order_rep.ErrNoRowsAffected): // Если репозиторий вернул NoRowsAffected
			logger.Warn("Deletion failed: No rows affected")
			return ErrOrderNotFound // Считаем, что это тоже означает, что заказ не был найден
		case errors.Is(err, order_rep.ErrDatabaseError):
			return fmt.Errorf("%w: database error deleting order", ErrServiceDatabaseError)
		default:
			return fmt.Errorf("%w: failed to delete order", ErrServiceDatabaseError)
		}
	}

	logger.Info("Order deleted successfully")
	return nil
}

func (s *orderService) GetOrderByID(
	ctx context.Context,
	orderID uint,
	userID uint,
) (*order_model.Order, error) {
	logger := s.log.WithContext(ctx).WithField(
		"method",
		"OrderService.GetOrderByID").WithField("order_id", orderID).WithField("user_id", userID)

	// **Базовая валидация входных данных сервиса**
	if orderID == 0 || userID == 0 {
		logger.Warn("Attempted to get order with zero order ID or user ID")
		return nil, fmt.Errorf("%w: order ID and user ID must be positive", ErrInvalidServiceInput)
	}

	// Repository GetByID includes the userID check
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		logger.WithError(err).Error("Failed to get order from repository")
		// **Маппинг ошибок репозитория**
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound): // Репозиторий вернул OrderNotFound (включая проверку user_id)
			logger.Warn("Order not found for the given ID and User ID")
			return nil, ErrOrderNotFound // Маппинг на сервисную ошибку "не найдено"
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: database error getting order by ID", ErrServiceDatabaseError)
		default:
			return nil, fmt.Errorf("%w: failed to retrieve order by ID", ErrServiceDatabaseError)
		}
	}

	logger.Info("Order retrieved successfully")
	return order, nil
}

func (s *orderService) GetAllOrdersByUser(
	ctx context.Context,
	userID uint,
	page,
	limit int,
) ([]order_model.Order, int64, error) {
	logger := s.log.WithContext(ctx).WithField(
		"method",
		"OrderService.GetAllOrdersByUser").WithField("user_id", userID).WithFields(logrus.Fields{"page": page, "limit": limit})

	// **Базовая валидация и преобразование пагинации для репозитория**
	if userID == 0 {
		logger.Warn("Attempted to get orders for zero user ID")
		return nil, 0, fmt.Errorf("%w: user ID must be positive", ErrInvalidServiceInput)
	}
	if page <= 0 {
		page = 1
		logger.Warn("Invalid page number provided, defaulting to 1")
	}
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
		logger.Warn("Invalid limit provided, defaulting to 10")
	}

	offset := (page - 1) * limit
	logger = logger.WithField("offset", offset) // Добавляем offset в лог

	// Call repository method with offset and limit
	orders, total, err := s.orderRepo.GetAllByUser(ctx, userID, offset, limit)
	if err != nil {
		logger.WithError(err).Error("Failed to get orders for user from repository")
		// **Маппинг ошибок репозитория**
		switch {
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, 0, fmt.Errorf("%w: database error getting all orders for user", ErrServiceDatabaseError)
		default:
			return nil, 0, fmt.Errorf("%w: failed to retrieve all orders for user", ErrServiceDatabaseError)
		}
	}

	logger.WithFields(
		logrus.Fields{
			"count": len(orders),
			"total": total,
		}).Info("Retrieved orders for user successfully")
	return orders, total, nil
}
