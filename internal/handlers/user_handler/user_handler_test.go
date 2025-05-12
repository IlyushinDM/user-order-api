package user_handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/services/user_service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*user_model.User, error) {
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

// --- Helpers ---

func addAuthUserID(c *gin.Context, userID uint) {
	c.Set("userID", userID)
}

func setupUserHandlerTest() (*mockUserService, *mockCommonHandler, *UserHandler, *logrus.Logger) {
	mockSvc := new(mockUserService)
	mockCommon := new(mockCommonHandler)
	log := logrus.New()
	handler := NewUserHandler(mockSvc, mockCommon, log)
	return mockSvc, mockCommon, handler, log
}

// --- Тесты ---

func TestCreateUser_Success(t *testing.T) {
	mockSvc, _, handler, _ := setupUserHandlerTest()
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)

	reqBody := user_model.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Password: "password123",
	}
	user := &user_model.User{
		ID:    1,
		Name:  reqBody.Name,
		Email: reqBody.Email,
		Age:   reqBody.Age,
	}
	mockSvc.On("CreateUser", mock.Anything, reqBody).Return(user, nil)

	body, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/users", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateUser(c)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp user_model.UserResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, user.Name, resp.Name)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Age, resp.Age)
}

func TestCreateUser_BadRequest(t *testing.T) {
	_, _, handler, _ := setupUserHandlerTest()
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/users", bytes.NewReader([]byte("{invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateUser(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUser_ServiceError(t *testing.T) {
	mockSvc, _, handler, _ := setupUserHandlerTest()
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)

	reqBody := user_model.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Password: "password123",
	}
	mockSvc.On("CreateUser", mock.Anything, reqBody).Return(nil, user_service.ErrUserAlreadyExists)

	body, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/users", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateUser(c)
	assert.Equal(t, http.StatusConflict, w.Code)
}
