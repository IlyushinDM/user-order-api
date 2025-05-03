package models

import "time"

type Order struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Product   string    `gorm:"not null" json:"product"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10, 2);not null" json:"price"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// type Order struct {
// 	ID        uint      `json:"id" gorm:"primaryKey"`
// 	UserID    uint      `json:"user_id"`
// 	Product   string    `json:"product"`
// 	Quantity  int       `json:"quantity"`
// 	Price     float64   `json:"price"`
// 	CreatedAt time.Time `json:"created_at"`
// }
