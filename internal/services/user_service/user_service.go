package user_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/repository/user_rep" // Используем рефакторингованный репозиторий
	"github.com/IlyushinDM/user-order-api/internal/utils/jwt_util"
	"github.com/IlyushinDM/user-order-api/internal/utils/password_util"
	"github.com/sirupsen/logrus"
	// Возможно, still needed for errors.Is() check on gorm.ErrRecordNotFound if propagated from repo
)

// **Определение ошибок сервисного слоя**
// Эти ошибки возвращаются вызывающему коду сервиса. Они могут быть
// более высокоуровневыми, чем ошибки репозитория.
var (
	// ErrUserNotFound возвращается, когда пользователь не найден.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists возвращается при попытке создать пользователя с уже существующим email.
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	// ErrEmailAlreadyTaken возвращается при попытке обновить email на уже занятый другим пользователем.
	ErrEmailAlreadyTaken = errors.New("email already taken by another user")
	// ErrInvalidCredentials возвращается при неудачной попытке входа (неверный email/пароль).
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInternalServiceError возвращается при внутренних ошибках сервиса (не связанных напрямую с БД или входными данными).
	ErrInternalServiceError = errors.New("internal service error")
	// ErrServiceDatabaseError возвращается при ошибках, возникших при взаимодействии с репозиторием.
	ErrServiceDatabaseError = errors.New("service database error")
	// ErrNoUpdateFields возвращается, если при обновлении не были предоставлены поля для изменения.
	ErrNoUpdateFields = errors.New("no fields to update")
	// ErrInvalidServiceInput возвращается, если входные данные для метода сервиса недопустимы.
	ErrInvalidServiceInput = errors.New("invalid service input")
)

// UserService defines the interface for user business logic.
type UserService interface {
	CreateUser(ctx context.Context, req user_model.CreateUserRequest) (*user_model.User, error)
	UpdateUser(ctx context.Context, id uint, req user_model.UpdateUserRequest) (*user_model.User, error)
	DeleteUser(ctx context.Context, id uint) error
	GetUserByID(ctx context.Context, id uint) (*user_model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user_model.User, error) // Needed for auth
	// Сигнатура осталась прежней для гибкости API сервиса, преобразование будет внутри метода.
	GetAllUsers(ctx context.Context, page, limit int, filters map[string]any) ([]user_model.User, int64, error)
	LoginUser(ctx context.Context, req user_model.LoginRequest) (string, error)
}

type userService struct {
	userRepo  user_rep.UserRepository
	log       *logrus.Logger
	jwtSecret string
	jwtExpSec int
}

// NewUserService creates a new user service.
// **Улучшена обработка nil зависимостей.**
func NewUserService(repo user_rep.UserRepository, log *logrus.Logger, jwtSecret string, jwtExp int) UserService {
	if repo == nil {
		logrus.Fatal("UserRepository instance is nil in NewUserService")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Logrus logger instance is nil in NewUserService, using default logger")
		log = defaultLog
	}
	// Базовая проверка секретов JWT, хотя их валидность лучше проверять при загрузке конфигурации.
	if jwtSecret == "" {
		log.Warn("JWT secret is empty in NewUserService")
	}
	if jwtExp <= 0 {
		log.Warn("JWT expiration is not set or invalid (<= 0) in NewUserService")
	}

	return &userService{userRepo: repo, log: log, jwtSecret: jwtSecret, jwtExpSec: jwtExp}
}

