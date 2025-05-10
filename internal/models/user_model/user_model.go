package user_model

import (
	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
)

// User represents the user model in the database.
// swagger:model User
type User struct {
	ID           uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string              `gorm:"not null;size:255" json:"name" binding:"required"`
	Email        string              `gorm:"unique;not null;size:255" json:"email" binding:"required,email"`
	Age          int                 `gorm:"not null" json:"age" binding:"required,gt=0"`
	PasswordHash string              `gorm:"not null" json:"-"` // Never expose hash
	Orders       []order_model.Order `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"orders,omitempty"`
}

// UserResponse defines the data returned for a user (excluding sensitive info).
// swagger:response UserResponse
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// CreateUserRequest defines the structure for creating a new user.
// swagger:parameters CreateUser
type CreateUserRequest struct {
	// User's full name
	// required: true
	// example: John Doe
	Name string `json:"name" binding:"required"`
	// User's unique email address
	// required: true
	// example: john.doe@example.com
	Email string `json:"email" binding:"required,email"`
	// User's age (must be positive)
	// required: true
	// example: 30
	Age int `json:"age" binding:"required,gt=0"`
	// User's password (min 6 characters)
	// required: true
	// example: password123
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest defines the structure for updating an existing user.
// swagger:parameters UpdateUser
type UpdateUserRequest struct {
	// User's full name (optional)
	Name string `json:"name"`
	// User's unique email address (optional)
	Email string `json:"email" binding:"omitempty,email"`
	// User's age (must be positive, optional)
	Age int `json:"age" binding:"omitempty,gt=0"`
}

// PaginatedUsersResponse defines the structure for paginated user lists.
// swagger:response PaginatedUsersResponse
type PaginatedUsersResponse struct {
	// Current page number
	Page int `json:"page"`
	// Number of items per page
	Limit int `json:"limit"`
	// Total number of users matching the criteria
	Total int64 `json:"total"`
	// List of users on the current page
	Users []UserResponse `json:"users"`
}

// LoginRequest defines the structure for user login.
// swagger:parameters LoginUser
type LoginRequest struct {
	// User's email address
	// required: true
	Email string `json:"email" binding:"required,email"`
	// User's password
	// required: true
	Password string `json:"password" binding:"required"`
}

// LoginResponse defines the structure for the login response (JWT token).
// swagger:response LoginResponse
type LoginResponse struct {
	// JWT authentication token
	Token string `json:"token"`
}
