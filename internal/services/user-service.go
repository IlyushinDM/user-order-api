package services

import (
	"github.com/IlyushinDM/user-order-api/internal/models"     // Replace with your module path
	"github.com/IlyushinDM/user-order-api/internal/repository" // Replace with your module path
	"golang.org/x/crypto/bcrypt"                               // For password hashing
)

type UserService interface {
	CreateUser(name, email, password string, age int) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	ListUsers(page, limit, minAge, maxAge int) ([]models.User, int64, error)
	UpdateUser(id uint, name, email *string, age *int) (*models.User, error)
	DeleteUser(id uint) error
	AuthenticateUser(email, password string) (*models.User, error) // For auth
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(name, email, password string, age int) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err // Handle error appropriately
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		Age:          age,
		PasswordHash: string(hashedPassword),
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		// Check for unique email constraint violation
		return nil, err // Handle error appropriately, e.g., return a specific error for duplicate email
	}

	return user, nil
}

// Implement other methods: GetUserByID, ListUsers, UpdateUser, DeleteUser, AuthenticateUser
