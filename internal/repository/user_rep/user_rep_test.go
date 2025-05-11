package user_rep

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite DB and migrates the User model.
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&user_model.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db, func() {}
}

func newTestRepo(t *testing.T) (*GormUserRepository, func()) {
	db, cleanup := setupTestDB(t)
	logger := logrus.New()
	return &GormUserRepository{db: db, log: logger}, cleanup
}

func TestCreateUser_Success(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	user := &user_model.User{Name: "Alice", Email: "alice@example.com", Age: 30}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID == 0 {
		t.Error("expected user ID to be set after creation")
	}
}

func TestCreateUser_NilUser(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	err := repo.Create(ctx, nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	user := &user_model.User{Name: "Bob", Email: "bob@example.com", Age: 25}
	_ = repo.Create(ctx, user)

	user.Name = "Bobby"
	err := repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	updated, _ := repo.GetByID(ctx, user.ID)
	if updated.Name != "Bobby" {
		t.Errorf("expected name to be updated, got %s", updated.Name)
	}
}

func TestUpdateUser_ZeroID(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	user := &user_model.User{Name: "Zero", Email: "zero@example.com", Age: 20}

	err := repo.Update(ctx, user)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateUser_NoRowsAffected(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	user := &user_model.User{ID: 9999, Name: "Ghost", Email: "ghost@example.com", Age: 40}

	err := repo.Update(ctx, user)
	if !errors.Is(err, ErrNoRowsAffected) {
		t.Errorf("expected ErrNoRowsAffected, got %v", err)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	user := &user_model.User{Name: "Del", Email: "del@example.com", Age: 22}
	_ = repo.Create(ctx, user)

	err := repo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteUser_ZeroID(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	err := repo.Delete(ctx, 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	err := repo.Delete(ctx, 9999)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	user := &user_model.User{Name: "FindMe", Email: "findme@example.com", Age: 28}
	_ = repo.Create(ctx, user)

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, got.Email)
	}
}

func TestGetByID_ZeroID(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 0)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 9999)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetByEmail_Success(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	user := &user_model.User{Name: "Mail", Email: "mail@example.com", Age: 33}
	_ = repo.Create(ctx, user)

	got, err := repo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Name != user.Name {
		t.Errorf("expected name %s, got %s", user.Name, got.Name)
	}
}

func TestGetByEmail_EmptyEmail(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetByEmail_NotFound(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "notfound@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetAll_BasicPagination(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	// Insert 15 users
	for i := 1; i <= 15; i++ {
		user := &user_model.User{
			Name:  fmt.Sprintf("User%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   20 + i,
		}
		_ = repo.Create(ctx, user)
	}

	params := ListQueryParams{Page: 2, Limit: 5}
	users, total, err := repo.GetAll(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 5 {
		t.Errorf("expected 5 users, got %d", len(users))
	}
	if total != 15 {
		t.Errorf("expected total 15, got %d", total)
	}
}

func TestGetAll_WithFilters(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	// Insert users with different ages and names
	_ = repo.Create(ctx, &user_model.User{Name: "Anna", Email: "anna@example.com", Age: 21})
	_ = repo.Create(ctx, &user_model.User{Name: "Annabelle", Email: "annabelle@example.com", Age: 25})
	_ = repo.Create(ctx, &user_model.User{Name: "Bob", Email: "bob@example.com", Age: 30})

	minAge := 22
	name := "Ann"
	params := ListQueryParams{Page: 1, Limit: 10, MinAge: &minAge, Name: &name}
	users, total, err := repo.GetAll(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 user, got %d", total)
	}
	if len(users) != 1 || users[0].Name != "Annabelle" {
		t.Errorf("expected Annabelle, got %+v", users)
	}
}

func TestGetAll_InvalidPaginationDefaults(t *testing.T) {
	repo, cleanup := newTestRepo(t)
	defer cleanup()
	ctx := context.Background()
	_ = repo.Create(ctx, &user_model.User{Name: "Test", Email: "test@example.com", Age: 18})

	params := ListQueryParams{Page: 0, Limit: 0}
	users, total, err := repo.GetAll(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}
