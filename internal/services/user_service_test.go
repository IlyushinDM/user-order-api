package services

import (
	"context"
	"errors"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	mockRepo := new(MockUserRepository) // Use the MockUserRepository from order_service_test.go
	jwtSecret := "test-secret"
	jwtExp := 3600
	userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

	// Test case 1: Successful user creation
	t.Run("Success", func(t *testing.T) {
		req := models.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Age:      30,
			Password: "password123",
		}

		// Define what the mock should return.  Crucially, we don't know the password hash, so we use a matcher.
		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound) // Simulate user not found
		mockRepo.On("Create", ctx, mock.MatchedBy(func(user *models.User) bool {
			// Check all fields *except* PasswordHash, which we can't know ahead of time
			return user.Name == req.Name && user.Email == req.Email && user.Age == req.Age
		})).Return(nil).Once() // Expect Create to be called once

		user, err := userService.CreateUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Age, user.Age)
		assert.NotEmpty(t, user.PasswordHash) // Very important:  Make sure the password *was* hashed
		mockRepo.AssertExpectations(t)        // Verify that all expected calls were made
	})

	// Test case 2: User already exists
	t.Run("UserExists", func(t *testing.T) {
		req := models.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Age:      30,
			Password: "password123",
		}

		existingUser := &models.User{
			ID:           1,
			Name:         "Existing User",
			Email:        req.Email,
			Age:          35,
			PasswordHash: "somehash",
		}
		mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil) // Simulate user found

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "user with this email already exists", err.Error())
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything) //Expect create to not be called
		mockRepo.AssertExpectations(t)
	})

	// Test case 3: Hashing fails
	t.Run("HashingFailure", func(t *testing.T) {
		req := models.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Age:      30,
			Password: "", // Causes hashing error
		}

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything)
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		mockRepo.AssertExpectations(t)
	})

	// Test case 4: Repo Create fails
	t.Run("RepoFailure", func(t *testing.T) {
		req := models.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Age:      30,
			Password: "password123",
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	mockRepo := new(MockUserRepository)
	jwtSecret := "test-secret"
	jwtExp := 3600
	userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

	t.Run("Success", func(t *testing.T) {
		userID := uint(1)
		req := models.UpdateUserRequest{
			Name:  "Updated User",
			Age:   31,
			Email: "updated@example.com",
		}

		existingUser := &models.User{
			ID:           userID,
			Name:         "Original User",
			Email:        "original@example.com",
			Age:          30,
			PasswordHash: "somehash",
		}

		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)

		// Ensure that GetByEmail is called when updating the email
		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(user *models.User) bool {
			return user.ID == userID && user.Name == req.Name && user.Email == req.Email && user.Age == req.Age
		})).Return(nil)

		updatedUser, err := userService.UpdateUser(ctx, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, req.Name, updatedUser.Name)
		assert.Equal(t, req.Email, updatedUser.Email)
		assert.Equal(t, req.Age, updatedUser.Age)

		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userID := uint(1)
		req := models.UpdateUserRequest{
			Name:  "Updated User",
			Age:   31,
			Email: "updated@example.com",
		}

		mockRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

		updatedUser, err := userService.UpdateUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		mockRepo.AssertExpectations(t)
	})

	t.Run("EmailAlreadyTaken", func(t *testing.T) {
		userID := uint(1)
		req := models.UpdateUserRequest{
			Name:  "Updated User",
			Age:   31,
			Email: "taken@example.com",
		}

		existingUser := &models.User{
			ID:           userID,
			Name:         "Original User",
			Email:        "original@example.com",
			Age:          30,
			PasswordHash: "somehash",
		}

		anotherUser := &models.User{
			ID:           2,
			Name:         "Another User",
			Email:        req.Email,
			Age:          40,
			PasswordHash: "anotherhash",
		}

		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
		mockRepo.On("GetByEmail", ctx, req.Email).Return(anotherUser, nil)

		updatedUser, err := userService.UpdateUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.Equal(t, "email already taken by another user", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateRepoFailure", func(t *testing.T) {
		userID := uint(1)
		req := models.UpdateUserRequest{
			Name:  "Updated User",
			Age:   31,
			Email: "updated@example.com",
		}

		existingUser := &models.User{
			ID:           userID,
			Name:         "Original User",
			Email:        "original@example.com",
			Age:          30,
			PasswordHash: "somehash",
		}

		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Update", ctx, mock.Anything).Return(errors.New("database error"))

		updatedUser, err := userService.UpdateUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedUser)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	mockRepo := new(MockUserRepository)
	jwtSecret := "test-secret"
	jwtExp := 3600
	userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

	t.Run("Success", func(t *testing.T) {
		userID := uint(1)

		mockRepo.On("Delete", ctx, userID).Return(nil)

		err := userService.DeleteUser(ctx, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userID := uint(1)

		mockRepo.On("Delete", ctx, userID).Return(gorm.ErrRecordNotFound)

		err := userService.DeleteUser(ctx, userID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	mockRepo := new(MockUserRepository)
	jwtSecret := "test-secret"
	jwtExp := 3600
	userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

	t.Run("Success", func(t *testing.T) {
		userID := uint(1)

		expectedUser := &models.User{
			ID:           userID,
			Name:         "Test User",
			Email:        "test@example.com",
			Age:          30,
			PasswordHash: "somehash",
		}

		mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		user, err := userService.GetUserByID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userID := uint(1)

		mockRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

		user, err := userService.GetUserByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	mockRepo := new(MockUserRepository)
	jwtSecret := "test-secret"
	jwtExp := 3600
	userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

	t.Run("Success", func(t *testing.T) {
		email := "test@example.com"

		expectedUser := &models.User{
			ID:           1,
			Name:         "Test User",
			Email:        email,
			Age:          30,
			PasswordHash: "somehash",
		}

		mockRepo.On("GetByEmail", ctx, email).Return(expectedUser, nil)

		user, err := userService.GetUserByEmail(ctx, email)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		email := "test@example.com"

		mockRepo.On("GetByEmail", ctx, email).Return(nil, gorm.ErrRecordNotFound)

		user, err := userService.GetUserByEmail(ctx, email)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetAllUsers(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	mockRepo := new(MockUserRepository)
	jwtSecret := "test-secret"
	jwtExp := 3600
	userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

	t.Run("Success", func(t *testing.T) {
		page := 1
		limit := 10
		filters := map[string]interface{}{"name": "Test"}

		expectedUsers := []models.User{
			{ID: 1, Name: "Test User 1", Email: "test1@example.com", Age: 30, PasswordHash: "somehash1"},
			{ID: 2, Name: "Test User 2", Email: "test2@example.com", Age: 40, PasswordHash: "somehash2"},
		}
		var total int64 = 2

		mockRepo.On("GetAll", ctx, page, limit, filters).Return(expectedUsers, total, nil)

		users, totalUsers, err := userService.GetAllUsers(ctx, page, limit, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		assert.Equal(t, total, totalUsers)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepoFailure", func(t *testing.T) {
		page := 1
		limit := 10
		filters := map[string]interface{}{"name": "Test"}

		mockRepo.On("GetAll", ctx, page, limit, filters).Return(nil, int64(0), errors.New("database error"))

		users, totalUsers, err := userService.GetAllUsers(ctx, page, limit, filters)

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Equal(t, int64(0), totalUsers)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_LoginUser(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	mockRepo := new(MockUserRepository)
	jwtSecret := "test-secret"
	jwtExp := 3600
	userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

	t.Run("Success", func(t *testing.T) {
		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		hashedPassword, _ := utils.HashPassword(req.Password)
		expectedUser := &models.User{
			ID:           1,
			Name:         "Test User",
			Email:        req.Email,
			Age:          30,
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(expectedUser, nil)

		token, err := userService.LoginUser(ctx, req)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		hashedPassword, _ := utils.HashPassword("password123") // Correct password
		expectedUser := &models.User{
			ID:           1,
			Name:         "Test User",
			Email:        req.Email,
			Age:          30,
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(expectedUser, nil)

		token, err := userService.LoginUser(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid credentials", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)

		token, err := userService.LoginUser(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid credentials", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("TokenGenerationFailure", func(t *testing.T) {
		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		hashedPassword, _ := utils.HashPassword(req.Password)
		expectedUser := &models.User{
			ID:           1,
			Name:         "Test User",
			Email:        req.Email,
			Age:          30,
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(expectedUser, nil)

		jwtSecret := "" // Making JWT Secret empty will cause error

		userService := NewUserService(mockRepo, logger, jwtSecret, jwtExp)

		token, err := userService.LoginUser(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, token)

		mockRepo.AssertExpectations(t)
	})
}
