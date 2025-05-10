package user_rep

import (
	"context"
	"errors"
	"fmt" // Добавлено для форматирования ошибок

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// **Определение пользовательских ошибок**
// Использование пользовательских ошибок позволяет вызывающему коду более гранулярно обрабатывать specific ситуации
// и делает код менее зависимым от конкретной реализации ORM (GORM).
var (
	// ErrUserNotFound возвращается, когда пользователь не найден по ID или Email.
	ErrUserNotFound = errors.New("user not found")
	// ErrDatabaseError возвращается при возникновении общей ошибки базы данных.
	ErrDatabaseError = errors.New("database operation failed")
	// ErrNoRowsAffected возвращается операциями Update или Delete, когда ни одна запись не была изменена/удалена.
	ErrNoRowsAffected = errors.New("no rows affected")
	// ErrInvalidInput возвращается, если входные параметры репозитория недопустимы (например, нулевой ID).
	ErrInvalidInput = errors.New("invalid input parameter")
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *user_model.User) error
	Update(ctx context.Context, user *user_model.User) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*user_model.User, error)
	GetByEmail(ctx context.Context, email string) (*user_model.User, error)
	// Изменена сигнатура GetAll для использования типобезопасной структуры фильтров.
	GetAll(ctx context.Context, params ListQueryParams) ([]user_model.User, int64, error)
}

// **Структура ListQueryParams для типобезопасных фильтров GetAll**
// Использование структуры вместо map[string]any делает сигнатуру метода GetAll более явной,
// улучшает типобезопасность и облегчает проверку входных параметров на этапе компиляции.
type ListQueryParams struct {
	Page  int
	Limit int
	// Фильтры (опционально) - использование указателей позволяет явно различать
	// отсутствие фильтра (nil) и фильтр со значением по умолчанию (например, 0).
	MinAge *int
	MaxAge *int
	Name   *string
	// Добавьте другие поля для фильтрации здесь по мере необходимости.
}

// GormUserRepository implements UserRepository using GORM.
type GormUserRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewGormUserRepository creates a new user repository using GORM.
// **Улучшена обработка nil зависимостей.**
func NewGormUserRepository(db *gorm.DB, log *logrus.Logger) UserRepository {
	if db == nil {
		// Если экземпляр DB не предоставлен, это критическая ошибка.
		// log.Fatal завершит работу приложения.
		logrus.Fatal("GORM DB instance is nil")
	}
	if log == nil {
		// Если логгер не предоставлен, можно использовать логгер по умолчанию с предупреждением.
		// Это позволяет репозиторию работать даже без внешнего логгера.
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Logrus logger instance is nil in NewGormUserRepository, using default logger")
		log = defaultLog
	}
	return &GormUserRepository{db: db, log: log}
}

// Create creates a new user record in the database.
func (r *GormUserRepository) Create(ctx context.Context, user *user_model.User) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Create")

	// **Добавлена проверка входных данных.**
	if user == nil {
		logger.Error("Attempted to create a nil user")
		return fmt.Errorf("%w: user object is nil", ErrInvalidInput) // Используем ErrInvalidInput
	}

	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to create user")
		// **Оборачиваем ошибку GORM в пользовательскую ошибку базы данных.**
		// Это сохраняет оригинальную ошибку GORM, но позволяет вызывающему коду
		// проверить, является ли ошибка общей ошибкой базы данных с помощью errors.Is(err, ErrDatabaseError).
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	// Проверка RowsAffected в Create обычно не требуется, так как GORM возвращает ошибку,
	// если создание не удалось до изменения записей.

	logger.WithField("user_id", user.ID).Info("User created successfully")
	return nil
}

// Update updates an existing user record.
// **Уточнен комментарий про GORM.Updates и добавлена проверка ID.**
// Note: GORM's Updates method only updates non-zero/non-empty fields by default.
// If you need to update zero values (e.g., set an integer field to 0 or a boolean to false),
// you must use db.Model(&user).Select("FieldName1", "FieldName2").Updates(user)
// or db.Save(user) after fetching the user (be cautious with Save as it replaces the entire record).
func (r *GormUserRepository) Update(ctx context.Context, user *user_model.User) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Update").WithField("user_id", user.ID)

	// **Добавлена проверка входных данных.**
	if user == nil {
		logger.Error("Attempted to update a nil user")
		return fmt.Errorf("%w: user object is nil", ErrInvalidInput)
	}
	if user.ID == 0 {
		logger.Error("Attempted to update user with zero ID")
		return fmt.Errorf("%w: user ID is zero, cannot update", ErrInvalidInput)
	}

	// Model(user) scopes the update to the user with the given ID.
	// Updates(user) updates non-zero fields from the user object.
	result := r.db.WithContext(ctx).Model(user).Updates(user)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to update user")
		// **Оборачиваем ошибку GORM.**
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	// **Улучшена обработка случая, когда ни одна запись не затронута.**
	if result.RowsAffected == 0 {
		logger.Warn("User update attempted but no rows affected (possibly user not found or no changes)")
		// Возвращаем ErrNoRowsAffected, так как мы не можем с уверенностью сказать,
		// был ли пользователь не найден или просто не было изменений.
		// Если требуется специфичная ошибка ErrUserNotFound при отсутствии записи,
		// необходимо выполнить GetByID перед Update на уровне сервисного слоя.
		return ErrNoRowsAffected
	}

	logger.Info("User updated successfully")
	return nil
}

