package services

import (
	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/repository"
)

type UserService interface {
	CreateUser(user models.User) (models.User, error)
	GetUser(id uint) (models.User, error)
}

type userService struct {
	userRepository repository.UserRepository
}

// func NewUserService(userRepository repository.UserRepository) UserService {
// 	return &userService{userRepository: userRepository}
// }

// func (s *userService) CreateUser(user models.User) (models.User, error) {
// 	// Допустим, здесь происходит валидация данных, хеширование пароля и т.д.
// 	// ...
// 	return s.userRepository.Create(user)
// }

// func (s *userService) GetUser(id uint) (models.User, error) {
// 	return s.userRepository.Get(id)
// }
