package repository

import (
	"fmt"
	"os"

	"github.com/IlyushinDM/user-order-api/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	_ = godotenv.Load()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(&models.User{}, &models.Order{})
}

// type PostgresRepository struct {
//     db *gorm.DB
// }

// func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
//     return &PostgresRepository{db: db}
// }

// Методы для работы с пользователями и заказами
