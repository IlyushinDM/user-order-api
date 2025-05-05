package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/repository"
	"github.com/IlyushinDM/user-order-api/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserService defines the interface for user business logic.
type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
	UpdateUser(ctx context.Context, id uint, req models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error) // Needed for auth
	GetAllUsers(ctx context.Context, page, limit int, filters map[string]interface{}) ([]models.User, int64, error)
	LoginUser(ctx context.Context, req models.LoginRequest) (string, error)
}

type userService struct {
	userRepo  repository.UserRepository
	log       *logrus.Logger
	jwtSecret string
	jwtExpSec int
}

// NewUserService creates a new user service.
func NewUserService(repo repository.UserRepository, log *logrus.Logger, jwtSecret string, jwtExp int) UserService {
	return &userService{userRepo: repo, log: log, jwtSecret: jwtSecret, jwtExpSec: jwtExp}
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.CreateUser").WithField("email", req.Email)

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.WithError(err).Error("Error checking for existing user")
		return nil, fmt.Errorf("database error checking user existence: %w", err)
	}
	if existingUser != nil {
		logger.Warn("User creation attempted with existing email")
		return nil, errors.New("user with this email already exists") // Consider a custom error type
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		Age:          req.Age,
		PasswordHash: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.WithError(err).Error("Failed to create user in repository")
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	logger.WithField("user_id", user.ID).Info("User created successfully")
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint, req models.UpdateUserRequest) (*models.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.UpdateUser").WithField("user_id", id)

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Update failed: User not found")
			return nil, err // Return specific error
		}
		logger.WithError(err).Error("Failed to get user for update")
		return nil, fmt.Errorf("database error finding user: %w", err)
	}

	updated := false
	if req.Name != "" && req.Name != user.Name {
		user.Name = req.Name
		updated = true
		logger.Debug("Updating user name")
	}
	if req.Age > 0 && req.Age != user.Age {
		user.Age = req.Age
		updated = true
		logger.Debug("Updating user age")
	}
	if req.Email != "" && req.Email != user.Email {
		// Check if the new email is already taken by another user
		existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.WithError(err).Error("Error checking for existing email during update")
			return nil, fmt.Errorf("database error checking email uniqueness: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			logger.Warn("Update failed: Email already taken by another user")
			return nil, errors.New("email already taken by another user")
		}
		user.Email = req.Email
		updated = true
		logger.Debug("Updating user email")
	}

	if !updated {
		logger.Info("No fields to update for user")
		return user, nil // Return current user data if no changes
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.WithError(err).Error("Failed to update user in repository")
		return nil, fmt.Errorf("failed to save updated user: %w", err)
	}

	logger.Info("User updated successfully")
	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.DeleteUser").WithField("user_id", id)
	err := s.userRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Deletion failed: User not found")
		} else {
			logger.WithError(err).Error("Failed to delete user in repository")
		}
		return err // Propagate error (could be RecordNotFound or DB error)
	}
	logger.Info("User deleted successfully")
	return nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetUserByID").WithField("user_id", id)
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("User not found")
		} else {
			logger.WithError(err).Error("Failed to get user from repository")
		}
		return nil, err
	}
	logger.Info("User retrieved successfully")
	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetUserByEmail").WithField("email", email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("User not found by email")
		} else {
			logger.WithError(err).Error("Failed to get user by email from repository")
		}
		return nil, err
	}
	logger.Info("User retrieved successfully by email")
	return user, nil
}

func (s *userService) GetAllUsers(ctx context.Context, page, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.GetAllUsers")
	users, total, err := s.userRepo.GetAll(ctx, page, limit, filters)
	if err != nil {
		logger.WithError(err).Error("Failed to get all users from repository")
		return nil, 0, err
	}
	logger.WithFields(logrus.Fields{"count": len(users), "total": total}).Info("Retrieved all users successfully")
	return users, total, nil
}

func (s *userService) LoginUser(ctx context.Context, req models.LoginRequest) (string, error) {
	logger := s.log.WithContext(ctx).WithField("method", "UserService.LoginUser").WithField("email", req.Email)

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Login attempt failed: User not found")
			return "", errors.New("invalid credentials") // Generic error
		}
		logger.WithError(err).Error("Database error during login")
		return "", fmt.Errorf("database error: %w", err)
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		logger.Warn("Login attempt failed: Invalid password")
		return "", errors.New("invalid credentials") // Generic error
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, s.jwtSecret, s.jwtExpSec)
	if err != nil {
		logger.WithError(err).Error("Failed to generate JWT token")
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	logger.WithField("user_id", user.ID).Info("User logged in successfully")
	return token, nil
}
