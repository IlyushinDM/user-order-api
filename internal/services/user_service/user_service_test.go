package user_service_test

import (
	"context"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/repository/user_rep"
	"github.com/IlyushinDM/user-order-api/internal/services/user_service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository реализует интерфейс UserRepository для тестирования
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *user_model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *user_model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*user_model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*user_model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*user_model.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context, params user_rep.ListQueryParams) ([]user_model.User, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]user_model.User), args.Get(1).(int64), args.Error(2)
}

// TestNewUserService тестирует создание нового сервиса
func TestNewUserService(t *testing.T) {
	t.Run("Успешное создание сервиса", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		logger := logrus.New()

		service := user_service.NewUserService(mockRepo, logger, "secret", 3600)
		assert.NotNil(t, service)
	})

	t.Run("Создание сервиса с nil репозиторием", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Ожидалась паника при передаче nil репозитория")
			}
		}()

		user_service.NewUserService(nil, logrus.New(), "secret", 3600)
	})
}

// TestGetUserByID тестирует получение пользователя по ID
func TestGetUserByID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := user_service.NewUserService(mockRepo, logrus.New(), "secret", 3600)

	testUser := &user_model.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}

	t.Run("Успешное получение пользователя", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, testUser.ID).Return(testUser, nil)

		user, err := service.GetUserByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("Ошибка: пользователь не найден", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, uint(2)).Return((*user_model.User)(nil), user_rep.ErrUserNotFound)

		_, err := service.GetUserByID(ctx, 2)
		assert.Error(t, err)
		assert.Equal(t, user_service.ErrUserNotFound, err)
	})

	t.Run("Ошибка: нулевой ID", func(t *testing.T) {
		_, err := service.GetUserByID(ctx, 0)
		assert.Error(t, err)
		assert.Equal(t, user_service.ErrUserNotFound, err)
	})
}

// TestUpdateUser тестирует обновление пользователя
func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := user_service.NewUserService(mockRepo, logrus.New(), "secret", 3600)

	existingUser := &user_model.User{
		ID:    1,
		Name:  "Old Name",
		Email: "old@example.com",
		Age:   30,
	}

	t.Run("Успешное обновление пользователя", func(t *testing.T) {
		updateReq := user_model.UpdateUserRequest{
			Name:  "New Name",
			Email: "new@example.com",
			Age:   35,
		}

		mockRepo.On("GetByID", ctx, existingUser.ID).Return(existingUser, nil)
		mockRepo.On("GetByEmail", ctx, updateReq.Email).Return((*user_model.User)(nil), user_rep.ErrUserNotFound)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*user_model.User")).Return(nil)

		updatedUser, err := service.UpdateUser(ctx, existingUser.ID, updateReq)
		assert.NoError(t, err)
		assert.Equal(t, updateReq.Name, updatedUser.Name)
		assert.Equal(t, updateReq.Email, updatedUser.Email)
		assert.Equal(t, updateReq.Age, updatedUser.Age)
	})

	t.Run("Ошибка: email уже занят другим пользователем", func(t *testing.T) {
		updateReq := user_model.UpdateUserRequest{
			Email: "taken@example.com",
		}

		otherUser := &user_model.User{
			ID:    2,
			Email: "taken@example.com",
		}

		mockRepo.On("GetByID", ctx, existingUser.ID).Return(existingUser, nil)
		mockRepo.On("GetByEmail", ctx, updateReq.Email).Return(otherUser, nil)

		_, err := service.UpdateUser(ctx, existingUser.ID, updateReq)
		assert.Error(t, err)
		assert.Equal(t, user_service.ErrEmailAlreadyTaken, err)
	})
}

// TestDeleteUser тестирует удаление пользователя
func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := user_service.NewUserService(mockRepo, logrus.New(), "secret", 3600)

	t.Run("Успешное удаление пользователя", func(t *testing.T) {
		userID := uint(1)
		mockRepo.On("Delete", ctx, userID).Return(nil)

		err := service.DeleteUser(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("Ошибка: пользователь не найден", func(t *testing.T) {
		userID := uint(2)
		mockRepo.On("Delete", ctx, userID).Return(user_rep.ErrUserNotFound)

		err := service.DeleteUser(ctx, userID)
		assert.Error(t, err)
		assert.Equal(t, user_service.ErrUserNotFound, err)
	})
}

// TestGetAllUsers тестирует получение списка пользователей
func TestGetAllUsers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := user_service.NewUserService(mockRepo, logrus.New(), "secret", 3600)

	testUsers := []user_model.User{
		{ID: 1, Name: "User 1"},
		{ID: 2, Name: "User 2"},
	}

	t.Run("Успешное получение пользователей", func(t *testing.T) {
		mockRepo.On("GetAll", ctx, mock.AnythingOfType("user_rep.ListQueryParams")).Return(testUsers, int64(2), nil)

		users, total, err := service.GetAllUsers(ctx, 1, 10, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, int64(2), total)
	})

	t.Run("Применение фильтров", func(t *testing.T) {
		filters := map[string]any{
			"min_age": 18,
		}

		mockRepo.On("GetAll", ctx, mock.MatchedBy(func(params user_rep.ListQueryParams) bool {
			return params.MinAge != nil && *params.MinAge == 18
		})).Return(testUsers, int64(2), nil)

		_, _, err := service.GetAllUsers(ctx, 1, 10, filters)
		assert.NoError(t, err)
	})
}
