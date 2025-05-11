package user_model

import (
	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
)

// User представляет собой модель пользователя в базе данных
type User struct {
	ID           uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string              `gorm:"not null;size:255" json:"name" binding:"required"`
	Email        string              `gorm:"unique;not null;size:255" json:"email" binding:"required,email"`
	Age          int                 `gorm:"not null" json:"age" binding:"required,gt=0"`
	PasswordHash string              `gorm:"not null" json:"-"`
	Orders       []order_model.Order `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"orders,omitempty"`
}

// UserResponse определяет данные, возвращаемые пользователю (за исключением конфиденциальной информации)
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// CreateUserRequest определяет структуру для создания нового пользователя
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,gt=0"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest определяет структуру для обновления существующего пользователя
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"omitempty,email"`
	Age   int    `json:"age" binding:"omitempty,gt=0"`
}

// PaginatedUsersResponse определяет структуру постраничных списков пользователей
type PaginatedUsersResponse struct {
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	Total int64          `json:"total"`
	Users []UserResponse `json:"users"`
}

// LoginRequest определяет структуру для входа пользователя в систему
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse определяет структуру ответа на вход в систему (токен JWT)
type LoginResponse struct {
	Token string `json:"token"`
}
