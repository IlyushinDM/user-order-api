package repository

import (
	"github.com/IlyushinDM/user-order-api/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	ListUsers(page, limit, minAge, maxAge int) ([]models.User, int64, error) // Add filtering/pagination
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func GetUserByID() {

}

func GetUserByEmail() {

}

func ListUsers() {

}

func UpdateUser() {

}

func DeleteUser() {

}

// Implement other methods: GetUserByID, GetUserByEmail, ListUsers, UpdateUser, DeleteUser

// import (
// 	"fmt"
// 	"os"

// 	"github.com/IlyushinDM/user-order-api/internal/models"

// 	"github.com/joho/godotenv"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// var DB *gorm.DB

// func InitDB() error {
// 	_ = godotenv.Load()

// 	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
// 		os.Getenv("DB_HOST"),
// 		os.Getenv("DB_USER"),
// 		os.Getenv("DB_PASSWORD"),
// 		os.Getenv("DB_NAME"),
// 		os.Getenv("DB_PORT"),
// 	)

// 	var err error
// 	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		return err
// 	}

// 	return DB.AutoMigrate(&models.User{}, &models.Order{})
// }

// type PostgresRepository struct {
//     db *gorm.DB
// }

// func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
//     return &PostgresRepository{db: db}
// }

// Методы для работы с пользователями и заказами
