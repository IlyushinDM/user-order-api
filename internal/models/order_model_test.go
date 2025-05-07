package models

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestOrderModelValidation(t *testing.T) {
	validate := validator.New()

	t.Run("Valid Order", func(t *testing.T) {
		order := Order{
			UserID:      1,
			ProductName: "Laptop",
			Quantity:    1,
			Price:       1200.50,
		}
		err := validate.Struct(order)
		assert.NoError(t, err)
	})

	t.Run("Invalid Order - Missing ProductName", func(t *testing.T) {
		order := Order{
			UserID:   1,
			Quantity: 1,
			Price:    1200.50,
		}
		err := validate.Struct(order)
		assert.Error(t, err)
	})

	t.Run("Invalid Order - Quantity less than 1", func(t *testing.T) {
		order := Order{
			UserID:      1,
			ProductName: "Laptop",
			Quantity:    0,
			Price:       1200.50,
		}
		err := validate.Struct(order)
		assert.Error(t, err)
	})

	t.Run("Invalid Order - Price less than 0", func(t *testing.T) {
		order := Order{
			UserID:      1,
			ProductName: "Laptop",
			Quantity:    1,
			Price:       -1200.50,
		}
		err := validate.Struct(order)
		assert.Error(t, err)
	})

	t.Run("CreateOrderRequest Validation", func(t *testing.T) {
		req := CreateOrderRequest{
			ProductName: "Keyboard",
			Quantity:    2,
			Price:       75.00,
		}
		err := validate.Struct(req)
		assert.NoError(t, err)

		invalidReq := CreateOrderRequest{
			ProductName: "",
			Quantity:    0,
			Price:       -1.00,
		}
		err = validate.Struct(invalidReq)
		assert.Error(t, err)
	})

	t.Run("UpdateOrderRequest Validation", func(t *testing.T) {
		req := UpdateOrderRequest{
			ProductName: "Mouse",
			Quantity:    3,
			Price:       25.00,
		}
		err := validate.Struct(req)
		assert.NoError(t, err)

		invalidReq := UpdateOrderRequest{
			Quantity: 0,
			Price:    -1.00,
		}
		err = validate.Struct(invalidReq)
		// Expecting errors on Quantity and Price
		assert.Error(t, err)
	})

	t.Run("Test Order timestamps", func(t *testing.T) {
		order := Order{
			UserID:      1,
			ProductName: "Monitor",
			Quantity:    1,
			Price:       300.00,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Assert that CreatedAt and UpdatedAt are not zero
		assert.NotZero(t, order.CreatedAt)
		assert.NotZero(t, order.UpdatedAt)
	})
}