// CreateUser creates a new user after checking for existence and hashing the password.
func (s *userService) CreateUser(ctx context.Context, req user_model.CreateUserRequest) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.CreateUser").WithField("email", req.Email)

	// **Базовая валидация входных данных сервиса**
	if req.Email == "" || req.Password == "" || req.Name == "" {
		logger.Warn("Invalid input for user creation")
		return nil, ErrInvalidServiceInput // Возвращаем ошибку сервисного слоя
	}
	// Добавьте здесь валидацию email формата, сложности пароля и т.д.

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	// **Обработка ошибок репозитория с использованием пользовательских ошибок.**
	if err != nil && !errors.Is(err, user_rep.ErrUserNotFound) {
		// Если это не ошибка "не найдено" и не nil, это ошибка БД.
		logger.WithError(err).Error("Error checking for existing user by email")
		// Оборачиваем ошибку репозитория в сервисную ошибку БД.
		return nil, fmt.Errorf("%w: %v", ErrServiceDatabaseError, err)
	}
	if existingUser != nil {
		// Пользователь найден, значит email уже занят.
		logger.Warn("User creation attempted with existing email")
		return nil, ErrUserAlreadyExists // Возвращаем специфическую сервисную ошибку
	}

	// Hash password
	hashedPassword, err := password_util.HashPassword(req.Password)
	if err != nil {
		logger.WithError(err).Error("Failed to hash password")
		// Оборачиваем ошибку утилиты в сервисную внутреннюю ошибку.
		return nil, fmt.Errorf("%w: failed to process password", ErrInternalServiceError)
	}

	user := &user_model.User{
		Name:         req.Name,
		Email:        req.Email,
		Age:          req.Age, // Валидация Age >= 0 может быть здесь или в модели/репозитории.
		PasswordHash: hashedPassword,
	}

	// Create user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.WithError(err).Error("Failed to create user in repository")
		// Репозиторий уже оборачивает ошибки GORM. Мы просто оборачиваем ошибку репозитория
		// в сервисную ошибку БД.
		return nil, fmt.Errorf("%w: failed to save user via repository", err) // %w здесь сохраняет цепочку user_rep.ErrDatabaseError
	}

	logger.WithField("user_id", user.ID).Info("User created successfully")
	return user, nil
}

// UpdateUser updates an existing user based on the provided request.
func (s *userService) UpdateUser(
	ctx context.Context,
	id uint,
	req user_model.UpdateUserRequest,
) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.UpdateUser").WithField("user_id", id)

	// **Базовая валидация ID**
	if id == 0 {
		logger.Warn("Attempted to update user with zero ID")
		return nil, fmt.Errorf("%w: user ID must be positive", ErrInvalidServiceInput)
	}

	// Get the existing user
	user, err := s.userRepo.GetByID(ctx, id)
	// **Обработка ошибок репозитория.**
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("Update failed: User not found in repository")
			return nil, ErrUserNotFound // Возвращаем сервисную ошибку "не найдено"
		}
		logger.WithError(err).Error("Failed to get user for update from repository")
		// Оборачиваем другие ошибки репозитория.
		return nil, fmt.Errorf("%w: database error finding user for update", err) // %w здесь сохраняет цепочку user_rep.ErrDatabaseError
	}

	updated := false
	if req.Name != "" && req.Name != user.Name {
		user.Name = req.Name
		updated = true
		logger.Debug("Updating user name")
	}
	if req.Age > 0 && req.Age != user.Age { // Возможно, стоит разрешить Age = 0
		user.Age = req.Age
		updated = true
		logger.Debug("Updating user age")
	}
	if req.Email != "" && req.Email != user.Email {
		// Check if the new email is already taken by another user
		existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
		// **Обработка ошибок репозитория.**
		if err != nil && !errors.Is(err, user_rep.ErrUserNotFound) {
			logger.WithError(err).Error("Error checking for existing email during update")
			// Оборачиваем другие ошибки репозитория.
			return nil, fmt.Errorf("%w: database error checking email uniqueness during update", err) // %w сохраняет цепочку user_rep.ErrDatabaseError
		}
		// Если пользователь найден И его ID отличается от текущего обновляемого пользователя.
		if existingUser != nil && existingUser.ID != id {
			logger.Warn("Update failed: Email already taken by another user")
			return nil, ErrEmailAlreadyTaken // Возвращаем специфическую сервисную ошибку
		}
		user.Email = req.Email
		updated = true
		logger.Debug("Updating user email")
	}

	if !updated {
		logger.Info("No fields to update for user")
		return user, ErrNoUpdateFields // Возвращаем специфическую сервисную ошибку
	}

	// Perform the update in the repository
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.WithError(err).Error("Failed to update user in repository")
		// Репозиторий возвращает ErrNoRowsAffected, если запись не найдена ИЛИ не было изменений (в зависимости от реализации Update).
		// Наш репозиторий возвращает ErrNoRowsAffected только если RowsAffected == 0,
		// что после GetByID и проверки `updated` маловероятно, но возможно (например, конкурентное удаление).
		// Также он может вернуть ErrDatabaseError.
		if errors.Is(err, user_rep.ErrNoRowsAffected) {
			// Это может произойти при конкурентном удалении после GetByID, но до Update.
			logger.Warn("Update failed: User not found or no changes during repository update")
			return nil, ErrUserNotFound // Маппим к "не найдено", как наиболее вероятная причина 0 затронутых строк после успешного GetByID.
		}
		// Оборачиваем другие ошибки репозитория (вероятно, ErrDatabaseError).
		return nil, fmt.Errorf("%w: failed to save updated user via repository", err) // %w сохраняет цепочку user_rep.ErrDatabaseError
	}

	logger.Info("User updated successfully")
	return user, nil
}

