package repository

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/IlyushinDM/user-order-api/internal/models"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"
// )

// // Define an interface for the database connection
// type TestDB interface {
// 	 автоматом Migrate(dst ...interface{}) error
// 	Create(value interface{}) *gorm.DB
// 	First(dest interface{}, conds ...interface{}) *gorm.DB
// 	Save(value interface{}) *gorm.DB
// 	Delete(value interface{}, conds ...interface{}) *gorm.DB
// 	Migrator() gorm.Migrator
// }

// var testDB *gorm.DB
// var orderRepo OrderRepository
// var testLogger *logrus.Logger

// // TestMain function to set up and tear down the test environment
// func TestMain(m *testing.M) {
// 	// Setup test database and repository
// 	var err error
// 	dsn := os.Getenv("DATABASE_URL") // Retrieve the DATABASE_URL from the environment
// 	if dsn == "" {
// 		dsn = "host=localhost user=postgres password=postgres dbname=user_order_api_test port=5432 sslmode=disable" // Default local testing URL
// 		log.Println("DATABASE_URL not set, using default:", dsn)                                                    // Using standard log for visibility
// 	}

// 	newLogger := logger.New(
// 		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
// 		logger.Config{
// 			SlowThreshold: time.Second, // Slow SQL threshold
// 			LogLevel:      logger.Info, // Log level
// 			Colorful:      true,        // Disable color
// 		},
// 	)

// 	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
// 		Logger: newLogger,
// 	})

// 	if err != nil {
// 		log.Fatalf("Failed to connect to test database: %v", err) // Using standard log for visibility
// 	}

// 	// AutoMigrate the schema
// 	err = testDB.AutoMigrate(&models.User{}, &models.Order{})
// 	if err != nil {
// 		log.Fatalf("Failed to migrate test database schema: %v", err)
// 	}

// 	testLogger = logrus.New()
// 	testLogger.SetLevel(logrus.DebugLevel) // Set the log level to Debug
// 	testLogger.SetOutput(os.Stdout)        // Send logs to stdout

// 	orderRepo = NewGormOrderRepository(testDB, testLogger)

// 	// Run tests
// 	exitCode := m.Run()

// 	// Teardown: Clean up the database after tests
// 	// Drop all tables to ensure a clean slate for the next test run.
// 	err = testDB.Migrator().DropTable(&models.Order{}, &models.User{})
// 	if err != nil {
// 		log.Printf("Failed to drop tables: %v", err) // Non-fatal, log the error
// 	}

// 	os.Exit(exitCode)
// }

// func TestOrderRepository(t *testing.T) {
// 	// Prepare test data
// 	ctx := context.Background()

// 	// Create a test user
// 	testUser := &models.User{
// 		Name:         "Test User",
// 		Email:        "test@example.com",
// 		Age:          25,
// 		PasswordHash: "hashed_password",
// 	}
// 	result := testDB.Create(testUser)
// 	assert.NoError(t, result.Error)
// 	assert.NotZero(t, testUser.ID)

// 	// Create a test order
// 	testOrder := &models.Order{
// 		UserID:      testUser.ID,
// 		ProductName: "Test Product",
// 		Quantity:    2,
// 		Price:       99.99,
// 	}

// 	t.Run("Create Order", func(t *testing.T) {
// 		err := orderRepo.Create(ctx, testOrder)
// 		assert.NoError(t, err)
// 		assert.NotZero(t, testOrder.ID)
// 	})

// 	t.Run("Get Order By ID", func(t *testing.T) {
// 		order, err := orderRepo.GetByID(ctx, testOrder.ID, testUser.ID)
// 		assert.NoError(t, err)
// 		assert.Equal(t, testOrder.ProductName, order.ProductName)
// 		assert.Equal(t, testUser.ID, order.UserID)
// 	})

// 	t.Run("Update Order", func(t *testing.T) {
// 		testOrder.ProductName = "Updated Product Name"
// 		err := orderRepo.Update(ctx, testOrder)
// 		assert.NoError(t, err)

// 		updatedOrder, err := orderRepo.GetByID(ctx, testOrder.ID, testUser.ID)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Updated Product Name", updatedOrder.ProductName)
// 	})

// 	t.Run("Update Order - Not Found", func(t *testing.T) {
// 		invalidOrder := &models.Order{
// 			ID:          99999, //some random ID
// 			UserID:      testUser.ID,
// 			ProductName: "Should not update",
// 			Quantity:    1,
// 			Price:       10.00,
// 		}
// 		err := orderRepo.Update(ctx, invalidOrder)
// 		assert.Error(t, err)
// 		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
// 	})

// 	t.Run("Delete Order", func(t *testing.T) {
// 		err := orderRepo.Delete(ctx, testOrder.ID, testUser.ID)
// 		assert.NoError(t, err)

// 		_, err = orderRepo.GetByID(ctx, testOrder.ID, testUser.ID)
// 		assert.Error(t, err)
// 		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
// 	})

// 	t.Run("Delete Order - Not Found", func(t *testing.T) {
// 		err := orderRepo.Delete(ctx, 9999, testUser.ID) //some random ID
// 		assert.Error(t, err)
// 		assert.ErrorContains(t, err, "permission denied or record not found")
// 	})

// 	t.Run("GetAllByUser", func(t *testing.T) {
// 		// Create multiple orders for the user
// 		order1 := &models.Order{UserID: testUser.ID, ProductName: "Product 1", Quantity: 1, Price: 50.00}
// 		order2 := &models.Order{UserID: testUser.ID, ProductName: "Product 2", Quantity: 2, Price: 100.00}
// 		order3 := &models.Order{UserID: testUser.ID, ProductName: "Product 3", Quantity: 3, Price: 150.00}
// 		assert.NoError(t, orderRepo.Create(ctx, order1))
// 		assert.NoError(t, orderRepo.Create(ctx, order2))
// 		assert.NoError(t, orderRepo.Create(ctx, order3))

// 		page := 1
// 		limit := 2
// 		orders, total, err := orderRepo.GetAllByUser(ctx, testUser.ID, page, limit)
// 		assert.NoError(t, err)
// 		assert.Equal(t, int64(3), total) // Total number of orders for the user
// 		assert.Len(t, orders, 2)         // Number of orders returned in the current page

// 		// Verify the contents of the returned orders (at least partially)
// 		assert.Equal(t, "Product 1", orders[0].ProductName)
// 		assert.Equal(t, "Product 2", orders[1].ProductName)

// 		// Clean up the created orders
// 		assert.NoError(t, orderRepo.Delete(ctx, order1.ID, testUser.ID))
// 		assert.NoError(t, orderRepo.Delete(ctx, order2.ID, testUser.ID))
// 		assert.NoError(t, orderRepo.Delete(ctx, order3.ID, testUser.ID))
// 	})

// 	// Cleanup: Delete the test user
// 	err := testDB.Delete(testUser).Error
// 	if err != nil {
// 		fmt.Printf("Failed to delete test user: %v\n", err)
// 	}
// }
