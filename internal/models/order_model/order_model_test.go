package order_model_test

import (
	"testing"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model" // Путь к вашему пакету моделей заказа
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Тест для структуры Order
func TestOrderStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры Order
	now := time.Now()
	order := order_model.Order{
		ID:          1,
		UserID:      101,
		ProductName: "Test Product",
		Quantity:    5,
		Price:       100.50,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   gorm.DeletedAt{}, // Пример инициализации для soft delete
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, uint(1), order.ID, "Поле ID должно быть доступно и иметь правильное значение")
	assert.Equal(t, uint(101), order.UserID, "Поле UserID должно быть доступно и иметь правильное значение")
	assert.Equal(t, "Test Product", order.ProductName, "Поле ProductName должно быть доступно и иметь правильное значение")
	assert.Equal(t, 5, order.Quantity, "Поле Quantity должно быть доступно и иметь правильное значение")
	assert.Equal(t, 100.50, order.Price, "Поле Price должно быть доступно и иметь правильное значение")
	assert.WithinDuration(t, now, order.CreatedAt, time.Second, "Поле CreatedAt должно быть доступно и иметь правильное значение")
	assert.WithinDuration(t, now, order.UpdatedAt, time.Second, "Поле UpdatedAt должно быть доступно и иметь правильное значение")
	// Проверка GORM-специфичных полей и тегов более уместна в тестах репозиториев.
}

// Тест для структуры OrderResponse
func TestOrderResponseStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры OrderResponse
	response := order_model.OrderResponse{
		ID:          1,
		UserID:      101,
		ProductName: "Test Product",
		Quantity:    5,
		Price:       100.50,
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, uint(1), response.ID, "Поле ID должно быть доступно и иметь правильное значение")
	assert.Equal(t, uint(101), response.UserID, "Поле UserID должно быть доступно и иметь правильное значение")
	assert.Equal(t, "Test Product", response.ProductName, "Поле ProductName должно быть доступно и иметь правильное значение")
	assert.Equal(t, 5, response.Quantity, "Поле Quantity должно быть доступно и иметь правильное значение")
	assert.Equal(t, 100.50, response.Price, "Поле Price должно быть доступно и иметь правильное значение")
	// JSON-теги проверяются при кодировании/декодировании JSON, что обычно делается в тестах обработчиков.
}

// Тест для структуры CreateOrderRequest
func TestCreateOrderRequestStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры CreateOrderRequest
	request := order_model.CreateOrderRequest{
		ProductName: "New Product",
		Quantity:    10,
		Price:       250.75,
		// UserID здесь закомментирован, как в модели, т.к. он выводится из JWT
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, "New Product", request.ProductName, "Поле ProductName должно быть доступно и иметь правильное значение")
	assert.Equal(t, 10, request.Quantity, "Поле Quantity должно быть доступно и иметь правильное значение")
	assert.Equal(t, 250.75, request.Price, "Поле Price должно быть доступно и иметь правильное значение")
	// Binding-теги проверяются в тестах обработчиков или кастомных валидаторов.
}

// Тест для структуры UpdateOrderRequest
func TestUpdateOrderRequestStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры UpdateOrderRequest
	request := order_model.UpdateOrderRequest{
		ProductName: "Updated Product",
		Quantity:    20,
		Price:       300.00,
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, "Updated Product", request.ProductName, "Поле ProductName должно быть доступно и иметь правильное значение")
	assert.Equal(t, 20, request.Quantity, "Поле Quantity должно быть доступно и иметь правильное значение")
	assert.Equal(t, 300.00, request.Price, "Поле Price должно быть доступно и иметь правильное значение")

	// Проверяем случай с частичным обновлением
	partialRequest := order_model.UpdateOrderRequest{
		Quantity: 25, // Обновляем только количество
	}
	assert.Equal(t, "", partialRequest.ProductName, "Поле ProductName должно быть пустым при частичном обновлении")
	assert.Equal(t, 25, partialRequest.Quantity, "Поле Quantity должно быть доступно и иметь правильное значение при частичном обновлении")
	assert.Equal(t, 0.0, partialRequest.Price, "Поле Price должно быть нулевым при частичном обновлении")
}

// Тест для структуры PaginatedOrdersResponse
func TestPaginatedOrdersResponseStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры PaginatedOrdersResponse
	ordersList := []order_model.OrderResponse{
		{ID: 1, UserID: 101, ProductName: "P1", Quantity: 1, Price: 10.0},
		{ID: 2, UserID: 101, ProductName: "P2", Quantity: 2, Price: 20.0},
	}
	response := order_model.PaginatedOrdersResponse{
		Page:   1,
		Limit:  10,
		Total:  25,
		Orders: ordersList,
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, 1, response.Page, "Поле Page должно быть доступно и иметь правильное значение")
	assert.Equal(t, 10, response.Limit, "Поле Limit должно быть доступно и иметь правильное значение")
	assert.Equal(t, int64(25), response.Total, "Поле Total должно быть доступно и иметь правильное значение")
	assert.Equal(t, len(ordersList), len(response.Orders), "Поле Orders должно содержать список заказов")
	assert.Equal(t, ordersList, response.Orders, "Список заказов должен соответствовать ожидаемому")
}
