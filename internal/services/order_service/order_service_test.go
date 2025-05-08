package order_service_test

import (
	"bytes" // Для отключения вывода логов логгера
	"context"
	"errors"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/models/order_model"     // Импортируем пакет моделей заказа
	"github.com/IlyushinDM/user-order-api/internal/models/user_model"      // Импортируем пакет моделей пользователя (нужен для мока userRepo)
	"github.com/IlyushinDM/user-order-api/internal/repository/order_db"    // Импортируем интерфейс репозитория заказа
	"github.com/IlyushinDM/user-order-api/internal/repository/user_db"     // Импортируем интерфейс репозитория пользователя
	"github.com/IlyushinDM/user-order-api/internal/services/order_service" // Импортируем тестируемый сервис
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockOrderRepository - мок реализация интерфейса order_db.OrderRepository.
// Используется для изоляции сервиса от реальной базы данных в модульных тестах.
type MockOrderRepository struct {
	mock.Mock
}

// Create имитирует метод Create репозитория заказа
func (m *MockOrderRepository) Create(ctx context.Context, order *order_model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// Update имитирует метод Update репозитория заказа
func (m *MockOrderRepository) Update(ctx context.Context, order *order_model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// Delete имитирует метод Delete репозитория заказа
func (m *MockOrderRepository) Delete(ctx context.Context, orderID uint, userID uint) error {
	args := m.Called(ctx, orderID, userID)
	return args.Error(0)
}

// GetByID имитирует метод GetByID репозитория заказа
func (m *MockOrderRepository) GetByID(ctx context.Context, orderID uint, userID uint) (*order_model.Order, error) {
	args := m.Called(ctx, orderID, userID)
	// Обрабатываем случай возврата nil для Order, если первый аргумент nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order_model.Order), args.Error(1)
}

// GetAllByUser имитирует метод GetAllByUser репозитория заказа
func (m *MockOrderRepository) GetAllByUser(ctx context.Context, userID uint, page, limit int) ([]order_model.Order, int64, error) {
	args := m.Called(ctx, userID, page, limit)
	// Обрабатываем случай возврата nil для списка Order
	var orders []order_model.Order
	if args.Get(0) != nil {
		orders = args.Get(0).([]order_model.Order)
	}
	return orders, args.Get(1).(int64), args.Error(2)
}

// MockUserRepository - мок реализация интерфейса user_db.UserRepository.
// Может потребоваться в сервисе заказа, например, для проверки существования пользователя.
type MockUserRepository struct {
	mock.Mock
}

// Методы мока UserRepository (только те, которые могут потребоваться в OrderService)
// В данном случае, в OrderService userRepo закомментирован, но мок все равно нужен
// для создания экземпляра orderService.
func (m *MockUserRepository) Create(ctx context.Context, user *user_model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) Update(ctx context.Context, user *user_model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*user_model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_model.User), args.Error(1)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_model.User), args.Error(1)
}
func (m *MockUserRepository) GetAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]user_model.User, int64, error) {
	args := m.Called(ctx, page, limit, filters)
	var users []user_model.User
	if args.Get(0) != nil {
		users = args.Get(0).([]user_model.User)
	}
	return users, args.Get(1).(int64), args.Error(2)
}

// setupService - вспомогательная функция для создания экземпляра OrderService с моками.
func setupService(orderRepo order_db.OrderRepository, userRepo user_db.UserRepository) order_service.OrderService {
	// Используем простой логгер для тестов и отключаем его вывод.
	log := logrus.New()
	log.SetOutput(bytes.NewBuffer(nil)) // Отключаем вывод логов в консоль во время тестов
	log.SetLevel(logrus.InfoLevel)
	return order_service.NewOrderService(orderRepo, userRepo, log)
}

// Тест для метода CreateOrder - успешное создание заказа.
func TestOrderService_CreateOrder(t *testing.T) {
	// Создаем моки репозиториев.
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	// Создаем тестируемый сервис.
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background() // Контекст для запроса.

	// Подготавливаем данные для запроса на создание.
	req := order_model.CreateOrderRequest{
		ProductName: "Test Product",
		Quantity:    10,
		Price:       99.99,
	}
	userID := uint(1) // ID пользователя, от имени которого создается заказ.

	// Настраиваем мок: ожидаем вызов Create в репозитории с любым объектом Order
	// (т.к. сервис создает его на основе req) и возвращаем nil (успех).
	mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*order_model.Order")).Return(nil).Once()

	// Вызываем тестируемый метод сервиса.
	order, err := service.CreateOrder(ctx, userID, req)

	// Проверяем результаты: отсутствие ошибки, ненулевой результат и соответствие полей.
	assert.NoError(t, err, "CreateOrder не должна возвращать ошибку при успешном создании")
	assert.NotNil(t, order, "CreateOrder должна вернуть созданный заказ")
	assert.Equal(t, userID, order.UserID, "ID пользователя в заказе должен соответствовать переданному")
	assert.Equal(t, req.ProductName, order.ProductName, "Имя продукта должно соответствовать запросу")
	assert.Equal(t, req.Quantity, order.Quantity, "Количество должно соответствовать запросу")
	assert.Equal(t, req.Price, order.Price, "Цена должна соответствовать запросу")

	// Проверяем, что настроенные методы моков были вызваны.
	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t) // Проверяем, даже если вызовов нет в этом тесте.
}

