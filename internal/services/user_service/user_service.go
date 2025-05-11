package user_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/repository/user_rep"
	"github.com/IlyushinDM/user-order-api/internal/utils/jwt_util"
	"github.com/IlyushinDM/user-order-api/internal/utils/password_util"
	"github.com/sirupsen/logrus"
)

// Определение ошибок сервисного слоя
var (
	ErrUserNotFound         = errors.New("пользователь не найден")
	ErrUserAlreadyExists    = errors.New("попытка создать пользователя с уже существующим email")
	ErrEmailAlreadyTaken    = errors.New("попытка обновить email на уже занятый другим пользователем")
	ErrInvalidCredentials   = errors.New("неверный email или пароль")
	ErrInternalServiceError = errors.New("внутренняя ошибка сервиса")
	ErrServiceDatabaseError = errors.New("ошибка при взаимодействии с репозиторием")
	ErrNoUpdateFields       = errors.New("не были предоставлены поля для изменения")
	ErrInvalidServiceInput  = errors.New("входные данные для метода сервиса недопустимы")
)

// UserService определяет интерфейс для бизнес-логики пользователей.
type UserService interface {
	CreateUser(ctx context.Context, req user_model.CreateUserRequest) (*user_model.User, error)
	UpdateUser(ctx context.Context, id uint, req user_model.UpdateUserRequest) (*user_model.User, error)
	DeleteUser(ctx context.Context, id uint) error
	GetUserByID(ctx context.Context, id uint) (*user_model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user_model.User, error)
	GetAllUsers(ctx context.Context, page, limit int, filters map[string]any) ([]user_model.User, int64, error)
	LoginUser(ctx context.Context, req user_model.LoginRequest) (string, error)
}

type userService struct {
	userRepo  user_rep.UserRepository
	log       *logrus.Logger
	jwtSecret string
	jwtExpSec int
}

// NewUserService создает новый сервис пользователей
func NewUserService(repo user_rep.UserRepository, log *logrus.Logger, jwtSecret string, jwtExp int) UserService {
	if repo == nil {
		logrus.Fatal("Экземпляр UserRepository равен nil в NewUserService")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Экземпляр логгера Logrus равен nil в NewUserService, используется логгер по умолчанию")
		log = defaultLog
	}
	if jwtSecret == "" {
		log.Warn("Секрет JWT пуст в NewUserService")
	}
	if jwtExp <= 0 {
		log.Warn("Срок действия JWT не установлен или некорректен (<= 0) в NewUserService")
	}

	return &userService{userRepo: repo, log: log, jwtSecret: jwtSecret, jwtExpSec: jwtExp}
}

// CreateUser создает нового пользователя после проверки на существование и хеширования пароля
func (s *userService) CreateUser(ctx context.Context, req user_model.CreateUserRequest) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.CreateUser").WithField("email", req.Email)

	if req.Email == "" || req.Password == "" || req.Name == "" {
		logger.Warn("Недопустимые входные данные для создания пользователя")
		return nil, ErrInvalidServiceInput
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, user_rep.ErrUserNotFound) {
		logger.WithError(err).Error("Ошибка при проверке существования пользователя по email")
		return nil, fmt.Errorf("%w: ошибка базы данных при поиске существующего пользователя", err)
	}
	if existingUser != nil {
		logger.Warn("Попытка создания пользователя с уже существующим email")
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := password_util.HashPassword(req.Password)
	if err != nil {
		logger.WithError(err).Error("Не удалось хешировать пароль")
		return nil, fmt.Errorf("%w: не удалось обработать пароль", ErrInternalServiceError)
	}

	user := &user_model.User{
		Name:         req.Name,
		Email:        req.Email,
		Age:          req.Age,
		PasswordHash: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.WithError(err).Error("Не удалось создать пользователя в репозитории")
		return nil, fmt.Errorf("%w: не удалось сохранить пользователя через репозиторий", err)
	}

	logger.WithField("user_id", user.ID).Info("Пользователь успешно создан")
	return user, nil
}

