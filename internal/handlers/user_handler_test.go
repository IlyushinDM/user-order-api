package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/IlyushinDM/user-order-api/internal/models"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers(ctx context.Context, page, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	args := m.Called(ctx, page, limit, filters)
	return args.Get(0).([]models.User), int64(args.Int(1)), args.Error(2)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint, req models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) LoginUser(ctx context.Context, req models.LoginRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

// Add this method to implement the UserService interface
func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func setupUserHandlerTest(t *testing.T) (*gin.Context, *httptest.ResponseRecorder, *MockUserService, *UserHandler) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("logger", &logrus.Logger{})

	mockUserService := new(MockUserService)
	userHandler := NewUserHandler(mockUserService, &logrus.Logger{})

	return c, w, mockUserService, userHandler
}

func TestCreateUser(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Mock request body
	reqBody := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      30,
		Password: "password",
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	// Mock service response
	mockUserService.On("CreateUser", mock.Anything, reqBody).Return(&models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Age:   30,
	}, nil)

	// Call the handler
	userHandler.CreateUser(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"email":"test@example.com"`)
	mockUserService.AssertExpectations(t)
}

func TestGetUserByID(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Set up context
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/users/1", nil)

	// Mock service response
	mockUserService.On("GetUserByID", mock.Anything, uint(1)).Return(&models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Age:   30,
	}, nil)

	// Call the handler
	userHandler.GetUserByID(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"email":"test@example.com"`)
	mockUserService.AssertExpectations(t)
}

func TestGetAllUsers(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Set up context
	c.Request, _ = http.NewRequest(http.MethodGet, "/users?page=1&limit=10", nil)

	// Mock service response
	mockUserService.On("GetAllUsers", mock.Anything, 1, 10, map[string]interface{}(nil)).Return([]models.User{
		{ID: 1, Name: "Test User", Email: "test@example.com", Age: 30},
	}, int64(1), nil)

	// Call the handler
	userHandler.GetAllUsers(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"email":"test@example.com"`)
	mockUserService.AssertExpectations(t)
}

func TestUpdateUser(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Set up context
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	reqBody := models.UpdateUserRequest{
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   31,
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", uint(1)) // Simulate authenticated user

	// Mock service response
	mockUserService.On("UpdateUser", mock.Anything, uint(1), reqBody).Return(&models.User{
		ID:    1,
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   31,
	}, nil)

	// Call the handler
	userHandler.UpdateUser(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"email":"updated@example.com"`)
	mockUserService.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Set up context
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/users/1", nil)
	c.Set("userID", uint(1)) // Simulate authenticated user

	// Mock service response
	mockUserService.On("DeleteUser", mock.Anything, uint(1)).Return(nil)

	// Call the handler
	userHandler.DeleteUser(c)

	// Assertions
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUserService.AssertExpectations(t)
}

func TestLoginUser(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Mock request body
	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password",
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	// Mock service response
	mockUserService.On("LoginUser", mock.Anything, reqBody).Return("test_token", nil)

	// Call the handler
	userHandler.LoginUser(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"token":"test_token"`)
	mockUserService.AssertExpectations(t)
}

func TestCreateUser_Conflict(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Mock request body
	reqBody := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      30,
		Password: "password",
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	// Mock service to return an error indicating a conflict (email already exists)
	mockUserService.On("CreateUser", mock.Anything, reqBody).Return(nil, errors.New("User with this email already exists"))

	// Call the handler
	userHandler.CreateUser(c)

	// Assertions
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "User with this email already exists")
	mockUserService.AssertExpectations(t)
}

func TestGetUserByID_NotFound(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Set up context
	userID := uint(999) // Non-existent user ID
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(int(userID))}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/users/"+strconv.Itoa(int(userID)), nil)

	// Mock service to return error indicating user not found
	mockUserService.On("GetUserByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	// Call the handler
	userHandler.GetUserByID(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
	mockUserService.AssertExpectations(t)
}

func TestUpdateUser_Forbidden(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Set up context
	userID := uint(2)     // Different user ID than the one in the context
	authUserID := uint(1) // Authenticated user ID
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(int(userID))}}
	reqBody := models.UpdateUserRequest{
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   31,
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPut, "/users/"+strconv.Itoa(int(userID)), bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", authUserID) // Simulate authenticated user

	// Call the handler
	userHandler.UpdateUser(c)

	// Assertions
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Forbidden: You can only update your own profile")
	mockUserService.AssertNotCalled(t, "UpdateUser") // Ensure service is not called
}

func TestDeleteUser_Forbidden(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Set up context
	userID := uint(2)     // Different user ID than the one in the context
	authUserID := uint(1) // Authenticated user ID
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(int(userID))}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/users/"+strconv.Itoa(int(userID)), nil)
	c.Set("userID", authUserID) // Simulate authenticated user

	// Call the handler
	userHandler.DeleteUser(c)

	// Assertions
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Forbidden: You can only delete your own account")
	mockUserService.AssertNotCalled(t, "DeleteUser") // Ensure service is not called
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	c, w, mockUserService, userHandler := setupUserHandlerTest(t)

	// Mock request body
	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong_password",
	}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	// Mock service to return an error indicating invalid credentials
	mockUserService.On("LoginUser", mock.Anything, reqBody).Return("", errors.New("invalid credentials"))

	// Call the handler
	userHandler.LoginUser(c)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid credentials")
	mockUserService.AssertExpectations(t)
}
