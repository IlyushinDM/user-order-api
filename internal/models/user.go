package models

type User struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Name         string  `gorm:"not null" json:"name"`
	Email        string  `gorm:"unique;not null" json:"email"`
	Age          int     `gorm:"not null" json:"age"`
	PasswordHash string  `gorm:"not null" json:"-"` // Use '-' to hide password hash in JSON
	Orders       []Order `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}

// type User struct {
// 	ID           uint   `json:"id" gorm:"primaryKey"` // ;autoIncrement"
// 	Name         string `json:"name"`                 // gorm:"size:255;not null"
// 	Email        string `json:"email" gorm:"unique"`  // ;size:255;not null"
// 	Age          int    `json:"age"`                  //  gorm:"default:0"`
// 	PasswordHash string `json:"-"`                    // gorm:"not null"`
// }
