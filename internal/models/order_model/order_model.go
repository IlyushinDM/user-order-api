package order_model

import (
	"time"

	"gorm.io/gorm"
)

// Order represents the order model in the database.
// swagger:model Order
type Order struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"` // Foreign key
	ProductName string         `gorm:"not null;size:255" json:"product_name" binding:"required"`
	Quantity    int            `gorm:"not null" json:"quantity" binding:"required,gt=0"`
	Price       float64        `gorm:"not null" json:"price" binding:"required,gt=0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support
}

// OrderResponse defines the data returned for an order.
// swagger:response OrderResponse
type OrderResponse struct {
	ID          uint    `json:"id"`
	UserID      uint    `json:"user_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// CreateOrderRequest defines the structure for creating a new order.
// swagger:parameters CreateOrder
type CreateOrderRequest struct {
	// Name of the product
	// required: true
	// example: Laptop
	ProductName string `json:"product_name" binding:"required"`
	// Quantity of the product (must be positive)
	// required: true
	// example: 1
	Quantity int `json:"quantity" binding:"required,gt=0"`
	// Price per unit (must be positive)
	// required: true
	// example: 1200.50
	Price float64 `json:"price" binding:"required,gt=0"`
	// ID of the user placing the order (will be inferred from JWT in handler)
	// required: false
	// UserID uint `json:"user_id"` // Usually inferred from authenticated user
}

// UpdateOrderRequest defines the structure for updating an existing order.
// swagger:parameters UpdateOrder
type UpdateOrderRequest struct {
	// New name of the product (optional)
	// example: Gaming Laptop
	ProductName string `json:"product_name"`
	// New quantity (must be positive, optional)
	// example: 2
	Quantity int `json:"quantity" binding:"omitempty,gt=0"`
	// New price per unit (must be positive, optional)
	// example: 1300.00
	Price float64 `json:"price" binding:"omitempty,gt=0"`
}

// PaginatedOrdersResponse defines the structure for paginated order lists.
// swagger:response PaginatedOrdersResponse
type PaginatedOrdersResponse struct {
	// Current page number
	// example: 1
	Page int `json:"page"`
	// Number of items per page
	// example: 10
	Limit int `json:"limit"`
	// Total number of orders matching the criteria
	// example: 50
	Total int64 `json:"total"`
	// List of orders on the current page
	Orders []OrderResponse `json:"orders"`
}

// package models

// import (
// 	"context"
// 	"time"
// )

// // Order represents an order entity in the system
// type Order struct {
// 	User      User      `gorm:"foreignKey:UserID" json:"-"`
// 	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
// 	UserID    uint      `gorm:"not null" json:"user_id"`
// 	Product   string    `gorm:"not null;size:255" json:"product"`
// 	Quantity  int       `gorm:"not null" json:"quantity"`
// 	Price     float64   `gorm:"type:decimal(10, 2);not null" json:"price"`
// 	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
// }

// // OrderRepository defines the interface for order operations
// type OrderRepository interface {
// 	Create(ctx context.Context, order *Order) error
// 	GetByID(ctx context.Context, id uint) (*Order, error)
// 	Update(ctx context.Context, order *Order) error
// 	Delete(ctx context.Context, id uint) error
// }

// func (o *Order) GetProduct() string { return o.Product }

// func (o *Order) GetQuantity() int { return o.Quantity }

// func (o *Order) GetPrice() float64 { return o.Price }

// func (o *Order) GetUserID() uint { return o.UserID }
