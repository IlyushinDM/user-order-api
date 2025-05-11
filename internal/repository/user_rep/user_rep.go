package user_rep

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Определение пользовательских ошибок
var (
	ErrUserNotFound   = errors.New("пользователь не найден")
	ErrDatabaseError  = errors.New("операция с базой данных не удалась")
	ErrNoRowsAffected = errors.New("нет затронутых записей")
	ErrInvalidInput   = errors.New("неверный входной параметр")
)

// UserRepository определяет интерфейс для операций с пользовательскими данными.
type UserRepository interface {
	Create(ctx context.Context, user *user_model.User) error
	Update(ctx context.Context, user *user_model.User) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*user_model.User, error)
	GetByEmail(ctx context.Context, email string) (*user_model.User, error)
	GetAll(ctx context.Context, params ListQueryParams) ([]user_model.User, int64, error)
}

// Структура ListQueryParams для типобезопасных фильтров GetAll
type ListQueryParams struct {
	Page  int
	Limit int
	// Фильтры: отсутствие фильтра (nil) и фильтр со значением по умолчанию (например, 0).
	MinAge *int
	MaxAge *int
	Name   *string
}

// GormUserRepository реализует UserRepository с использованием GORM
type GormUserRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewGormUserRepository создает новый репозиторий пользователей с использованием GORM
func NewGormUserRepository(db *gorm.DB, log *logrus.Logger) UserRepository {
	if db == nil {
		logrus.Fatal("Экземпляр GORM DB равен nil")
	}
	if log == nil {
		// Если логгер не предоставлен, можно использовать логгер по умолчанию с предупреждением
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Экземпляр логгера Logrus равен nil в NewGormUserRepository, используется логгер по умолчанию")
		log = defaultLog
	}
	return &GormUserRepository{db: db, log: log}
}

// Create создает новую запись пользователя в базе данных
func (r *GormUserRepository) Create(ctx context.Context, user *user_model.User) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Create")

	if user == nil {
		logger.Error("Попытка создать nil пользователя")
		return fmt.Errorf("%w: объект пользователя равен nil", ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Не удалось создать пользователя")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.WithField("user_id", user.ID).Info("Пользователь успешно создан")
	return nil
}

// Update обновляет существующую запись пользователя
func (r *GormUserRepository) Update(ctx context.Context, user *user_model.User) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Update").WithField("user_id", user.ID)

	if user == nil {
		logger.Error("Попытка обновить nil пользователя")
		return fmt.Errorf("%w: объект пользователя равен nil", ErrInvalidInput)
	}
	if user.ID == 0 {
		logger.Error("Попытка обновить пользователя с нулевым ID")
		return fmt.Errorf("%w: ID пользователя равен нулю, невозможно обновить", ErrInvalidInput)
	}

	// Model(user) ограничивает обновление пользователем с заданным ID
	// Updates(user) обновляет ненулевые поля из объекта пользователя
	result := r.db.WithContext(ctx).Model(user).Updates(user)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Не удалось обновить пользователя")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Warn("Попытка обновления пользователя, но нет затронутых записей (возможно, пользователь не найден или нет изменений)")
		return ErrNoRowsAffected
	}

	logger.Info("Пользователь успешно обновлен")
	return nil
}

// Delete удаляет пользователя по ID. Выполняет мягкое удаление, если модель имеет поле DeletedAt
func (r *GormUserRepository) Delete(ctx context.Context, id uint) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Delete").WithField("user_id", id)

	if id == 0 {
		logger.Error("Попытка удалить пользователя с нулевым ID")
		return fmt.Errorf("%w: ID пользователя равен нулю, невозможно удалить", ErrInvalidInput)
	}

	// Delete GORM по умолчанию выполняет мягкое удаление, если модель имеет поле DeletedAt
	result := r.db.WithContext(ctx).Delete(&user_model.User{}, id)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Не удалось удалить пользователя")
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	// Улучшена обработка случая, когда запись не найдена для удаления
	if result.RowsAffected == 0 {
		logger.Warn("Попытка удаления пользователя, но нет затронутых записей (пользователь не найден)")
		return ErrUserNotFound
	}

	logger.Info("Пользователь успешно удален (мягкое удаление)")
	return nil
}

// GetByID извлекает пользователя по его ID
func (r *GormUserRepository) GetByID(ctx context.Context, id uint) (*user_model.User, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetByID").WithField("user_id", id)
	var user user_model.User

	if id == 0 {
		logger.Warn("Попытка получить пользователя с нулевым ID")
		return nil, ErrUserNotFound
	}

	result := r.db.WithContext(ctx).First(&user, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("Пользователь не найден по ID")
			return nil, ErrUserNotFound
		}
		logger.WithError(result.Error).Error("Не удалось получить пользователя по ID")
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.Info("Пользователь успешно получен по ID")
	return &user, nil
}

// GetByEmail извлекает пользователя по его адресу электронной почты
func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetByEmail").WithField("email", email)
	var user user_model.User

	if email == "" {
		logger.Warn("Попытка получить пользователя с пустым адресом электронной почты")
		return nil, ErrUserNotFound
	}

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("Пользователь не найден по адресу электронной почты")
			return nil, ErrUserNotFound
		}
		logger.WithError(result.Error).Error("Не удалось получить пользователя по адресу электронной почты")
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.Info("Пользователь успешно получен по адресу электронной почты")
	return &user, nil
}

// GetAll извлекает постраничный список пользователей с необязательными фильтрами
func (r *GormUserRepository) GetAll(
	ctx context.Context,
	params ListQueryParams,
) ([]user_model.User, int64, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetAll")
	var users []user_model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user_model.User{})

	// Применение фильтров на основе структуры ListQueryParams
	if params.MinAge != nil && *params.MinAge > 0 {
		query = query.Where("age >= ?", *params.MinAge)
		logger.Debugf("Применение фильтра: age >= %d", *params.MinAge)
	}
	if params.MaxAge != nil && *params.MaxAge > 0 {
		query = query.Where("age <= ?", *params.MaxAge)
		logger.Debugf("Применение фильтра: age <= %d", *params.MaxAge)
	}
	if params.Name != nil && *params.Name != "" {
		query = query.Where("name LIKE ?", "%"+*params.Name+"%")
		logger.Debugf("Применение фильтра: name LIKE %%%s%%", *params.Name)
	}

	// Подсчет общего количества записей, соответствующих фильтрам
	// Создаем копию запроса перед применением Limit/Offset для подсчета общего количества.
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		logger.WithError(err).Error("Не удалось подсчитать количество пользователей")
		return nil, 0, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	logger.WithFields(logrus.Fields{
		"filters_applied":                 params,
		"total_records_before_pagination": total,
	}).Debug("Подсчитано количество пользователей, соответствующих фильтрам")

	// Применение пагинации с базовой проверкой параметров
	page := params.Page
	limit := params.Limit

	if page <= 0 {
		page = 1
		logger.Warn("Предоставлен неверный номер страницы, по умолчанию используется страница 1")
	}
	if limit <= 0 {
		limit = 10 // Установка значения по умолчанию, если limit <= 0
		logger.Warnf("Предоставлен неверный лимит, по умолчанию используется лимит %d", limit)
	}

	offset := (page - 1) * limit
	result := query.Offset(offset).Limit(limit).Find(&users)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Не удалось получить постраничный список пользователей")
		return nil, 0, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.WithFields(logrus.Fields{
		"page":            page,
		"limit":           limit,
		"offset":          offset,
		"retrieved_count": len(users),
		"total_count":     total, // общее количество до пагинации
	}).Info("Пользователи успешно получены")

	return users, total, nil
}