// UpdateUser обновляет существующего пользователя на основе предоставленного запроса
func (s *userService) UpdateUser(
	ctx context.Context,
	id uint,
	req user_model.UpdateUserRequest,
) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.UpdateUser").WithField("user_id", id)

	// Базовая валидация ID
	if id == 0 {
		logger.Warn("Попытка обновить пользователя с нулевым ID")
		return nil, fmt.Errorf("%w: ID пользователя должен быть положительным числом", ErrInvalidServiceInput)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	// Обработка ошибок репозитория
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("Обновление не удалось: Пользователь не найден в репозитории")
			return nil, ErrUserNotFound
		}
		logger.WithError(err).Error("Не удалось получить пользователя для обновления из репозитория")
		return nil, fmt.Errorf("%w: ошибка базы данных при поиске пользователя для обновления", err)
	}

	updated := false
	if req.Name != "" && req.Name != user.Name {
		user.Name = req.Name
		updated = true
		logger.Debug("Обновление имени пользователя")
	}
	if req.Age > 0 && req.Age != user.Age {
		user.Age = req.Age
		updated = true
		logger.Debug("Обновление возраста пользователя")
	}
	if req.Email != "" && req.Email != user.Email {
		existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
		// Обработка ошибок репозитория
		if err != nil && !errors.Is(err, user_rep.ErrUserNotFound) {
			logger.WithError(err).Error("Ошибка при проверке существования email во время обновления")
			return nil, fmt.Errorf("%w: ошибка базы данных при проверке уникальности email во время обновления", err)
		}
		if existingUser != nil && existingUser.ID != id {
			logger.Warn("Обновление не удалось: Email уже занят другим пользователем")
			return nil, ErrEmailAlreadyTaken
		}
		user.Email = req.Email
		updated = true
		logger.Debug("Обновление email пользователя")
	}

	if !updated {
		logger.Info("Нет полей для обновления у пользователя")
		return user, ErrNoUpdateFields
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.WithError(err).Error("Не удалось обновить пользователя в репозитории")
		if errors.Is(err, user_rep.ErrNoRowsAffected) {
			logger.Warn("Обновление не удалось: Пользователь не найден или нет изменений при обновлении в репозитории")
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: не удалось сохранить обновленного пользователя через репозиторий", err)
	}

	logger.Info("Пользователь успешно обновлен")
	return user, nil
}

// DeleteUser удаляет пользователя по ID
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.DeleteUser").WithField("user_id", id)

	// Базовая валидация ID
	if id == 0 {
		logger.Warn("Попытка удалить пользователя с нулевым ID")
		return fmt.Errorf("%w: ID пользователя должен быть положительным числом", ErrInvalidServiceInput)
	}

	err := s.userRepo.Delete(ctx, id)
	// Обработка ошибок репозитория
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("Удаление не удалось: Пользователь не найден в репозитории")
			return ErrUserNotFound
		}
		logger.WithError(err).Error("Не удалось удалить пользователя в репозитории")
		return fmt.Errorf("%w: не удалось удалить пользователя через репозиторий", err)
	}
	logger.Info("Пользователь успешно удален")
	return nil
}

// GetUserByID получает пользователя по ID
func (s *userService) GetUserByID(ctx context.Context, id uint) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetUserByID").WithField("user_id", id)

	// Базовая валидация ID
	if id == 0 {
		logger.Warn("Попытка получить пользователя с нулевым ID")
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, id)
	// Обработка ошибок репозитория
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("Пользователь не найден по ID в репозитории")
			return nil, ErrUserNotFound
		}
		logger.WithError(err).Error("Не удалось получить пользователя из репозитория")
		return nil, fmt.Errorf("%w: не удалось получить пользователя по ID через репозиторий", err)
	}
	logger.Info("Пользователь успешно получен по ID")
	return user, nil
}