// Delete deletes a user by ID. Performs a soft delete if the model has a DeletedAt field.
// **Добавлена проверка ID и улучшена обработка ErrRecordNotFound.**
func (r *GormUserRepository) Delete(ctx context.Context, id uint) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Delete").WithField("user_id", id)

	// **Добавлена проверка входных данных.**
	if id == 0 {
		logger.Error("Attempted to delete user with zero ID")
		return fmt.Errorf("%w: user ID is zero, cannot delete", ErrInvalidInput)
	}

	// GORM's default Delete performs a soft delete if the model has DeletedAt field
	result := r.db.WithContext(ctx).Delete(&user_model.User{}, id)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to delete user")
		// **Оборачиваем ошибку GORM.**
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	// **Улучшена обработка случая, когда запись не найдена для удаления.**
	if result.RowsAffected == 0 {
		logger.Warn("User deletion attempted but no rows affected (user not found)")
		// Явно возвращаем ErrUserNotFound, так как это наиболее вероятная причина 0 RowsAffected при удалении по ID.
		return ErrUserNotFound
	}

	logger.Info("User deleted successfully (soft delete)")
	return nil
}

// GetByID retrieves a user by their ID.
// **Улучшена обработка ErrRecordNotFound с использованием errors.Is.**
func (r *GormUserRepository) GetByID(ctx context.Context, id uint) (*user_model.User, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetByID").WithField("user_id", id)
	var user user_model.User

	// **Добавлена проверка входных данных.**
	if id == 0 {
		logger.Warn("Attempted to get user with zero ID")
		// Возвращаем ErrUserNotFound, так как нулевой ID не может соответствовать существующей записи в большинстве баз данных.
		return nil, ErrUserNotFound
	}

	// Use Preload to load associated orders if needed, e.g., r.db.WithContext(ctx).Preload("Orders").First(&user, id)
	result := r.db.WithContext(ctx).First(&user, id)

	if result.Error != nil {
		// **Явная проверка на ErrRecordNotFound с использованием errors.Is.**
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("User not found by ID")
			return nil, ErrUserNotFound // Возвращаем пользовательскую ошибку
		}
		logger.WithError(result.Error).Error("Failed to get user by ID")
		// **Оборачиваем другие ошибки GORM.**
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.Info("User retrieved successfully by ID")
	return &user, nil
}

// GetByEmail retrieves a user by their email address.
// **Улучшена обработка ErrRecordNotFound с использованием errors.Is и добавлена проверка email.**
func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetByEmail").WithField("email", email)
	var user user_model.User

	// **Добавлена проверка входных данных.**
	if email == "" {
		logger.Warn("Attempted to get user with empty email")
		// Возвращаем ErrUserNotFound, так как пустой email не может соответствовать существующей записи.
		return nil, ErrUserNotFound
	}

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)

	if result.Error != nil {
		// **Явная проверка на ErrRecordNotFound с использованием errors.Is.**
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("User not found by email")
			return nil, ErrUserNotFound // Возвращаем пользовательскую ошибку
		}
		logger.WithError(result.Error).Error("Failed to get user by email")
		// **Оборачиваем другие ошибки GORM.**
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.Info("User retrieved successfully by email")
	return &user, nil
}

// GetAll retrieves a paginated list of users with optional filters.
// **Изменена сигнатура для использования ListQueryParams и обновлена логика фильтрации.**
func (r *GormUserRepository) GetAll(
	ctx context.Context,
	params ListQueryParams) ([]user_model.User, int64, error) {

	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetAll")
	var users []user_model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user_model.User{})

	// **Применение фильтров на основе структуры ListQueryParams.**
	// Проверяем, является ли указатель nil, чтобы понять, задан ли фильтр.
	if params.MinAge != nil && *params.MinAge > 0 {
		query = query.Where("age >= ?", *params.MinAge)
		logger.Debugf("Applying filter: age >= %d", *params.MinAge)
	}
	if params.MaxAge != nil && *params.MaxAge > 0 {
		query = query.Where("age <= ?", *params.MaxAge)
		logger.Debugf("Applying filter: age <= %d", *params.MaxAge)
	}
	if params.Name != nil && *params.Name != "" {
		// Использование ILIKE для регистронезависимого поиска (часто используется в PostgreSQL).
		// Для других СУБД может потребоваться корректировка (например, использование функции lower()).
		query = query.Where("name ILIKE ?", "%"+*params.Name+"%")
		logger.Debugf("Applying filter: name ILIKE %%%s%%", *params.Name)
	}
	// Добавьте логику применения других фильтров из структуры ListQueryParams здесь.

	// Count total records matching filters
	// Создаем копию запроса перед применением Limit/Offset для подсчета общего количества.
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count users")
		// **Оборачиваем ошибку GORM.**
		return nil, 0, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	logger.WithFields(logrus.Fields{
		// Логируем саму структуру параметров для лучшей отладки.
		"filters_applied":                 params,
		"total_records_before_pagination": total,
	}).Debug("Counted users matching filters")

	// **Применение пагинации с базовой проверкой параметров.**
	page := params.Page
	limit := params.Limit

	if page <= 0 {
		page = 1
		logger.Warn("Invalid page number provided, defaulting to page 1")
	}
	if limit <= 0 {
		limit = 10 // Установка значения по умолчанию, если limit <= 0
		logger.Warnf("Invalid limit provided, defaulting to limit %d", limit)
	}

	offset := (page - 1) * limit
	result := query.Offset(offset).Limit(limit).Find(&users)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to retrieve paginated users")
		// **Оборачиваем ошибку GORM.**
		return nil, 0, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	logger.WithFields(logrus.Fields{
		"page":            page,
		"limit":           limit,
		"offset":          offset,
		"retrieved_count": len(users),
		"total_count":     total, // Это общее количество до пагинации
	}).Info("Users retrieved successfully")

	return users, total, nil
}