// DeleteUser deletes a user by ID.
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.DeleteUser").WithField("user_id", id)

	// **Базовая валидация ID**
	if id == 0 {
		logger.Warn("Attempted to delete user with zero ID")
		return fmt.Errorf("%w: user ID must be positive", ErrInvalidServiceInput)
	}

	// Delete user in repository
	err := s.userRepo.Delete(ctx, id)
	// **Обработка ошибок репозитория.**
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("Deletion failed: User not found in repository")
			return ErrUserNotFound // Возвращаем сервисную ошибку "не найдено"
		}
		// Оборачиваем другие ошибки репозитория (вероятно, ErrDatabaseError).
		logger.WithError(err).Error("Failed to delete user in repository")
		return fmt.Errorf("%w: failed to delete user via repository", err) // %w сохраняет цепочку user_rep.ErrDatabaseError
	}
	logger.Info("User deleted successfully")
	return nil
}

// GetUserByID retrieves a user by ID.
func (s *userService) GetUserByID(ctx context.Context, id uint) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetUserByID").WithField("user_id", id)

	// **Базовая валидация ID**
	if id == 0 {
		logger.Warn("Attempted to get user with zero ID")
		// Возвращаем "не найдено", т.к. нулевой ID не соответствует реальному пользователю.
		return nil, ErrUserNotFound
	}

	// Get user from repository
	user, err := s.userRepo.GetByID(ctx, id)
	// **Обработка ошибок репозитория.**
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("User not found by ID in repository")
			return nil, ErrUserNotFound // Возвращаем сервисную ошибку "не найдено"
		}
		// Оборачиваем другие ошибки репозитория (вероятно, ErrDatabaseError).
		logger.WithError(err).Error("Failed to get user from repository")
		return nil, fmt.Errorf("%w: failed to get user by ID via repository", err) // %w сохраняет цепочку user_rep.ErrDatabaseError
	}
	logger.Info("User retrieved successfully by ID")
	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetUserByEmail").WithField("email", email)

	// **Базовая валидация email**
	if email == "" {
		logger.Warn("Attempted to get user with empty email")
		// Возвращаем "не найдено", т.к. пустой email не соответствует реальному пользователю.
		return nil, ErrUserNotFound
	}
	// Добавьте здесь валидацию формата email.

	// Get user from repository
	user, err := s.userRepo.GetByEmail(ctx, email)
	// **Обработка ошибок репозитория.**
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("User not found by email in repository")
			return nil, ErrUserNotFound // Возвращаем сервисную ошибку "не найдено"
		}
		// Оборачиваем другие ошибки репозитория (вероятно, ErrDatabaseError).
		logger.WithError(err).Error("Failed to get user by email from repository")
		return nil, fmt.Errorf("%w: failed to get user by email via repository", err) // %w сохраняет цепочку user_rep.ErrDatabaseError
	}
	logger.Info("User retrieved successfully by email")
	return user, nil
}

