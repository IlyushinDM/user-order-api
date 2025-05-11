package user_handler

import (
	"context"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// --- Моки ---

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) CreateUser(ctx context.Context, req user_model.CreateUserRequest) (*user_model.User, error) {
	args := m.Called(ctx, req)
	user, _ := args.Get(0).(*user_model.User)
	return user, args.Error(1)
}

func (m *mockUserService) GetUserByID(ctx context.Context, id uint) (*user_model.User, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*user_model.User)
	return user, args.Error(1)
}

func (m *mockUserService) GetAllUsers(ctx context.Context, page, limit int, filters map[string]any) ([]user_model.User, int64, error) {
	args := m.Called(ctx, page, limit, filters)
	return args.Get(0).([]user_model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserService) UpdateUser(ctx context.Context, id uint, req user_model.UpdateUserRequest) (*user_model.User, error) {
	args := m.Called(ctx, id, req)
	user, _ := args.Get(0).(*user_model.User)
	return user, args.Error(1)
}

func (m *mockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserService) LoginUser(ctx context.Context, req user_model.LoginRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email uint) (*user_model.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(*user_model.User)
	return user, args.Error(1)
}

type mockCommonHandler struct {
	mock.Mock
}

func (m *mockCommonHandler) GetPaginationParams(c *gin.Context) (int, int, error) {
	args := m.Called(c)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockCommonHandler) GetFilteringParams(c *gin.Context) (map[string]any, error) {
	args := m.Called(c)
	return args.Get(0).(map[string]any), args.Error(1)
}

// --- Тесты ---

// func setupRouter(handler *UserHandler) *gin.Engine {
// 	gin.SetMode(gin.TestMode)
// 	r := gin.New()
// 	r.POST("/api/users", handler.CreateUser)
// 	r.GET("/api/users/:id", handler.GetUserByID)
// 	r.GET("/api/users", handler.GetAllUsers)
// 	r.PUT("/api/users/:id", handler.UpdateUser)
// 	r.DELETE("/api/users/:id", handler.DeleteUser)
// 	r.POST("/auth/login", handler.LoginUser)
// 	return r
// }
