package models

type User struct {
	ID           uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string  `gorm:"not null;size:255" json:"name"`
	Email        string  `gorm:"unique;not null;size:255" json:"email"`
	Age          int     `gorm:"not null;size:3" json:"age"`
	PasswordHash string  `gorm:"not null" json:"-"`
	Orders       []Order `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"orders,omitempty"`
}
