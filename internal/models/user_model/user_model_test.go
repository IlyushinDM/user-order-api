package user_model_test

import (
	"testing"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model" // Нужен для поля Orders
	"github.com/IlyushinDM/user-order-api/internal/models/user_model"  // Путь к вашему пакету моделей пользователя
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Тест для структуры User
func TestUserStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры User
	now := time.Now()
	user := user_model.User{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		Age:          30,
		PasswordHash: "hashedpassword",
		Orders: []order_model.Order{ // Пример заполнения поля Orders
			{ID: 10, UserID: 1, ProductName: "Product A"},
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, uint(1), user.ID, "Поле ID должно быть доступно и иметь правильное значение")
	assert.Equal(t, "Test User", user.Name, "Поле Name должно быть доступно и иметь правильное значение")
	assert.Equal(t, "test@example.com", user.Email, "Поле Email должно быть доступно и иметь правильное значение")
	assert.Equal(t, 30, user.Age, "Поле Age должно быть доступно и иметь правильное значение")
	assert.Equal(t, "hashedpassword", user.PasswordHash, "Поле PasswordHash должно быть доступно и иметь правильное значение")
	assert.NotNil(t, user.Orders, "Поле Orders должно быть инициализировано")
	assert.Len(t, user.Orders, 1, "Поле Orders должно содержать список заказов")
	assert.Equal(t, uint(10), user.Orders[0].ID, "Элемент в списке Orders должен иметь правильное значение ID")
	assert.WithinDuration(t, now, user.CreatedAt, time.Second, "Поле CreatedAt должно быть доступно и иметь правильное значение")
	assert.WithinDuration(t, now, user.UpdatedAt, time.Second, "Поле UpdatedAt должно быть доступно и иметь правильное значение")
	// Проверка GORM-специфичных полей и тегов более уместна в тестах репозиториев.
}

// Тест для структуры UserResponse
func TestUserResponseStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры UserResponse
	response := user_model.UserResponse{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Age:   30,
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, uint(1), response.ID, "Поле ID должно быть доступно и иметь правильное значение")
	assert.Equal(t, "Test User", response.Name, "Поле Name должно быть доступно и иметь правильное значение")
	assert.Equal(t, "test@example.com", response.Email, "Поле Email должно быть доступно и иметь правильное значение")
	assert.Equal(t, 30, response.Age, "Поле Age должно быть доступно и иметь правильное значение")
	// JSON-теги проверяются при кодировании/декодировании JSON, что обычно делается в тестах обработчиков.
}

// Тест для структуры CreateUserRequest
func TestCreateUserRequestStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры CreateUserRequest
	request := user_model.CreateUserRequest{
		Name:     "New User",
		Email:    "new@example.com",
		Age:      25,
		Password: "securepassword",
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, "New User", request.Name, "Поле Name должно быть доступно и иметь правильное значение")
	assert.Equal(t, "new@example.com", request.Email, "Поле Email должно быть доступно и иметь правильное значение")
	assert.Equal(t, 25, request.Age, "Поле Age должно быть доступно и иметь правильное значение")
	assert.Equal(t, "securepassword", request.Password, "Поле Password должно быть доступно и иметь правильное значение")
	// Binding-теги проверяются в тестах обработчиков или кастомных валидаторов.
}

// Тест для структуры UpdateUserRequest
func TestUpdateUserRequestStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры UpdateUserRequest
	request := user_model.UpdateUserRequest{
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   35,
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, "Updated User", request.Name, "Поле Name должно быть доступно и иметь правильное значение")
	assert.Equal(t, "updated@example.com", request.Email, "Поле Email должно быть доступно и иметь правильное значение")
	assert.Equal(t, 35, request.Age, "Поле Age должно быть доступно и иметь правильное значение")

	// Проверяем случай с частичным обновлением
	partialRequest := user_model.UpdateUserRequest{
		Name: "Partial Update", // Обновляем только имя
	}
	assert.Equal(t, "Partial Update", partialRequest.Name, "Поле Name должно быть доступно и иметь правильное значение при частичном обновлении")
	assert.Equal(t, "", partialRequest.Email, "Поле Email должно быть пустым при частичном обновлении")
	assert.Equal(t, 0, partialRequest.Age, "Поле Age должно быть нулевым при частичном обновлении")
}

// Тест для структуры PaginatedUsersResponse
func TestPaginatedUsersResponseStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры PaginatedUsersResponse
	usersList := []user_model.UserResponse{
		{ID: 1, Name: "User A", Email: "a@example.com", Age: 20},
		{ID: 2, Name: "User B", Email: "b@example.com", Age: 22},
	}
	response := user_model.PaginatedUsersResponse{
		Page:  1,
		Limit: 10,
		Total: 50,
		Users: usersList,
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, 1, response.Page, "Поле Page должно быть доступно и иметь правильное значение")
	assert.Equal(t, 10, response.Limit, "Поле Limit должно быть доступно и иметь правильное значение")
	assert.Equal(t, int64(50), response.Total, "Поле Total должно быть доступно и иметь правильное значение")
	assert.Equal(t, len(usersList), len(response.Users), "Поле Users должно содержать список пользователей")
	assert.Equal(t, usersList, response.Users, "Список пользователей должен соответствовать ожидаемому")
}

// Тест для структуры LoginRequest
func TestLoginRequestStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры LoginRequest
	request := user_model.LoginRequest{
		Email:    "login@example.com",
		Password: "mypassword",
	}

	// Проверяем, что поля структуры доступны и имеют ожидаемые значения
	assert.Equal(t, "login@example.com", request.Email, "Поле Email должно быть доступно и иметь правильное значение")
	assert.Equal(t, "mypassword", request.Password, "Поле Password должно быть доступно и иметь правильное значение")
	// Binding-теги проверяются в тестах обработчиков или кастомных валидаторов.
}

// Тест для структуры LoginResponse
func TestLoginResponseStruct(t *testing.T) {
	// Проверяем создание и инициализацию структуры LoginResponse
	response := user_model.LoginResponse{
		Token: "some.jwt.token",
	}

	// Проверяем, что поле структуры доступно и имеет ожидаемое значение
	assert.Equal(t, "some.jwt.token", response.Token, "Поле Token должно быть доступно и иметь правильное значение")
	// JSON-теги проверяются при кодировании/декодировании JSON, что обычно делается в тестах обработчиков.
}
