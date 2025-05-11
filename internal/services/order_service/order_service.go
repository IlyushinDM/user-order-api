package order_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/IlyushinDM/user-order-api/internal/repository/order_rep"
	"github.com/sirupsen/logrus"
)

// Определение ошибок сервисного слоя для заказов
var (
	ErrOrderNotFound        = errors.New("заказ не найден")
	ErrInvalidServiceInput  = errors.New("недопустимые входные данные сервиса")
	ErrServiceDatabaseError = errors.New("ошибка базы данных сервиса")
	ErrNoUpdateFields       = errors.New("нет полей для обновления")
)

// OrderService определяет интерфейс для бизнес-логики заказов
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
	log       *logrus.Logger
}

// NewOrderService создает новый сервис заказов
func NewOrderService(
	orderRepo order_rep.OrderRepository,
	log *logrus.Logger,
) OrderService {
	if orderRepo == nil {
		logrus.Fatal("Экземпляр OrderRepository равен nil в NewOrderService")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Экземпляр логгера Logrus равен nil в NewOrderService, используется логгер по умолчанию")
		log = defaultLog
	}
	return &orderService{orderRepo: orderRepo, log: log}
}

func (s *orderService) CreateOrder(ctx context.Context,
	userID uint,
	req order_model.CreateOrderRequest,
) (*order_model.Order, error) {
	logger := s.log.WithContext(ctx).WithField(
		"method",
		"OrderService.CreateOrder").WithField("user_id", userID)

	// Базовая валидация входных данных сервиса
	if userID == 0 {
		logger.Warn("Попытка создать заказ с нулевым ID пользователя")
		return nil, fmt.Errorf("%w: ID пользователя должен быть положительным", ErrInvalidServiceInput)
	}
	if req.ProductName == "" || req.Quantity <= 0 || req.Price <= 0 {
		logger.WithFields(logrus.Fields{
			"product_name": req.ProductName,
			"quantity":     req.Quantity,
			"price":        req.Price,
		}).Warn("Недопустимые входные данные для создания заказа")
		return nil, fmt.Errorf(
			"%w: название продукта, количество и цена обязательны и должны быть положительными", ErrInvalidServiceInput)
	}

	order := &order_model.Order{
		UserID:      userID,
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		logger.WithError(err).Error("Не удалось создать заказ в репозитории")
		// Маппинг ошибок репозитория на ошибки сервиса
		switch {
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: не удалось сохранить заказ в базу данных", ErrServiceDatabaseError)
		default:
			return nil, fmt.Errorf("%w: не удалось создать заказ", ErrServiceDatabaseError)
		}
	}

	logger.WithField("order_id", order.ID).Info("Заказ успешно создан")
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

	// Базовая валидация входных данных сервиса
	if orderID == 0 || userID == 0 {
		logger.Warn("Попытка обновить заказ с нулевым ID заказа или ID пользователя")
		return nil, fmt.Errorf("%w: ID заказа и ID пользователя должны быть положительными", ErrInvalidServiceInput)
	}

	// Первым шагом получаем заказ для проверки существования и текущих данных
	// Репозиторий GetByID включает проверку user_id
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		logger.WithError(err).Error("Не удалось получить заказ для обновления из репозитория")
		// Маппинг ошибок репозитория
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound):
			logger.Warn("Обновление не удалось: Заказ не найден для данного ID и ID пользователя")
			return nil, ErrOrderNotFound
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: ошибка базы данных при поиске заказа для обновления", ErrServiceDatabaseError)
		default:
			return nil, fmt.Errorf("%w: не удалось найти заказ для обновления", ErrServiceDatabaseError)
		}
	}

	updated := false
	if req.ProductName != "" && req.ProductName != order.ProductName {
		order.ProductName = req.ProductName
		updated = true
		logger.Debug("Обновление названия продукта заказа")
	}
	if req.Quantity > 0 && req.Quantity != order.Quantity {
		order.Quantity = req.Quantity
		updated = true
		logger.Debug("Обновление количества заказа")
	}
	if req.Price > 0 && req.Price != order.Price {
		order.Price = req.Price
		updated = true
		logger.Debug("Обновление цены заказа")
	}

	if !updated {
		logger.Info("Нет полей для обновления заказа")
		return order, ErrNoUpdateFields
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.WithError(err).Error("Не удалось обновить заказ в репозитории")
		// Маппинг ошибок репозитория
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound):
			logger.Warn("Обновление не удалось в репозитории: Заказ не найден при сохранении")
			return nil, ErrOrderNotFound
		case errors.Is(err, order_rep.ErrNoRowsAffected):
			logger.Warn("Обновление не удалось в репозитории: Строки не затронуты при сохранении")
			return nil, ErrOrderNotFound
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: ошибка базы данных при сохранении обновленного заказа", ErrServiceDatabaseError)
		default:
			return nil, fmt.Errorf("%w: не удалось сохранить обновленный заказ", ErrServiceDatabaseError)
		}
	}

	logger.Info("Заказ успешно обновлен")
	return order, nil
}