// Тест для метода CreateOrder - ошибка репозитория при создании.
func TestOrderService_CreateOrder_RepoError(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	req := order_model.CreateOrderRequest{
		ProductName: "Test Product",
		Quantity:    10,
		Price:       99.99,
	}
	userID := uint(1)
	repoError := errors.New("ошибка базы данных") // Имитируем ошибку репозитория.

	// Настраиваем мок: имитируем ошибку при вызове Create.
	mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*order_model.Order")).Return(repoError).Once()

	// Вызываем тестируемый метод.
	order, err := service.CreateOrder(ctx, userID, req)

	// Проверяем результаты: наличие ошибки и нулевой результат.
	assert.Error(t, err, "CreateOrder должна возвращать ошибку при ошибке репозитория")
	assert.Nil(t, order, "CreateOrder не должна возвращать заказ при ошибке")
	assert.Contains(t, err.Error(), "failed to save order", "Ошибка должна содержать информацию об источнике")
	assert.True(t, errors.Is(err, repoError), "Возвращенная ошибка должна оборачивать ошибку репозитория")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест для метода UpdateOrder - успешное обновление заказа.
func TestOrderService_UpdateOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(10)
	userID := uint(1)
	// Существующий заказ, который будет найден.
	existingOrder := &order_model.Order{
		ID:          orderID, // Инициализируем ID напрямую
		UserID:      userID,
		ProductName: "Old Product",
		Quantity:    5,
		Price:       50.00,
	}
	// Запрос на обновление с новыми значениями.
	req := order_model.UpdateOrderRequest{
		ProductName: "New Product",
		Quantity:    15,
		Price:       150.00,
	}

	// Настраиваем мок: ожидаем GetByID и возвращаем существующий заказ.
	mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(existingOrder, nil).Once()
	// Настраиваем мок: ожидаем Update с любым объектом Order (обновленным) и возвращаем nil (успех).
	// В реальном моке можно было бы проверить, что переданный объект Order содержит обновленные поля.
	mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*order_model.Order")).Return(nil).Once()

	// Вызываем тестируемый метод.
	updatedOrder, err := service.UpdateOrder(ctx, orderID, userID, req)

	// Проверяем результаты.
	assert.NoError(t, err, "UpdateOrder не должна возвращать ошибку при успешном обновлении")
	assert.NotNil(t, updatedOrder, "UpdateOrder должна вернуть обновленный заказ")
	assert.Equal(t, req.ProductName, updatedOrder.ProductName, "Имя продукта должно быть обновлено")
	assert.Equal(t, req.Quantity, updatedOrder.Quantity, "Количество должно быть обновлено")
	assert.Equal(t, req.Price, updatedOrder.Price, "Цена должна быть обновлена")
	assert.Equal(t, orderID, updatedOrder.ID, "ID заказа не должно меняться")
	assert.Equal(t, userID, updatedOrder.UserID, "ID пользователя не должно меняться")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест UpdateOrder, когда заказ не найден для данного пользователя.
func TestOrderService_UpdateOrder_NotFound(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(99) // Несуществующий ID.
	userID := uint(1)
	req := order_model.UpdateOrderRequest{ProductName: "New Product"}

	// Настраиваем мок: имитируем ошибку "запись не найдена" от GetByID.
	mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(nil, gorm.ErrRecordNotFound).Once()
	// Не ожидаем вызовов Update.

	// Вызываем тестируемый метод.
	updatedOrder, err := service.UpdateOrder(ctx, orderID, userID, req)

	// Проверяем результаты: наличие ошибки и совпадение с ожидаемой ошибкой.
	assert.Error(t, err, "UpdateOrder должна возвращать ошибку, если заказ не найден")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Ошибка должна быть gorm.ErrRecordNotFound")
	assert.Nil(t, updatedOrder, "UpdateOrder не должна возвращать заказ, если он не найден")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест UpdateOrder, когда нет полей для обновления в запросе.
func TestOrderService_UpdateOrder_NoFieldsToUpdate(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(10)
	userID := uint(1)
	// Существующий заказ.
	existingOrder := &order_model.Order{
		ID:          orderID, // Инициализируем ID напрямую
		UserID:      userID,
		ProductName: "Existing Product",
		Quantity:    5,
		Price:       50.00,
	}
	// Запрос на обновление, где все поля совпадают с существующими или не заданы.
	req := order_model.UpdateOrderRequest{
		ProductName: "Existing Product",  // Тот же продукт
		Quantity:    0,                   // Не задано (omitempty)
		Price:       existingOrder.Price, // Та же цена
	}

	// Настраиваем мок: ожидаем GetByID и возвращаем существующий заказ.
	mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(existingOrder, nil).Once()
	// Не ожидаем вызов Update, т.к. сервис должен определить отсутствие изменений.

	// Вызываем тестируемый метод.
	updatedOrder, err := service.UpdateOrder(ctx, orderID, userID, req)

	// Проверяем результаты: отсутствие ошибки и возврат исходного объекта заказа.
	assert.NoError(t, err, "UpdateOrder не должна возвращать ошибку при отсутствии изменений")
	assert.NotNil(t, updatedOrder, "UpdateOrder должна вернуть текущий заказ при отсутствии изменений")
	assert.Equal(t, existingOrder, updatedOrder, "Должен быть возвращен исходный объект заказа")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест UpdateOrder, когда репозиторий возвращает ошибку при обновлении.
func TestOrderService_UpdateOrder_RepoUpdateError(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(10)
	userID := uint(1)
	existingOrder := &order_model.Order{
		ID:          orderID, // Инициализируем ID напрямую
		UserID:      userID,
		ProductName: "Old Product",
		Quantity:    5,
		Price:       50.00,
	}
	req := order_model.UpdateOrderRequest{ProductName: "New Product"}
	repoError := errors.New("ошибка обновления в БД")

	// Настраиваем мок: ожидаем GetByID и возвращаем существующий заказ.
	mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(existingOrder, nil).Once()
	// Настраиваем мок: имитируем ошибку при вызове Update.
	mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*order_model.Order")).Return(repoError).Once()

	// Вызываем тестируемый метод.
	updatedOrder, err := service.UpdateOrder(ctx, orderID, userID, req)

	// Проверяем результаты: наличие ошибки.
	assert.Error(t, err, "UpdateOrder должна возвращать ошибку при ошибке репозитория")
	assert.Nil(t, updatedOrder, "UpdateOrder не должна возвращать заказ при ошибке репозитория")
	assert.Contains(t, err.Error(), "failed to save updated order", "Ошибка должна содержать информацию об источнике")
	assert.True(t, errors.Is(err, repoError), "Возвращенная ошибка должна оборачивать ошибку репозитория")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест для метода DeleteOrder - успешное удаление заказа.
func TestOrderService_DeleteOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(10)
	userID := uint(1)

	// Настраиваем мок: ожидаем вызов Delete в репозитории и имитируем успех.
	mockOrderRepo.On("Delete", ctx, orderID, userID).Return(nil).Once()

	// Вызываем тестируемый метод.
	err := service.DeleteOrder(ctx, orderID, userID)

	// Проверяем результат: отсутствие ошибки.
	assert.NoError(t, err, "DeleteOrder не должна возвращать ошибку при успешном удалении")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест DeleteOrder, когда заказ не найден или нет прав на удаление.
func TestOrderService_DeleteOrder_NotFoundOrPermissionDenied(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(99) // Несуществующий ID.
	userID := uint(1)
	// Имитируем ошибку "запись не найдена" или "нет прав" от репозитория.
	repoError := gorm.ErrRecordNotFound // Или ваш специфический текст ошибки от репозитория.
	mockOrderRepo.On("Delete", ctx, orderID, userID).Return(repoError).Once()

	// Вызываем тестируемый метод.
	err := service.DeleteOrder(ctx, orderID, userID)

	// Проверяем результат: наличие ошибки и совпадение с ожидаемой ошибкой.
	assert.Error(t, err, "DeleteOrder должна возвращать ошибку, если заказ не найден или нет прав")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Ошибка должна быть либо gorm.ErrRecordNotFound, либо специфической ошибкой репозитория")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест для метода GetOrderByID - успешное получение заказа.
func TestOrderService_GetOrderByID(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(10)
	userID := uint(1)
	// Ожидаемый объект заказа.
	expectedOrder := &order_model.Order{
		ID:     orderID, // Инициализируем ID напрямую
		UserID: userID, ProductName: "Some Product", Quantity: 5, Price: 50.00,
	}

	// Настраиваем мок: ожидаем вызов GetByID в репозитории и имитируем возврат заказа.
	mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(expectedOrder, nil).Once()

	// Вызываем тестируемый метод.
	order, err := service.GetOrderByID(ctx, orderID, userID)

	// Проверяем результаты.
	assert.NoError(t, err, "GetOrderByID не должна возвращать ошибку при успешном получении")
	assert.Equal(t, expectedOrder, order, "Полученный заказ должен соответствовать ожидаемому")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Тест GetOrderByID, когда заказ не найден для данного пользователя.
func TestOrderService_GetOrderByID_NotFound(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockUserRepo := new(MockUserRepository)
	service := setupService(mockOrderRepo, mockUserRepo)
	ctx := context.Background()

	orderID := uint(99) // Несуществующий ID.
	userID := uint(1)

	// Настраиваем мок: имитируем ошибку "запись не найдена" от репозитория.
	mockOrderRepo.On("GetByID", ctx, orderID, userID).Return(nil, gorm.ErrRecordNotFound).Once()

	// Вызываем тестируемый метод.
	order, err := service.GetOrderByID(ctx, orderID, userID)

	// Проверяем результаты.
	assert.Error(t, err, "GetOrderByID должна возвращать ошибку, если заказ не найден")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Ошибка должна быть gorm.ErrRecordNotFound")
	assert.Nil(t, order, "GetOrderByID не должна возвращать заказ, если он не найден")

	mockOrderRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// // Тест для метода GetAllOrdersByUser - успешное получение списка заказов.
// func TestOrderService_GetAllOrdersByUser(t *testing.T) {
// 	mockOrderRepo := new(MockOrderRepository)
// 	mockUserRepo := new(MockUserRepository)
// 	service := setupService(mockOrderRepo, mockUserRepo)
// 	ctx := context.Background()

// 	userID := uint(1)
// 	page := 1
// 	limit := 10
// 	// Ожидаемый список заказов и общее количество.
// 	expectedOrders := []order_model.Order{
// 		{ID: 11, UserID: userID, ProductName: "P1"}, // Инициализируем ID напрямую
// 		{ID: 12, UserID: userID, ProductName: "P2"}, // Инициализируем ID напрямую
// 	}
// 	expectedTotal := int64(2)

// 	// Настраиваем мок: ожидаем вызов GetAllByUser в репозитории и имитируем возврат списка.
// 	mockOrderRepo.On("GetAllByUser", ctx, userID, page, limit).Return(expectedOrders, expectedTotal, nil).Once()

// 	// Вызываем тестируемый метод.
// 	orders, total, err := service.GetAllOrdersByUser(ctx, userID, page, limit)

// 	// Проверяем результаты.
// 	assert.NoError(t, err, "GetAllOrdersByUser не должна возвращать ошибку при успешном получении")
// 	assert.Equal(t, expectedOrders, orders, "Полученный список заказов должен соответствовать ожидаемому")
// 	assert.Equal(t, expectedTotal, total, "Общее количество заказов должно соответствовать ожидаемому")

// 	mockOrderRepo.AssertExpectations(t)
// 	mockUserRepo.AssertExpectations(t)
// }

// // Тест GetAllOrdersByUser при ошибке репозитория.
// func TestOrderService_GetAllOrdersByUser_RepoError(t *testing.T) {
// 	mockOrderRepo := new(MockOrderRepository)
// 	mockUserRepo := new(MockUserRepository)
// 	service := setupService(mockOrderRepo, mockUserRepo)
// 	ctx := context.Background()

// 	userID := uint(1)
// 	page := 1
// 	limit := 10
// 	repoError := errors.New("ошибка при запросе списка")

// 	// Имитируем ошибку при вызове GetAllByUser.
// 	mockOrderRepo.On("GetAllByUser", ctx, userID, page, limit).Return(nil, int64(0), repoError).Once()

// 	// Вызываем тестируемый метод.
// 	orders, total, err := service.GetAllOrdersByUser(ctx, userID, page, limit)

// 	// Проверяем результаты.
// 	assert.Error(t, err, "GetAllOrdersByUser должна возвращать ошибку при ошибке репозитория")
// 	assert.Nil(t, orders, "GetAllOrdersByUser не должна возвращать список заказов при ошибке")
// 	assert.Equal(t, int64(0), total, "Общее количество должно быть 0 при ошибке")
// 	assert.Contains(t, err.Error(), "Failed to get orders for user from repository", "Ошибка должна содержать информацию об источнике")
// 	assert.True(t, errors.Is(err, repoError), "Возвращенная ошибка должна оборачивать ошибку репозитория")

// 	mockOrderRepo.AssertExpectations(t)
// 	mockUserRepo.AssertExpectations(t)
// }
