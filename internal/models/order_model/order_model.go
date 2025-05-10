package order_model

import (
	"time"

	"gorm.io/gorm"
)

// Order представляет модель заказа в базе данных
type Order struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	ProductName string         `gorm:"not null;size:255" json:"product_name" binding:"required"`
	Quantity    int            `gorm:"not null" json:"quantity" binding:"required,gt=0"`
	Price       float64        `gorm:"not null" json:"price" binding:"required,gt=0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// OrderResponse определяет структуру ответа с данными заказа
type OrderResponse struct {
	ID          uint    `json:"id"`
	UserID      uint    `json:"user_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// CreateOrderRequest определяет структуру запроса для создания заказа
type CreateOrderRequest struct {
	ProductName string  `json:"product_name" binding:"required"`  // Название продукта (обязательно)
	Quantity    int     `json:"quantity" binding:"required,gt=0"` // Количество (положительное число)
	Price       float64 `json:"price" binding:"required,gt=0"`    // Цена за единицу (положительное число)
}

// UpdateOrderRequest определяет структуру запроса для обновления заказа
type UpdateOrderRequest struct {
	ProductName string  `json:"product_name"`                      // Новое название продукта
	Quantity    int     `json:"quantity" binding:"omitempty,gt=0"` // Новое количество (опционально)
	Price       float64 `json:"price" binding:"omitempty,gt=0"`    // Новая цена (опционально)
}

// PaginatedOrdersResponse определяет структуру для пагинированного списка заказов
type PaginatedOrdersResponse struct {
	Page   int             `json:"page"`   // Текущая страница
	Limit  int             `json:"limit"`  // Количество элементов на странице
	Total  int64           `json:"total"`  // Общее количество заказов
	Orders []OrderResponse `json:"orders"` // Список заказов
}