// GetAllUsers retrieves a paginated list of users with optional filters.
func (s *userService) GetAllUsers(
	ctx context.Context,
	page,
	limit int,
	filters map[string]interface{},
) ([]user_model.User, int64, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetAllUsers")

	// **Базовая валидация пагинации**
	if page <= 0 {
		page = 1 // Установка значения по умолчанию, если page <= 0
		logger.Warn("Invalid page number provided, defaulting to page 1")
	}
	if limit <= 0 {
		limit = 10 // Установка значения по умолчанию, если limit <= 0
		logger.Warnf("Invalid limit provided, defaulting to limit %d", limit)
	}

	// **Преобразование map[string]interface{} в user_rep.ListQueryParams**
	queryParams := user_rep.ListQueryParams{
		Page:  page,
		Limit: limit,
	}

	// Безопасное извлечение фильтров из map с проверкой типа
	if minAge, ok := filters["min_age"].(int); ok {
		queryParams.MinAge = &minAge
		logger.Debugf("Applying filter param: min_age = %d", minAge)
	}
	if maxAge, ok := filters["max_age"].(int); ok {
		queryParams.MaxAge = &maxAge
		logger.Debugf("Applying filter param: max_age = %d", maxAge)
	}
	if name, ok := filters["name"].(string); ok {
		queryParams.Name = &name
		logger.Debugf("Applying filter param: name = %s", name)
	}
	// Добавьте преобразование других фильтров здесь.

	// Get users from repository using the new query params struct
	users, total, err := s.userRepo.GetAll(ctx, queryParams)
	// **Обработка ошибок репозитория.**
	if err != nil {
		logger.WithError(err).Error("Failed to get all users from repository")
		// Оборачиваем ошибку репозитория (вероятно, ErrDatabaseError).
		return nil, 0, fmt.Errorf("%w: failed to get all users via repository", err) // %w сохраняет цепочку user_rep.ErrDatabaseError
	}

	logger.WithFields(logrus.Fields{"count": len(users), "total": total, "page": page, "limit": limit, "filters": filters}).Info("Retrieved all users successfully")
	return users, total, nil
}

// LoginUser authenticates a user and generates a JWT token.
func (s *userService) LoginUser(ctx context.Context, req user_model.LoginRequest) (string, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.LoginUser").WithField("email", req.Email)

	// **Базовая валидация входных данных сервиса**
	if req.Email == "" || req.Password == "" {
		logger.Warn("Invalid input for login")
		return "", ErrInvalidServiceInput
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	// **Обработка ошибок репозитория.**
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("Login attempt failed: User not found in repository")
			// Возвращаем generic "неверные учетные данные" для безопасности (предотвращение перечисления пользователей).
			return "", ErrInvalidCredentials
		}
		// Оборачиваем другие ошибки репозитория (вероятно, ErrDatabaseError).
		logger.WithError(err).Error("Database error during login attempt")
		return "", fmt.Errorf("%w: database error finding user for login", err) // %w сохраняет цепочку user_rep.ErrDatabaseError
	}

	// Check password
	if !password_util.CheckPasswordHash(req.Password, user.PasswordHash) {
		logger.Warn("Login attempt failed: Invalid password")
		// Возвращаем generic "неверные учетные данные".
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := jwt_util.GenerateJWT(user.ID, user.Email, s.jwtSecret, s.jwtExpSec)
	if err != nil {
		logger.WithError(err).Error("Failed to generate JWT token")
		// Оборачиваем ошибку утилиты в сервисную внутреннюю ошибку.
		return "", fmt.Errorf("%w: failed to generate authentication token", ErrInternalServiceError)
	}

	logger.WithField("user_id", user.ID).Info("User logged in successfully")
	return token, nil
}