func (s *orderService) DeleteOrder(ctx context.Context, orderID uint, userID uint) error {
	logger := s.log.WithContext(ctx).WithField(
		"method",
		"OrderService.DeleteOrder").WithField("order_id", orderID).WithField("user_id", userID)

	// Базовая валидация входных данных сервиса
	if orderID == 0 || userID == 0 {
		logger.Warn("Попытка удалить заказ с нулевым ID заказа или ID пользователя")
		return fmt.Errorf("%w: ID заказа и ID пользователя должны быть положительными", ErrInvalidServiceInput)
	}

	// Удаление в репозитории включает проверку user_id
	err := s.orderRepo.Delete(ctx, orderID, userID)
	if err != nil {
		logger.WithError(err).Error("Не удалось удалить заказ в репозитории")
		// Маппинг ошибок репозитория
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound):
			logger.Warn("Удаление не удалось: Заказ не найден для данного ID и ID пользователя")
			return ErrOrderNotFound
		case errors.Is(err, order_rep.ErrNoRowsAffected):
			logger.Warn("Удаление не удалось: Строки не затронуты")
			return ErrOrderNotFound
		case errors.Is(err, order_rep.ErrDatabaseError):
			return fmt.Errorf("%w: ошибка базы данных при удалении заказа", ErrServiceDatabaseError)
		default:
			return fmt.Errorf("%w: не удалось удалить заказ", ErrServiceDatabaseError)
		}
	}

	logger.Info("Заказ успешно удален")
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

	// Базовая валидация входных данных сервиса
	if orderID == 0 || userID == 0 {
		logger.Warn("Попытка получить заказ с нулевым ID заказа или ID пользователя")
		return nil, fmt.Errorf("%w: ID заказа и ID пользователя должны быть положительными", ErrInvalidServiceInput)
	}

	// Метод GetByID репозитория включает проверку user_id
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		logger.WithError(err).Error("Не удалось получить заказ из репозитория")
		// Маппинг ошибок репозитория
		switch {
		case errors.Is(err, order_rep.ErrOrderNotFound):
			logger.Warn("Заказ не найден для данного ID и ID пользователя")
			return nil, ErrOrderNotFound
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, fmt.Errorf("%w: ошибка базы данных при получении заказа по ID", ErrServiceDatabaseError)
		default:
			return nil, fmt.Errorf("%w: не удалось получить заказ по ID", ErrServiceDatabaseError)
		}
	}

	logger.Info("Заказ успешно получен")
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
		"OrderService.GetAllOrdersByUser").WithField(
		"user_id", userID).WithFields(logrus.Fields{"page": page, "limit": limit})

	// Базовая валидация и преобразование пагинации для репозитория
	if userID == 0 {
		logger.Warn("Попытка получить заказы для нулевого ID пользователя")
		return nil, 0, fmt.Errorf("%w: ID пользователя должен быть положительным", ErrInvalidServiceInput)
	}
	if page <= 0 {
		page = 1
		logger.Warn("Указан недопустимый номер страницы, используется значение по умолчанию 1")
	}
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
		logger.Warn("Указан недопустимый лимит, используется значение по умолчанию 10")
	}

	offset := (page - 1) * limit
	logger = logger.WithField("offset", offset)

	// Вызываем метод репозитория со смещением и лимитом
	orders, total, err := s.orderRepo.GetAllByUser(ctx, userID, offset, limit)
	if err != nil {
		logger.WithError(err).Error("Не удалось получить заказы для пользователя из репозитория")
		switch {
		case errors.Is(err, order_rep.ErrDatabaseError):
			return nil, 0, fmt.Errorf("%w: ошибка базы данных при получении всех заказов для пользователя", ErrServiceDatabaseError)
		default:
			return nil, 0, fmt.Errorf("%w: не удалось получить все заказы для пользователя", ErrServiceDatabaseError)
		}
	}

	logger.WithFields(
		logrus.Fields{
			"count": len(orders),
			"total": total,
		}).Info("Заказы для пользователя успешно получены")
	return orders, total, nil
}
