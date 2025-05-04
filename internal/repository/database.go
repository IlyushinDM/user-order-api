package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDB создает новое подключение к базе данных
func NewDB(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

type UserRepository interface {
	// Define methods for user database operations
}

type OrderRepository interface {
	// Define methods for order database operations
}

func NewUserDatabase(db *gorm.DB) UserRepository {
	return &userDatabase{DB: db}
}

type userDatabase struct {
	DB *gorm.DB
}

func NewOrderDatabase(db *gorm.DB) OrderRepository {
	return &orderDatabase{DB: db}
}

type orderDatabase struct {
	DB *gorm.DB
}
