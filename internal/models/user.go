package models

type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"` // ;autoIncrement"
	Name         string `json:"name"`                 // gorm:"size:255;not null"
	Email        string `json:"email" gorm:"unique"`  // ;size:255;not null"
	Age          int    `json:"age"`                  //  gorm:"default:0"`
	PasswordHash string `json:"-"`                    // gorm:"not null"`
}
