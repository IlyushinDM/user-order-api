package models

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestUserModelValidation(t *testing.T) {
	validate := validator.New()

	t.Run("Valid User", func(t *testing.T) {
		user := User{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			Age:   30,
		}
		err := validate.Struct(user)
		assert.NoError(t, err)
	})

	t.Run("Invalid User - Missing Name", func(t *testing.T) {
		user := User{
			Email: "john.doe@example.com",
			Age:   30,
		}
		err := validate.Struct(user)
		assert.Error(t, err)
	})

	t.Run("Invalid User - Invalid Email", func(t *testing.T) {
		user := User{
			Name:  "John Doe",
			Email: "invalid-email",
			Age:   30,
		}
		err := validate.Struct(user)
		assert.Error(t, err)
	})

	t.Run("Invalid User - Age less than 1", func(t *testing.T) {
		user := User{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			Age:   0,
		}
		err := validate.Struct(user)
		assert.Error(t, err)
	})

	t.Run("CreateUserRequest Validation", func(t *testing.T) {
		req := CreateUserRequest{
			Name:     "Jane Doe",
			Email:    "jane.doe@example.com",
			Age:      25,
			Password: "securePassword",
		}
		err := validate.Struct(req)
		assert.NoError(t, err)

		invalidReq := CreateUserRequest{
			Name:     "",
			Email:    "invalid-email",
			Age:      0,
			Password: "short",
		}
		err = validate.Struct(invalidReq)
		assert.Error(t, err)
	})

	t.Run("UpdateUserRequest Validation", func(t *testing.T) {
		req := UpdateUserRequest{
			Name:  "Jonathan Doe",
			Email: "jonathan.doe@example.com",
			Age:   31,
		}
		err := validate.Struct(req)
		assert.NoError(t, err)

		invalidReq := UpdateUserRequest{
			Email: "invalid-email",
			Age:   0,
		}
		err = validate.Struct(invalidReq)
		assert.Error(t, err)
	})

	t.Run("Test User timestamps", func(t *testing.T) {
		user := User{
			Name:      "Test User",
			Email:     "test.user@example.com",
			Age:       28,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Assert that CreatedAt and UpdatedAt are not zero
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	})
}
