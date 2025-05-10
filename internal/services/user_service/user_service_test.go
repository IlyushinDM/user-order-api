package user_service_test

import (
	"context"
	"errors"
	"io" // Импортируем io для io.Discard
	"testing"

	// Понадобится для мокирования времени в JWT тестах, если нужно, но для сервиса не обязательно
	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/repository/user_rep"
	"github.com/IlyushinDM/user-order-api/internal/services/user_service" // Тестируемый пакет

	// Предполагаем, что утилиты тестируются отдельно
	// Предполагаем, что утилиты тестируются отдельно
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Моки Зависимостей ---

// MockUserRepository мок структуры user_rep.UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Create имитация вызова репозитория для создания пользователя
func (m *MockUserRepository) Create(ctx context.Context, user *user_model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// GetByID имитация вызова репозитория для получения пользователя по ID
func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*user_model.User, error) {
	args := m.Called(ctx, id)
	ret0 := args.Get(0)
	if ret0 == nil {
		return nil, args.Error(1)
	}
	return ret0.(*user_model.User), args.Error(1)
}

// GetByEmail имитация вызова репозитория для получения пользователя по Email
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	args := m.Called(ctx, email)
	ret0 := args.Get(0)
	if ret0 == nil {
		return nil, args.Error(1)
	}
	return ret0.(*user_model.User), args.Error(1)
}

// GetAll имитация вызова репозитория для получения списка пользователей
func (m *MockUserRepository) GetAll(ctx context.Context, params user_rep.ListQueryParams) ([]user_model.User, int64, error) {
	args := m.Called(ctx, params)
	ret0 := args.Get(0)
	if ret0 == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return ret0.([]user_model.User), args.Get(1).(int64), args.Error(2)
}

// Update имитация вызова репозитория для обновления пользователя
func (m *MockUserRepository) Update(ctx context.Context, user *user_model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Delete имитация вызова репозитория для удаления пользователя
func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// --- Вспомогательные Функции для Тестов ---

// setupTestService создает экземпляр UserService с мокированными зависимостями
// и логгером, который пишет в io.Discard
func setupTestService() (*MockUserRepository, user_service.UserService, *logrus.Logger) {
	mockRepo := new(MockUserRepository)
	testLogger := logrus.New()
	testLogger.SetOutput(io.Discard)       // Отправляем вывод логгера в никуда во время тестов
	testLogger.SetLevel(logrus.PanicLevel) // Отключаем все логи, кроме паники, чтобы тесты не висели

	// Задаем тестовые значения для JWT секрета и времени жизни
	testJwtSecret := "test-secret"
	testJwtExp := 3600 // 1 час

	service := user_service.NewUserService(mockRepo, testLogger, testJwtSecret, testJwtExp)

	return mockRepo, service, testLogger
}

// helper для создания указателя на string
func strPtr(s string) *string {
	return &s
}

// helper для создания указателя на int
func intPtr(i int) *int {
	return &i
}

// --- Тесты для NewUserService ---

func TestNewUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	testLogger := logrus.New()
	testLogger.SetOutput(io.Discard)
	testLogger.SetLevel(logrus.PanicLevel)

	t.Run("Успешное создание с валидными зависимостями", func(t *testing.T) {
		service := user_service.NewUserService(mockRepo, testLogger, "secret", 3600)
		assert.NotNil(t, service, "Сервис должен быть создан")
	})

	t.Run("Паника при nil UserRepository", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Ожидалась паника при nil UserRepository")
			}
		}()
		user_service.NewUserService(nil, testLogger, "secret", 3600)
	})

	t.Run("Использование дефолтного логгера при nil Logger", func(t *testing.T) {
		// Этот тест проверяет, что не происходит паники и возвращается не nil сервис.
		// Проверить, что используется именно дефолтный логгер, сложнее без мокирования logrus.New().
		service := user_service.NewUserService(mockRepo, nil, "secret", 3600)
		assert.NotNil(t, service, "Сервис должен быть создан даже с nil логгером")
	})

	t.Run("Предупреждение при пустом JWT Secret", func(t *testing.T) {
		// Для проверки предупреждений логгера, нужно либо мокировать логгер более детально,
		// либо проверять вывод логгера (что сложнее). В данном тесте просто убеждаемся,
		// что сервис создается без паники.
		service := user_service.NewUserService(mockRepo, testLogger, "", 3600)
		assert.NotNil(t, service, "Сервис должен быть создан с пустым JWT Secret (с предупреждением)")
	})

	t.Run("Предупреждение при невалидном JWT Exp", func(t *testing.T) {
		service := user_service.NewUserService(mockRepo, testLogger, "secret", 0)
		assert.NotNil(t, service, "Сервис должен быть создан с невалидным JWT Exp (с предупреждением)")
	})
}

