package repository

// import (
// 	"context"
// 	"errors"
// 	"testing"

// 	"github.com/IlyushinDM/user-order-api/internal/models"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/gorm"
// )

// // Ensure the interface is defined in one place (e.g., order_db_test.go)
// // type TestDB interface {
// // 	AutoMigrate(dst ...interface{}) error
// // 	Create(value interface{}) *gorm.DB
// // 	First(dest interface{}, conds ...interface{}) *gorm.DB
// // 	Save(value interface{}) *gorm.DB
// // 	Delete(value interface{}, conds ...interface{}) *gorm.DB
// // 	Migrator() gorm.Migrator
// // }

// // var testDB TestDB // Re-use the testDB from order_db_test.go
// var userRepo UserRepository

// // var testLogger *logrus.Logger //Logger

// // TestMain function to set up and tear down the test environment
// // func TestMain(m *testing.M) {
// // 	// This TestMain is intentionally empty to avoid re-initialization
// // 	// The initialization is done in order_db_test.go
// // 	os.Exit(m.Run())
// // }

// // Inject the database connection and logger
// func NewTestUserRepository(db TestDB, logger *logrus.Logger) UserRepository {
// 	return &GormUserRepository{
// 		db:     db,
// 		logger: logger,
// 	}
// }

// func TestUserRepository(t *testing.T) {
// 	// Prepare test data
// 	ctx := context.Background()
// 	testUser := &models.User{
// 		Name:         "Test User",
// 		Email:        "test@example.com",
// 		Age:          25,
// 		PasswordHash: "hashed_password",
// 	}

// 	// Initialize the user repository with the test database and logger
// 	userRepo = NewTestUserRepository(testDB, testLogger)

// 	t.Run("Create User", func(t *testing.T) {
// 		err := userRepo.Create(ctx, testUser)
// 		assert.NoError(t, err)
// 		assert.NotZero(t, testUser.ID)
// 	})

// 	t.Run("Get User By ID", func(t *testing.T) {
// 		user, err := userRepo.GetByID(ctx, testUser.ID)
// 		assert.NoError(t, err)
// 		assert.Equal(t, testUser.Name, user.Name)
// 		assert.Equal(t, testUser.Email, user.Email)
// 	})

// 	t.Run("Get User By Email", func(t *testing.T) {
// 		user, err := userRepo.GetByEmail(ctx, testUser.Email)
// 		assert.NoError(t, err)
// 		assert.Equal(t, testUser.Name, user.Name)
// 		assert.Equal(t, testUser.Email, user.Email)
// 	})

// 	t.Run("Update User", func(t *testing.T) {
// 		testUser.Name = "Updated User Name"
// 		err := userRepo.Update(ctx, testUser)
// 		assert.NoError(t, err)

// 		updatedUser, err := userRepo.GetByID(ctx, testUser.ID)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Updated User Name", updatedUser.Name)
// 	})

// 	t.Run("Delete User", func(t *testing.T) {
// 		err := userRepo.Delete(ctx, testUser.ID)
// 		assert.NoError(t, err)

// 		_, err = userRepo.GetByID(ctx, testUser.ID)
// 		assert.Error(t, err)
// 		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
// 	})

// 	t.Run("Get All Users", func(t *testing.T) {
// 		// Create some test users
// 		user1 := &models.User{Name: "User 1", Email: "user1@example.com", Age: 30, PasswordHash: "hashed_password"}
// 		user2 := &models.User{Name: "User 2", Email: "user2@example.com", Age: 35, PasswordHash: "hashed_password"}
// 		assert.NoError(t, userRepo.Create(ctx, user1))
// 		assert.NoError(t, userRepo.Create(ctx, user2))

// 		page := 1
// 		limit := 2
// 		filters := map[string]interface{}{} // No filters for this test case

// 		users, total, err := userRepo.GetAll(ctx, page, limit, filters)
// 		assert.NoError(t, err)
// 		assert.GreaterOrEqual(t, total, int64(2)) // Ensure total is at least 2
// 		assert.Len(t, users, 2)                   //The limit is 2

// 		// Clean up the created users
// 		assert.NoError(t, userRepo.Delete(ctx, user1.ID))
// 		assert.NoError(t, userRepo.Delete(ctx, user2.ID))
// 	})

// 	t.Run("Get All Users with Filters", func(t *testing.T) {
// 		// Create some test users
// 		user1 := &models.User{Name: "Alice Smith", Email: "alice@example.com", Age: 20, PasswordHash: "hashed_password"}
// 		user2 := &models.User{Name: "Bob Smith", Email: "bob@example.com", Age: 30, PasswordHash: "hashed_password"}
// 		user3 := &models.User{Name: "Charlie Brown", Email: "charlie@example.com", Age: 40, PasswordHash: "hashed_password"}
// 		assert.NoError(t, userRepo.Create(ctx, user1))
// 		assert.NoError(t, userRepo.Create(ctx, user2))
// 		assert.NoError(t, userRepo.Create(ctx, user3))

// 		page := 1
// 		limit := 10

// 		// Filter by min_age
// 		filters := map[string]interface{}{"min_age": 30}
// 		users, total, err := userRepo.GetAll(ctx, page, limit, filters)
// 		assert.NoError(t, err)
// 		assert.Equal(t, int64(2), total) // Bob and Charlie are over 30
// 		assert.Len(t, users, 2)

// 		// Filter by max_age
// 		filters = map[string]interface{}{"max_age": 30}
// 		users, total, err = userRepo.GetAll(ctx, page, limit, filters)
// 		assert.NoError(t, err)
// 		assert.Equal(t, int64(2), total) // Alice and Bob are under 30
// 		assert.Len(t, users, 2)

// 		// Filter by name
// 		filters = map[string]interface{}{"name": "Smith"}
// 		users, total, err = userRepo.GetAll(ctx, page, limit, filters)
// 		assert.NoError(t, err)
// 		assert.Equal(t, int64(2), total) // Alice Smith and Bob Smith
// 		assert.Len(t, users, 2)

// 		// Clean up the created users
// 		assert.NoError(t, userRepo.Delete(ctx, user1.ID))
// 		assert.NoError(t, userRepo.Delete(ctx, user2.ID))
// 		assert.NoError(t, userRepo.Delete(ctx, user3.ID))
// 	})
// }