// GetAllUsers получает пагинированный список пользователей с опциональными фильтрами.
func (s *userService) GetAllUsers(
	ctx context.Context,
	page,
	limit int,
	filters map[string]interface{},
) ([]user_model.User, int64, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetAllUsers")

	// Базовая валидация пагинации
	if page <= 0 {
		page = 1 // Установка значения по умолчанию, если page <= 0
		logger.Warn("Предоставлен некорректный номер страницы, используется страница 1 по умолчанию")
	}
	if limit <= 0 {
		limit = 10 // Установка значения по умолчанию, если limit <= 0
		logger.Warnf("Предоставлен некорректный лимит, используется лимит %d по умолчанию", limit)
	}

	queryParams := user_rep.ListQueryParams{
		Page:  page,
		Limit: limit,
	}

	if minAge, ok := filters["min_age"].(int); ok {
		queryParams.MinAge = &minAge
		logger.Debugf("Применение параметра фильтра: min_age = %d", minAge)
	}
	if maxAge, ok := filters["max_age"].(int); ok {
		queryParams.MaxAge = &maxAge
		logger.Debugf("Применение параметра фильтра: max_age = %d", maxAge)
	}
	if name, ok := filters["name"].(string); ok {
		queryParams.Name = &name
		logger.Debugf("Применение параметра фильтра: name = %s", name)
	}

	users, total, err := s.userRepo.GetAll(ctx, queryParams)
	// Обработка ошибок репозитория.
	if err != nil {
		logger.WithError(err).Error("Не удалось получить всех пользователей из репозитория")
		return nil, 0, fmt.Errorf("%w: не удалось получить всех пользователей через репозиторий", err)
	}

	logger.WithFields(logrus.Fields{"count": len(users), "total": total, "page": page, "limit": limit, "filters": filters}).Info("Все пользователи успешно получены")
	return users, total, nil
}

// LoginUser аутентифицирует пользователя и генерирует JWT токен
func (s *userService) LoginUser(ctx context.Context, req user_model.LoginRequest) (string, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.LoginUser").WithField("email", req.Email)

	// Базовая валидация входных данных сервиса
	if req.Email == "" || req.Password == "" {
		logger.Warn("Недопустимые входные данные для входа")
		return "", ErrInvalidServiceInput
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	// Обработка ошибок репозитория
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("Попытка входа не удалась: Пользователь не найден в репозитории")
			return "", ErrInvalidCredentials
		}
		logger.WithError(err).Error("Ошибка базы данных при попытке входа")
		return "", fmt.Errorf("%w: ошибка базы данных при поиске пользователя для входа", err)
	}

	if !password_util.CheckPasswordHash(req.Password, user.PasswordHash) {
		logger.Warn("Попытка входа не удалась: Неверный пароль")
		return "", ErrInvalidCredentials
	}

	token, err := jwt_util.GenerateJWT(user.ID, user.Email, s.jwtSecret, s.jwtExpSec)
	if err != nil {
		logger.WithError(err).Error("Не удалось сгенерировать JWT токен")
		return "", fmt.Errorf("%w: не удалось сгенерировать токен аутентификации", ErrInternalServiceError)
	}

	logger.WithField("user_id", user.ID).Info("Пользователь успешно вошел в систему")
	return token, nil
}

// GetUserByEmail получает пользователя по Email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*user_model.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetUserByEmail").WithField("email", email)

	// Базовая валидация email
	if email == "" {
		logger.Warn("Attempted to get user with empty email")
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	// Обработка ошибок репозитория
	if err != nil {
		if errors.Is(err, user_rep.ErrUserNotFound) {
			logger.Warn("User not found by email in repository")
			return nil, ErrUserNotFound
		}
		logger.WithError(err).Error("Failed to get user by email from repository")
		return nil, fmt.Errorf("%w: failed to get user by email via repository", err)
	}
	logger.Info("User retrieved successfully by email")
	return user, nil
}