// --- Тесты для CreateUser ---

func TestCreateUser(t *testing.T) {
	mockRepo, service, _ := setupTestService()
	ctx := context.Background()

	req := user_model.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      30,
		Password: "password123",
	}

	t.Run("Успешное создание пользователя", func(t *testing.T) {
		// Мок репозитория: GetByEmail возвращает ErrUserNotFound (пользователя нет), Create успешен
		mockRepo.On("GetByEmail", ctx, req.Email).Return((*user_model.User)(nil), user_rep.ErrUserNotFound).Once()
		// Ожидаем вызов Create с пользователем, где PasswordHash будет установлен
		mockRepo.On("Create", ctx, mock.AnythingOfType("*user_model.User")).Return(nil).Once()

		user, err := service.CreateUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Age, user.Age)
		assert.NotEmpty(t, user.PasswordHash, "Хэш пароля должен быть установлен") // Проверяем, что хэш не пустой
		// Проверяем, что методы репозитория были вызваны
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка: пользователь уже существует", func(t *testing.T) {
		existingUser := &user_model.User{ID: 1, Email: req.Email}
		// Мок репозитория: GetByEmail находит пользователя
		mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil).Once()
		// Create не должен вызываться

		user, err := service.CreateUser(ctx, req)

		assert.ErrorIs(t, err, user_service.ErrUserAlreadyExists)
		assert.Nil(t, user)
		// Проверяем, что GetByEmail был вызван, а Create - нет
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка репозитория при проверке существования (не ErrUserNotFound)", func(t *testing.T) {
		dbError := errors.New("database connection failed")
		// Мок репозитория: GetByEmail возвращает другую ошибку
		mockRepo.On("GetByEmail", ctx, req.Email).Return((*user_model.User)(nil), dbError).Once()
		// Create не должен вызываться

		user, err := service.CreateUser(ctx, req)

		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
		assert.ErrorContains(t, err, dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка: некорректные входные данные (пустой email)", func(t *testing.T) {
		invalidReq := req
		invalidReq.Email = ""
		// Методы репозитория не должны вызываться

		user, err := service.CreateUser(ctx, invalidReq)

		assert.ErrorIs(t, err, user_service.ErrInvalidServiceInput)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t) // Проверяем, что моки не были вызваны
	})

	t.Run("Ошибка: некорректные входные данные (пустой пароль)", func(t *testing.T) {
		invalidReq := req
		invalidReq.Password = ""
		// Методы репозитория не должны вызываться

		user, err := service.CreateUser(ctx, invalidReq)

		assert.ErrorIs(t, err, user_service.ErrInvalidServiceInput)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка: некорректные входные данные (пустое имя)", func(t *testing.T) {
		invalidReq := req
		invalidReq.Name = ""
		// Методы репозитория не должны вызываться

		user, err := service.CreateUser(ctx, invalidReq)

		assert.ErrorIs(t, err, user_service.ErrInvalidServiceInput)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	// Примечание: Ошибки утилит (password_util, jwt_util) обычно сложнее мокировать,
	// так как они не являются зависимостями, передаваемыми в конструктор.
	// Предполагается, что эти утилиты тестируются отдельно.
	// Если бы password_util.HashPassword могла вернуть ошибку при валидных входных,
	// нужно было бы либо мокировать глобально (что нежелательно), либо
	// передавать хэшер как зависимость в UserService. В текущей структуре
	// мы можем полагаться на то, что HashPassword не вернет ошибку для валидных строк.

	t.Run("Ошибка репозитория при создании", func(t *testing.T) {
		dbError := errors.New("database insert failed")
		// Мок репозитория: GetByEmail возвращает ErrUserNotFound, Create возвращает ошибку
		mockRepo.On("GetByEmail", ctx, req.Email).Return((*user_model.User)(nil), user_rep.ErrUserNotFound).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*user_model.User")).Return(dbError).Once()

		user, err := service.CreateUser(ctx, req)

		// Ожидаем, что ошибка репозитория будет обернута или возвращена напрямую
		// Проверяем, что ошибка содержит исходную ошибку (согласно коду user_service)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), dbError.Error(), "Ошибка репозитория должна быть включена в возвращаемую ошибку сервиса")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

// --- Тесты для GetUserByID ---

func TestGetUserByID(t *testing.T) {
	mockRepo, service, _ := setupTestService()
	ctx := context.Background()
	userID := uint(1)
	expectedUser := &user_model.User{ID: userID, Name: "Test User", Email: "test@example.com"}

	t.Run("Успешное получение пользователя", func(t *testing.T) {
		// Мок репозитория: GetByID находит пользователя
		mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil).Once()

		user, err := service.GetUserByID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка: пользователь не найден", func(t *testing.T) {
		// Мок репозитория: GetByID возвращает ErrUserNotFound
		mockRepo.On("GetByID", ctx, userID).Return((*user_model.User)(nil), user_rep.ErrUserNotFound).Once()

		user, err := service.GetUserByID(ctx, userID)

		assert.ErrorIs(t, err, user_service.ErrUserNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка репозитория при получении (не ErrUserNotFound)", func(t *testing.T) {
		dbError := errors.New("database connection failed")
		// Мок репозитория: GetByID возвращает другую ошибку
		mockRepo.On("GetByID", ctx, userID).Return((*user_model.User)(nil), dbError).Once()

		user, err := service.GetUserByID(ctx, userID)

		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
		assert.ErrorContains(t, err, dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка: ID пользователя равен 0", func(t *testing.T) {
		// Репозиторий не должен вызываться
		user, err := service.GetUserByID(ctx, 0)

		assert.ErrorIs(t, err, user_service.ErrUserNotFound) // Сервис маппит 0 ID в Not Found
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

// --- Тесты для GetUserByEmail ---

func TestGetUserByEmail(t *testing.T) {
	mockRepo, service, _ := setupTestService()
	ctx := context.Background()
	userEmail := "test@example.com"
	expectedUser := &user_model.User{ID: 1, Name: "Test User", Email: userEmail}

	t.Run("Успешное получение пользователя по Email", func(t *testing.T) {
		// Мок репозитория: GetByEmail находит пользователя
		mockRepo.On("GetByEmail", ctx, userEmail).Return(expectedUser, nil).Once()

		user, err := service.GetUserByEmail(ctx, userEmail)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка: пользователь не найден по Email", func(t *testing.T) {
		// Мок репозитория: GetByEmail возвращает ErrUserNotFound
		mockRepo.On("GetByEmail", ctx, userEmail).Return((*user_model.User)(nil), user_rep.ErrUserNotFound).Once()

		user, err := service.GetUserByEmail(ctx, userEmail)

		assert.ErrorIs(t, err, user_service.ErrUserNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка репозитория при получении по Email (не ErrUserNotFound)", func(t *testing.T) {
		dbError := errors.New("database connection failed")
		// Мок репозитория: GetByEmail возвращает другую ошибку
		mockRepo.On("GetByEmail", ctx, userEmail).Return((*user_model.User)(nil), dbError).Once()

		user, err := service.GetUserByEmail(ctx, userEmail)

		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
		assert.ErrorContains(t, err, dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка: пустой Email", func(t *testing.T) {
		// Репозиторий не должен вызываться
		user, err := service.GetUserByEmail(ctx, "")

		assert.ErrorIs(t, err, user_service.ErrUserNotFound) // Сервис маппит пустой email в Not Found
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

// --- Тесты для GetAllUsers ---

func TestGetAllUsers(t *testing.T) {
	mockRepo, service, _ := setupTestService()
	ctx := context.Background()

	expectedUsers := []user_model.User{
		{ID: 1, Name: "User 1", Email: "user1@example.com"},
		{ID: 2, Name: "User 2", Email: "user2@example.com"},
	}
	expectedTotal := int64(10)
	page := 1
	limit := 5
	filters := map[string]interface{}{"min_age": 18, "name": "Test"}

	t.Run("Успешное получение списка пользователей с пагинацией и фильтрами", func(t *testing.T) {
		expectedParams := user_rep.ListQueryParams{
			Page: page, Limit: limit,
			MinAge: intPtr(18), Name: strPtr("Test"),
			// Добавьте другие ожидаемые поля фильтров, если они есть в user_rep.ListQueryParams
		}

		// Мок репозитория: GetAll возвращает список и общее количество
		mockRepo.On("GetAll", ctx, expectedParams).Return(expectedUsers, expectedTotal, nil).Once()

		users, total, err := service.GetAllUsers(ctx, page, limit, filters)

		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Equal(t, expectedUsers, users)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Успешное получение списка пользователей без фильтров", func(t *testing.T) {
		page = 2
		limit = 10
		filters = nil // Нет фильтров

		expectedParams := user_rep.ListQueryParams{
			Page: page, Limit: limit,
			// Поля фильтров nil
		}

		mockRepo.On("GetAll", ctx, expectedParams).Return(expectedUsers, expectedTotal, nil).Once()

		users, total, err := service.GetAllUsers(ctx, page, limit, filters)

		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Equal(t, expectedUsers, users)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Обработка невалидных параметров пагинации (page <= 0)", func(t *testing.T) {
		page = 0
		limit = 10
		filters = nil

		// Ожидается, что сервис установит page = 1
		expectedParams := user_rep.ListQueryParams{
			Page: 1, Limit: limit,
		}

		mockRepo.On("GetAll", ctx, expectedParams).Return(expectedUsers, expectedTotal, nil).Once()

		users, total, err := service.GetAllUsers(ctx, page, limit, filters)

		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Equal(t, expectedUsers, users)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Обработка невалидных параметров пагинации (limit <= 0)", func(t *testing.T) {
		page = 1
		limit = 0
		filters = nil

		// Ожидается, что сервис установит limit = 10 (или другое дефолтное значение из кода сервиса)
		// Проверьте код сервиса для точного дефолта
		expectedDefaultLimit := 10 // Предполагаем дефолтное значение из кода
		expectedParams := user_rep.ListQueryParams{
			Page: page, Limit: expectedDefaultLimit,
		}

		mockRepo.On("GetAll", ctx, expectedParams).Return(expectedUsers, expectedTotal, nil).Once()

		users, total, err := service.GetAllUsers(ctx, page, limit, filters)

		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Equal(t, expectedUsers, users)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Обработка невалидных типов фильтров", func(t *testing.T) {
		page = 1
		limit = 10
		invalidFilters := map[string]interface{}{"min_age": "not an int", "name": 123} // Некорректные типы

		// Ожидается, что сервис проигнорирует фильтры с некорректным типом
		expectedParams := user_rep.ListQueryParams{
			Page: page, Limit: limit,
			// Поля фильтров nil, т.к. не удалось распарсить
		}

		mockRepo.On("GetAll", ctx, expectedParams).Return(expectedUsers, expectedTotal, nil).Once()

		users, total, err := service.GetAllUsers(ctx, page, limit, invalidFilters)

		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Equal(t, expectedUsers, users)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка репозитория при получении списка", func(t *testing.T) {
		dbError := errors.New("database query failed")
		page = 1
		limit = 10
		filters = nil
		expectedParams := user_rep.ListQueryParams{Page: page, Limit: limit}

		// Мок репозитория: GetAll возвращает ошибку
		mockRepo.On("GetAll", ctx, expectedParams).Return(([]user_model.User)(nil), int64(0), dbError).Once()

		users, total, err := service.GetAllUsers(ctx, page, limit, filters)

		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
		assert.ErrorContains(t, err, dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
		assert.Nil(t, users)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}

// --- Тесты для UpdateUser ---

// func TestUpdateUser(t *testing.T) {
// 	mockRepo, service, _ := setupTestService()
// 	ctx := context.Background()
// 	userID := uint(1)
// 	existingUser := &user_model.User{
// 		ID: userID, Name: "Old Name", Email: "old@example.com", Age: 30, PasswordHash: "hashed_old_password",
// 	}

// 	t.Run("Успешное обновление пользователя (все поля)", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{
// 			Name:  strPtr("New Name"),
// 			Email: strPtr("new@example.com"),
// 			Age:   intPtr(35),
// 		}
// 		updatedUser := *existingUser // Копируем существующего пользователя
// 		updatedUser.Name = *req.Name
// 		updatedUser.Email = *req.Email
// 		updatedUser.Age = *req.Age

// 		// Мок репозитория: GetByID находит пользователя, GetByEmail не находит существующий email, Update успешен
// 		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
// 		mockRepo.On("GetByEmail", ctx, *req.Email).Return((*user_model.User)(nil), user_rep.ErrUserNotFound).Once() // Проверка email
// 		// Ожидаем вызов Update с обновленным пользователем
// 		mockRepo.On("Update", ctx, &updatedUser).Return(nil).Once()

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.NoError(t, err)
// 		assert.NotNil(t, user)
// 		assert.Equal(t, &updatedUser, user, "Возвращенный пользователь должен содержать обновленные данные")
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Успешное частичное обновление пользователя (только имя)", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Name: strPtr("Only Name Changed")}
// 		updatedUser := *existingUser
// 		updatedUser.Name = *req.Name

// 		// Мок репозитория: GetByID находит пользователя, Update успешен
// 		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
// 		// GetByEmail не вызывается, т.к. email не меняется
// 		mockRepo.On("Update", ctx, &updatedUser).Return(nil).Once()

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.NoError(t, err)
// 		assert.NotNil(t, user)
// 		assert.Equal(t, &updatedUser, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка: пользователь не найден", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Name: strPtr("New Name")}
// 		// Мок репозитория: GetByID возвращает ErrUserNotFound
// 		mockRepo.On("GetByID", ctx, userID).Return((*user_model.User)(nil), user_rep.ErrUserNotFound).Once()
// 		// GetByEmail и Update не должны вызываться

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.ErrorIs(t, err, user_service.ErrUserNotFound)
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка репозитория при получении пользователя для обновления", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Name: strPtr("New Name")}
// 		dbError := errors.New("database lookup failed")
// 		// Мок репозитория: GetByID возвращает другую ошибку
// 		mockRepo.On("GetByID", ctx, userID).Return((*user_model.User)(nil), dbError).Once()
// 		// GetByEmail и Update не должны вызываться

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
// 		assert.ErrorContains(t, err, dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка: попытка обновить Email на уже занятый другим пользователем", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Email: strPtr("taken@example.com")}
// 		anotherUser := &user_model.User{ID: 2, Email: *req.Email} // Другой пользователь с таким Email

// 		// Мок репозитория: GetByID находит пользователя, GetByEmail находит другого пользователя с таким Email
// 		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
// 		mockRepo.On("GetByEmail", ctx, *req.Email).Return(anotherUser, nil).Once()
// 		// Update не должен вызываться

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.ErrorIs(t, err, user_service.ErrEmailAlreadyTaken)
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка репозитория при проверке Email во время обновления", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Email: strPtr("new@example.com")}
// 		dbError := errors.New("database email check failed")

// 		// Мок репозитория: GetByID находит пользователя, GetByEmail возвращает другую ошибку
// 		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
// 		mockRepo.On("GetByEmail", ctx, *req.Email).Return((*user_model.User)(nil), dbError).Once()
// 		// Update не должен вызываться

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
// 		assert.ErrorContains(t, err, dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка: нет полей для обновления", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{} // Пустой запрос на обновление

// 		// Мок репозитория: GetByID находит пользователя
// 		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
// 		// GetByEmail и Update не должны вызываться

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.ErrorIs(t, err, user_service.ErrNoUpdateFields)
// 		assert.Equal(t, existingUser, user, "При отсутствии изменений должен возвращаться текущий пользователь")
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка репозитория при сохранении обновленного пользователя", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Name: strPtr("New Name")}
// 		dbError := errors.New("database update failed")
// 		updatedUser := *existingUser
// 		updatedUser.Name = *req.Name

// 		// Мок репозитория: GetByID находит пользователя, Update возвращает ошибку
// 		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
// 		mockRepo.On("Update", ctx, &updatedUser).Return(dbError).Once() // Мокируем ошибку Update

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError) // Ожидаем сервисную ошибку БД
// 		assert.ErrorContains(t, err, dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка репозитория: ErrNoRowsAffected при обновлении (конкурентное удаление)", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Name: strPtr("New Name")}
// 		updatedUser := *existingUser
// 		updatedUser.Name = *req.Name

// 		// Мок репозитория: GetByID находит пользователя, Update возвращает ErrNoRowsAffected
// 		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
// 		mockRepo.On("Update", ctx, &updatedUser).Return(user_rep.ErrNoRowsAffected).Once() // Мокируем ErrNoRowsAffected

// 		user, err := service.UpdateUser(ctx, userID, req)

// 		assert.ErrorIs(t, err, user_service.ErrUserNotFound) // Сервис маппит ErrNoRowsAffected в ErrUserNotFound
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка: ID пользователя равен 0", func(t *testing.T) {
// 		req := user_model.UpdateUserRequest{Name: strPtr("New Name")}
// 		// Репозиторий не должен вызываться
// 		user, err := service.UpdateUser(ctx, 0, req)

// 		assert.ErrorIs(t, err, user_service.ErrInvalidServiceInput)
// 		assert.ErrorContains(t, err.Error(), "user ID must be positive")
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})
// }

// // --- Тесты для DeleteUser ---

// func TestDeleteUser(t *testing.T) {
// 	mockRepo, service, _ := setupTestService()
// 	ctx := context.Background()
// 	userID := uint(1)

// 	t.Run("Успешное удаление пользователя", func(t *testing.T) {
// 		// Мок репозитория: Delete успешен
// 		mockRepo.On("Delete", ctx, userID).Return(nil).Once()

// 		err := service.DeleteUser(ctx, userID)

// 		assert.NoError(t, err)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка: пользователь не найден", func(t *testing.T) {
// 		// Мок репозитория: Delete возвращает ErrUserNotFound
// 		mockRepo.On("Delete", ctx, userID).Return(user_rep.ErrUserNotFound).Once()

// 		err := service.DeleteUser(ctx, userID)

// 		assert.ErrorIs(t, err, user_service.ErrUserNotFound)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка репозитория при удалении (не ErrUserNotFound)", func(t *testing.T) {
// 		dbError := errors.New("database delete failed")
// 		// Мок репозитория: Delete возвращает другую ошибку
// 		mockRepo.On("Delete", ctx, userID).Return(dbError).Once()

// 		err := service.DeleteUser(ctx, userID)

// 		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
// 		assert.ErrorContains(t, err.Error(), dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка: ID пользователя равен 0", func(t *testing.T) {
// 		// Репозиторий не должен вызываться
// 		err := service.DeleteUser(ctx, 0)

// 		assert.ErrorIs(t, err, user_service.ErrInvalidServiceInput)
// 		assert.Contains(t, err.Error(), "invalid service input: user ID must be positive")
// 		assert.Nil(t, user)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка сервиса: неверные учетные данные (пользователь не найден)", func(t *testing.T) {
// 		req := user_model.LoginRequest{Email: "nonexistent@example.com", Password: "anypassword"}
// 		// Мок репозитория: GetByEmail возвращает ErrUserNotFound
// 		mockRepo.On("GetByEmail", ctx, req.Email).Return((*user_model.User)(nil), user_rep.ErrUserNotFound).Once()

// 		token, err := service.LoginUser(ctx, req)

// 		assert.ErrorIs(t, err, user_service.ErrInvalidCredentials)
// 		assert.Empty(t, token)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка сервиса: неверные учетные данные (неверный пароль)", func(t *testing.T) {
// 		req := user_model.LoginRequest{Email: "test@example.com", Password: "wrongpassword"}
// 		existingUser := &user_model.User{ID: 1, Email: req.Email, PasswordHash: "correct_hashed_password"}

// 		// Используем реальную CheckPasswordHash, которая сравнит "wrongpassword" с хэшем.
// 		// Для надежности, здесь можно либо:
// 		// 1. Создать реальный хэш для "correctpassword" и использовать его как mock-данные PasswordHash.
// 		// 2. Мокировать функцию CheckPasswordHash, если она передается как зависимость.
// 		// В текущей структуре, полагаемся на password_util.CheckPasswordHash.
// 		// Для этого теста нам нужен хэш от "correctpassword". Сгенерируем его один раз.
// 		correctHashedPassword, _ := password_util.HashPassword("correctpassword")
// 		existingUser.PasswordHash = correctHashedPassword

// 		// Мок репозитория: GetByEmail находит пользователя
// 		mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil).Once()

// 		token, err := service.LoginUser(ctx, req)

// 		assert.ErrorIs(t, err, user_service.ErrInvalidCredentials)
// 		assert.Empty(t, token)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка репозитория при поиске пользователя для входа", func(t *testing.T) {
// 		req := user_model.LoginRequest{Email: "test@example.com", Password: "anypassword"}
// 		dbError := errors.New("database lookup failed")

// 		// Мок репозитория: GetByEmail возвращает другую ошибку
// 		mockRepo.On("GetByEmail", ctx, req.Email).Return((*user_model.User)(nil), dbError).Once()

// 		token, err := service.LoginUser(ctx, req)

// 		assert.ErrorIs(t, err, user_service.ErrServiceDatabaseError)
// 		assert.ErrorContains(t, err.Error(), dbError.Error(), "Исходная ошибка репозитория должна быть обернута")
// 		assert.Empty(t, token)
// 		mockRepo.AssertExpectations(t)
// 	})

// 	// Примечание: Мокирование jwt_util.GenerateJWT сложнее, так как это не зависимость.
// 	// Предполагаем, что jwt_util тестируется отдельно. Если GenerateJWT могла бы вернуть
// 	// ошибку при валидных входных, нужно было бы либо мокировать глобально, либо
// 	// передавать генератор токенов как зависимость.

// 	t.Run("Ошибка при генерации JWT токена", func(t *testing.T) {
// 		req := user_model.LoginRequest{Email: "test@example.com", Password: "correctpassword"}
// 		correctHashedPassword, _ := password_util.HashPassword("correctpassword")
// 		existingUser := &user_model.User{ID: 1, Email: req.Email, PasswordHash: correctHashedPassword}

// 		// Мок репозитория: GetByEmail находит пользователя
// 		mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil).Once()

// 		// В этом тесте, чтобы сымитировать ошибку генерации JWT, нам нужно временно
// 		// переопределить функцию jwt_util.GenerateJWT или изменить параметры сервиса
// 		// так, чтобы генерация гарантированно провалилась (например, пустой секрет,
// 		// но сервис уже обрабатывает пустой секрет предупреждением).
// 		// Самый чистый способ для теста - это мокировать jwt_util, если бы он был интерфейсом.
// 		// Без мокирования утилиты, тест на ошибку генерации JWT будет сложным и хрупким.
// 		// Пропустим детальный тест на ошибку генерации JWT без мокирования утилиты,
// 		// полагаясь на тесты самой jwt_util.
// 		// Если очень нужно протестировать этот путь ошибки, можно использовать технику
// 		// monkey patching (замена функции на время теста), но это не считается хорошей практикой в Go.
// 		t.Skip("Тест на ошибку генерации JWT требует мокирования jwt_util или monkey patching")
// 		/*
// 			// Пример с Monkey Patching (НЕ РЕКОМЕНДУЕТСЯ для продакшн тестов):
// 			originalGenerateJWT := jwt_util.GenerateJWT
// 			defer func() { jwt_util.GenerateJWT = originalGenerateJWT }() // Восстанавливаем функцию после теста

// 			jwt_util.GenerateJWT = func(userID uint, email, secret string, expSeconds int) (string, error) {
// 				return "", errors.New("fake jwt generation error")
// 			}

// 			token, err := service.LoginUser(ctx, req)
// 			assert.ErrorIs(t, err, user_service.ErrInternalServiceError)
// 			assert.ErrorContains(t, err.Error(), "failed to generate authentication token")
// 			assert.Empty(t, token)
// 			mockRepo.AssertExpectations(t)
// 		*/
// 	})
// }
