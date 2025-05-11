package order_rep

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Определение пользовательских ошибок репозитория заказов
var (
	ErrOrderNotFound         = errors.New("заказ не найден")
	ErrDatabaseError         = errors.New("ошибка базы данных")
	ErrNoRowsAffected        = errors.New("ни одна запись не затронута")
	ErrOrderNotBelongsToUser = errors.New("заказ не принадлежит пользователю")
)

// OrderRepository определяет интерфейс для взаимодействия с данными заказов в базе данных.
type OrderRepository interface {
	Create(ctx context.Context, order *order_model.Order) error
	GetByID(ctx context.Context, orderID uint, userID uint) (*order_model.Order, error)
	GetAllByUser(ctx context.Context, userID uint, offset, limit int) ([]order_model.Order, int64, error)
	Update(ctx context.Context, order *order_model.Order) error
	Delete(ctx context.Context, orderID uint, userID uint) error
}

// orderRepository реализует интерфейс OrderRepository с использованием GORM.
type orderRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewOrderRepository создает новый экземпляр репозитория заказов.
func NewGormOrderRepository(db *gorm.DB, log *logrus.Logger) OrderRepository {
	if db == nil {
		logrus.Fatal("Экземпляр GORM DB равен nil в NewOrderRepository")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Экземпляр Logrus logger равен nil в NewOrderRepository, используется логгер по умолчанию")
		log = defaultLog
	}
	return &orderRepository{db: db, log: log}
}

// Create вставляет новый заказ в базу данных.
func (r *orderRepository) Create(ctx context.Context, order *order_model.Order) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Create")
	if order == nil {
		logger.Warn("Попытка создать nil заказ")
		return fmt.Errorf("%w: невозможно создать nil заказ", ErrDatabaseError)
	}

	// Добавление логов для отладки
	logger = logger.WithFields(logrus.Fields{
		"user_id":      order.UserID,
		"product_name": order.ProductName,
		"quantity":     order.Quantity,
		"price":        order.Price,
	})
	logger.Debug("Создание нового заказа")

	result := r.db.WithContext(ctx).Create(order)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Не удалось создать заказ в базе данных")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.WithField("order_id", order.ID).Info("Заказ успешно создан")
	return nil
}

// GetByID извлекает заказ по его ID и ID пользователя для проверки владения.
func (r *orderRepository) GetByID(ctx context.Context, orderID uint, userID uint) (*order_model.Order, error) {
	logger := r.log.WithContext(ctx).WithField(
		"method", "OrderRepository.GetByID").WithFields(logrus.Fields{"order_id": orderID, "user_id": userID})
	if orderID == 0 || userID == 0 {
		logger.Warn("Попытка получить заказ с нулевым ID или ID пользователя")
		return nil, ErrOrderNotFound
	}

	logger.Debug("Получение заказа по ID и ID пользователя")
	var order order_model.Order
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).First(&order)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("Заказ не найден для данного ID и ID пользователя")
			return nil, ErrOrderNotFound
		}
		logger.WithError(result.Error).Error("Не удалось получить заказ из базы данных")
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.Info("Заказ успешно получен по ID и ID пользователя")
	return &order, nil
}

// GetAllByUser извлекает все заказы для конкретного пользователя с пагинацией.
func (r *orderRepository) GetAllByUser(
	ctx context.Context, userID uint, offset, limit int,
) ([]order_model.Order, int64, error) {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.GetAllByUser").WithField(
		"user_id", userID).WithFields(logrus.Fields{"offset": offset, "limit": limit})
	if userID == 0 {
		logger.Warn("Попытка получить заказы для пользователя с нулевым ID")
		return nil, 0, fmt.Errorf("%w: ID пользователя должен быть положительным", ErrDatabaseError)
	}
	// Валидация offset и limit (хотя сервис тоже может валидировать)
	if offset < 0 {
		offset = 0
		logger.Warn("Предоставлен отрицательный offset, по умолчанию установлено 0")
	}
	if limit <= 0 {
		limit = 10
		logger.Warn("Предоставлен неверный или неположительный limit, по умолчанию установлено 10")
	}

	logger.Debug("Получение всех заказов по ID пользователя с пагинацией")

	var orders []order_model.Order
	var total int64

	countResult := r.db.WithContext(ctx).Model(&order_model.Order{}).Where("user_id = ?", userID).Count(&total)
	if countResult.Error != nil {
		logger.WithError(countResult.Error).Error("Не удалось получить общее количество заказов для пользователя")
		return nil, 0, fmt.Errorf("%w: не удалось подсчитать заказы пользователя", ErrDatabaseError)
	}

	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&orders)
	if result.Error != nil {
		// Если это не ErrRecordNotFound (который для Find просто означает пустой список, а не ошибку)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.WithError(result.Error).Error("Не удалось получить постраничные заказы для пользователя")
			return nil, 0, fmt.Errorf("%w: не удалось получить постраничные заказы пользователя", ErrDatabaseError)
		}
	}

	logger.WithField("count", len(orders)).Info("Заказы успешно получены по ID пользователя")
	return orders, total, nil
}

// Обновление изменяет существующий порядок в базе данных
func (r *orderRepository) Update(ctx context.Context, order *order_model.Order) error {
	logger := r.log.WithContext(ctx).WithField("method", "OrderRepository.Update").WithField(
		"order_id", order.ID).WithField("user_id", order.UserID)
	if order == nil || order.ID == 0 || order.UserID == 0 {
		logger.Warn("Попытка обновить nil заказ или заказ с нулевым ID/ID пользователя")
		return fmt.Errorf("%w: неверный объект заказа для обновления", ErrDatabaseError)
	}

	logger.Debug("Обновление заказа в базе данных")
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", order.ID, order.UserID).Save(order)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Не удалось обновить заказ в базе данных")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	// Проверка rows affected полезна для операций обновления/удаления, чтобы понять, была ли запись найдена
	if result.RowsAffected == 0 {
		logger.Warn("Операция обновления затронула 0 записей, заказ не найден или не принадлежит пользователю?")
		return ErrOrderNotFound
	}

	logger.Info("Заказ успешно обновлен")
	return nil
}

// Удаляет заказ из базы данных по ID и ID пользователя для проверки владения
func (r *orderRepository) Delete(ctx context.Context, orderID uint, userID uint) error {
	logger := r.log.WithContext(ctx).WithField(
		"method", "OrderRepository.Delete").WithFields(logrus.Fields{"order_id": orderID, "user_id": userID})
	if orderID == 0 || userID == 0 {
		logger.Warn("Попытка удалить заказ с нулевым ID или ID пользователя")
		return fmt.Errorf("%w: неверный ID заказа или ID пользователя для удаления", ErrDatabaseError)
	}

	logger.Debug("Удаление заказа из базы данных")
	// Используем Delete и проверяем user_id для гарантии владения
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).Delete(&order_model.Order{})

	if result.Error != nil {
		logger.WithError(result.Error).Error("Не удалось удалить заказ из базы данных")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	// Проверка rows affected
	if result.RowsAffected == 0 {
		logger.Warn("Операция удаления затронула 0 записей, заказ не найден или не принадлежит пользователю?")
		return ErrOrderNotFound
	}

	logger.Info("Заказ успешно удален")
	return nil
}
