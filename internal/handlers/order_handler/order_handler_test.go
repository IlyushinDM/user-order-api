package order_handler_test

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"

// 	// Импортируем тестируемый пакет и его зависимости
// 	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
// 	"github.com/IlyushinDM/user-order-api/internal/handlers/order_handler"
// 	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
// 	"github.com/IlyushinDM/user-order-api/internal/services/order_service"
// 	// Замените "your_module_path" на фактический путь к вашему модулю.
// )

// // MockOrderService представляет мок-реализацию интерфейса order_service.OrderService
// type MockOrderService struct {
// 	mock.Mock
// }

// func (m *MockOrderService) CreateOrder(ctx context.Context, userID uint, req order_model.CreateOrderRequest) (*order_model.Order, error) {
// 	args := m.Called(ctx, userID, req)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*order_model.Order), args.Error(1)
// }

// func (m *MockOrderService) GetOrderByID(ctx context.Context, orderID uint, userID uint) (*order_model.Order, error) {
// 	args := m.Called(ctx, orderID, userID)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*order_model.Order), args.Error(1)
// }

// func (m *MockOrderService) GetAllOrdersByUser(ctx context.Context, userID uint, page, limit int) ([]order_model.Order, int64, error) {
// 	args := m.Called(ctx, userID, page, limit)
// 	total := args.Int(1)
// 	if args.Get(0) == nil {
// 		return nil, int64(total), args.Error(2)
// 	}
// 	return args.Get(0).([]order_model.Order), int64(total), args.Error(2)
// }

// func (m *MockOrderService) UpdateOrder(ctx context.Context, orderID uint, userID uint, req order_model.UpdateOrderRequest) (*order_model.Order, error) {
// 	args := m.Called(ctx, orderID, userID, req)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*order_model.Order), args.Error(1)
// }

// func (m *MockOrderService) DeleteOrder(ctx context.Context, orderID uint, userID uint) error {
// 	args := m.Called(ctx, orderID, userID)
// 	return args.Error(0)
// }

// // setupTestGinContext создает mock Gin контекст и ResponseRecorder для тестирования.
// // Возвращает ResponseRecorder и Request. Engine создается отдельно для каждого теста,
// // чтобы избежать побочных эффектов между тестами.
// func setupTestGinContext(method, reqURL string, body io.Reader, userID ...uint) (*httptest.ResponseRecorder, *http.Request) {
// 	gin.SetMode(gin.TestMode) // Устанавливаем тестовый режим для Gin
// 	w := httptest.NewRecorder()

// 	req, _ := http.NewRequest(method, reqURL, body)
// 	if body != nil {
// 		req.Header.Set("Content-Type", "application/json")
// 	}

// 	// Важно: Gin контекст для роутера создается внутри router.ServeHTTP.
// 	// Мы не можем напрямую модифицировать этот контекст ДО вызова ServeHTTP.
// 	// Установка userID в контекст должна происходить либо в middleware (которое нужно будет мокировать/иммитировать),
// 	// либо в тестовой заглушке хендлера, если тестируется только логика ПОСЛЕ middleware.
// 	// Для простоты в тестах ниже userID устанавливается прямо в контекст, создаваемый ServeHTTP,
// 	// внутри тестовой функции-заглушки или сразу в начале реального хендлера,
// 	// если мы уверены, что middleware отработал.

// 	return w, req
// }

// // TestNewOrderHandler тестирует функцию создания экземпляра OrderHandler
// func TestNewOrderHandler(t *testing.T) {
// 	mockService := new(MockOrderService)
// 	// Используем реальный CommonHandler, так как NewOrderHandler ожидает *common_handler.CommonHandler
// 	realCommonHandler := common_handler.NewCommonHandler(logrus.New()) // Передаем логгер в CommonHandler
// 	mockLogger := logrus.New()

// 	// Тестовый случай: все зависимости предоставлены
// 	t.Run("Все зависимости предоставлены", func(t *testing.T) {
// 		handler := order_handler.NewOrderHandler(mockService, realCommonHandler, mockLogger)
// 		assert.NotNil(t, handler, "NewOrderHandler не должен возвращать nil при предоставлении всех зависимостей")
// 	})

// 	// Тестовый случай: orderService равен nil (ожидаем Fatal)
// 	t.Run("orderService nil", func(t *testing.T) {
// 		t.Skip("Тестирование logrus.Fatal требует более сложного мокирования или перехвата паники")
// 	})

// 	// Тестовый случай: commonHandler равен nil (ожидаем Fatal)
// 	t.Run("commonHandler nil", func(t *testing.T) {
// 		t.Skip("Тестирование logrus.Fatal требует более сложного мокирования или перехвата паники")
// 	})

// 	// Тестовый случай: log равен nil (ожидаем создание логгера по умолчанию и Warn)
// 	t.Run("log nil", func(t *testing.T) {
// 		// Проверяем, что handler создан, и логгер по умолчанию используется
// 		handler := order_handler.NewOrderHandler(mockService, realCommonHandler, nil)
// 		assert.NotNil(t, handler, "NewOrderHandler не должен возвращать nil при log nil")
// 		// Не можем напрямую проверить тип или вызов Warn без мокирования logrus
// 	})
// }

// // TestCheckUserIDMatch тестирует вспомогательную функцию checkUserIDMatch
// // Тестирование производится путем вызова публичного метода хендлера, который
// // использует checkUserIDMatch, и проверки результирующего HTTP статуса и тела ответа.
// // TestCheckUserIDMatch тестирует вспомогательную функцию checkUserIDMatch
// // Тестирование производится путем вызова публичного метода хендлера, который
// // использует checkUserIDMatch, и проверки результирующего HTTP статуса и тела ответа.
// func TestCheckUserIDMatch(t *testing.T) {
// 	mockService := new(MockOrderService)
// 	realCommonHandler := common_handler.NewCommonHandler(logrus.New())
// 	mockLogger := logrus.New()
// 	handler := order_handler.NewOrderHandler(mockService, realCommonHandler, mockLogger)

// 	authenticatedUserID := uint(123)

// 	// Настраиваем мок сервиса, чтобы он не прерывал выполнение (если проверка userID прошла).
// 	// Используем .Maybe() потому что сервис может быть не вызван, если checkUserIDMatch прервет контекст.
// 	mockService.On("GetOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(&order_model.Order{}, nil).Maybe()

// 	// Тестовый случай: userID не найден в контексте (имитируем отсутствие middleware)
// 	t.Run("userID не найден в контексте", func(t *testing.T) {
// 		router := gin.New()
// 		// Реальный хендлер handler.GetOrderByID будет вызван.
// 		// Он ожидает userID в контексте, но middleware его не добавит в этом подтесте.
// 		router.GET("/api/users/:id/orders/:orderID", handler.GetOrderByID)

// 		// Запрос без middleware, который устанавливает userID в контекст
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/1/orders/1", nil)

// 		// Запрос обрабатывается роутером
// 		router.ServeHTTP(w, req) // Gin Context создается здесь, userID не будет в нем

// 		// checkUserIDMatch в GetOrderByID должен обнаружить отсутствие userID в контексте
// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Ошибка аутентификации", responseBody.Error, "Ожидается сообщение об ошибке аутентификации")
// 	})

// 	// Тестовый случай: Неверный формат userID в URL
// 	t.Run("Неверный формат userID в URL", func(t *testing.T) {
// 		router := gin.New()
// 		router.Use(func(c *gin.Context) {
// 			// Имитируем middleware, устанавливая userID в контекст ДО вызова хендлера
// 			c.Set("userID", authenticatedUserID)
// 			c.Next()
// 		})
// 		router.GET("/api/users/:id/orders/:orderID", handler.GetOrderByID) // Вызываем реальный хендлер

// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/abc/orders/1", nil) // URL id = "abc"

// 		router.ServeHTTP(w, req)

// 		// checkUserIDMatch в GetOrderByID должен обнаружить неверный формат id в URL
// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Некорректный формат ID пользователя", responseBody.Error, "Ожидается сообщение об ошибке формата ID пользователя")
// 	})

// 	// Тестовый случай: userID из URL не совпадает с аутентифицированным userID из контекста
// 	t.Run("userID из URL не совпадает с context userID", func(t *testing.T) {
// 		router := gin.New()
// 		router.Use(func(c *gin.Context) {
// 			// Имитируем middleware
// 			c.Set("userID", authenticatedUserID) // Устанавливаем аутентифицированный ID = 123
// 			c.Next()
// 		})
// 		router.GET("/api/users/:id/orders/:orderID", handler.GetOrderByID) // Вызываем реальный хендлер

// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/999/orders/1", nil) // URL id = "999"

// 		router.ServeHTTP(w, req)

// 		// checkUserIDMatch должен обнаружить несовпадение
// 		assert.Equal(t, http.StatusForbidden, w.Code, "Ожидается статус 403 Forbidden")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Доступ запрещен", responseBody.Error, "Ожидается сообщение об ошибке доступа запрещен")
// 	})

// 	// Тестовый случай: userID из URL совпадает с аутентифицированным userID из контекста
// 	t.Run("userID из URL совпадает с context userID", func(t *testing.T) {
// 		router := gin.New()
// 		router.Use(func(c *gin.Context) {
// 			// Имитируем middleware
// 			c.Set("userID", authenticatedUserID) // Устанавливаем аутентифицированный ID = 123
// 			c.Next()
// 		})
// 		router.GET("/api/users/:id/orders/:orderID", handler.GetOrderByID) // Вызываем реальный хендлер

// 		// Настраиваем мок для GetOrderByID, т.к. мы тестируем его поведение после checkUserIDMatch
// 		// В этом случае сервис не вернет ошибку, и хендлер должен вернуть 404, т.к. заказ с ID 1 не найден
// 		// (это поведение тестируется в TestGetOrderByID, но мок нужен и здесь).
// 		mockService.On("GetOrderByID", mock.Anything, uint(1), authenticatedUserID).Return(nil, order_service.ErrOrderNotFound).Once()

// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/1", nil) // URL id = 123

// 		router.ServeHTTP(w, req)

// 		// checkUserIDMatch должен пройти. GetOrderByID должен вернуть 404 из-за ErrOrderNotFound.
// 		assert.Equal(t, http.StatusNotFound, w.Code, "Ожидается статус 404 Not Found после успешной проверки userID")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Заказ не найден", responseBody.Error, "Ожидается сообщение об ошибке 'Заказ не найден'")

// 		mockService.AssertExpectations(t) // Проверяем, что мок был вызван
// 	})

// 	// Тестовый случай: userID из URL не предоставлен (например, в роуте, который не имеет :id)
// 	// Этот сценарий сложно полностью протестировать через стандартный роутинг,
// 	// т.к. Gin не найдет маршрут без параметра :id. checkUserIDMatch сильно завязан
// 	// на c.Param("id"). Пропускаем этот тест, т.к. он требует мокирования c.Param
// 	// или изменения структуры маршрутов для теста.
// 	t.Run("userID из URL не предоставлен", func(t *testing.T) {
// 		t.Skip("Пропуск: Тестирование этого сценария требует мокирования c.Param или изменения роутинга для теста")
// 	})
// }

// // TestCreateOrder тестирует хендлер CreateOrder
// func TestCreateOrder(t *testing.T) {
// 	mockService := new(MockOrderService)
// 	realCommonHandler := common_handler.NewCommonHandler(logrus.New())
// 	mockLogger := logrus.New()
// 	handler := order_handler.NewOrderHandler(mockService, realCommonHandler, mockLogger)

// 	authenticatedUserID := uint(1)
// 	validCreateRequest := order_model.CreateOrderRequest{
// 		ProductName: "Test Product",
// 		Quantity:    2,
// 		Price:       10.5,
// 	}
// 	expectedOrderResponse := order_model.OrderResponse{
// 		ID:          1,
// 		UserID:      authenticatedUserID,
// 		ProductName: "Test Product",
// 		Quantity:    2,
// 		Price:       10.5,
// 	}
// 	createdOrder := order_model.Order{
// 		ID:          1,
// 		UserID:      authenticatedUserID,
// 		ProductName: "Test Product",
// 		Quantity:    2,
// 		Price:       10.5,
// 	}

// 	// Настраиваем роутер для корректной работы с URL параметрами и middleware
// 	router := gin.New()
// 	// Имитируем middleware, устанавливающее userID в контекст
// 	router.Use(func(c *gin.Context) {
// 		// В реальном приложении здесь была бы логика аутентификации/авторизации
// 		// Для теста просто устанавливаем предопределенный userID
// 		c.Set("userID", authenticatedUserID)
// 		c.Next() // Передаем управление следующему хендлеру
// 	})
// 	router.POST("/api/users/:id/orders", handler.CreateOrder)

// 	// Тестовый случай: Успешное создание заказа
// 	t.Run("Успешное создание заказа", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validCreateRequest)
// 		w, req := setupTestGinContext(http.MethodPost, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders", bytes.NewBuffer(jsonBody))

// 		// Настраиваем ожидание вызова сервиса и его результат
// 		mockService.On("CreateOrder", mock.Anything, authenticatedUserID, validCreateRequest).Return(&createdOrder, nil).Once()

// 		// Выполняем запрос через роутер
// 		router.ServeHTTP(w, req)

// 		// Проверяем статус ответа
// 		assert.Equal(t, http.StatusCreated, w.Code, "Ожидается статус 201 Created")

// 		// Проверяем тело ответа
// 		var actualResponse order_model.OrderResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
// 		assert.NoError(t, err, "Ошибка при парсинге JSON ответа")
// 		assert.Equal(t, expectedOrderResponse, actualResponse, "Тело ответа не соответствует ожидаемому")

// 		// Проверяем, что метод сервиса был вызван ровно один раз с ожидаемыми параметрами
// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: checkUserIDMatch возвращает false (например, ID пользователя не совпадает)
// 	t.Run("checkUserIDMatch fails", func(t *testing.T) {
// 		// Для этого теста нужно создать отдельный роутер без тестового middleware,
// 		// или имитировать ситуацию, когда middleware не установил userID,
// 		// или изменить URL так, чтобы checkUserIDMatch вернул ошибку.
// 		// Используем несовпадающие ID в URL и "имитируем" middleware, устанавливающее другой ID.
// 		routerFailCheck := gin.New()
// 		routerFailCheck.Use(func(c *gin.Context) {
// 			c.Set("userID", authenticatedUserID) // Устанавливаем userID = 1
// 			c.Next()
// 		})
// 		routerFailCheck.POST("/api/users/:id/orders", handler.CreateOrder)

// 		jsonBody, _ := json.Marshal(validCreateRequest)
// 		w, req := setupTestGinContext(http.MethodPost, "/api/users/999/orders", bytes.NewBuffer(jsonBody)) // URL id = 999

// 		// Сервис не должен быть вызван
// 		mockService.On("CreateOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("service should not be called")).Maybe()

// 		// Выполняем запрос через роутер, настроенный для проверки checkUserIDMatch
// 		routerFailCheck.ServeHTTP(w, req)

// 		// Проверяем статус ответа от checkUserIDMatch
// 		assert.Equal(t, http.StatusForbidden, w.Code, "Ожидается статус 403 Forbidden от checkUserIDMatch")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Доступ запрещен", responseBody.Error, "Ожидается сообщение об ошибке доступа запрещен от checkUserIDMatch")

// 		// Проверяем, что метод сервиса НЕ был вызван
// 		mockService.AssertNotCalled(t, "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Некорректный JSON в теле запроса
// 	t.Run("Некорректный JSON в теле запроса", func(t *testing.T) {
// 		invalidJsonBody := []byte(`{"product_name": "Test", "quantity": "два", "price": 10.5}`) // quantity - строка вместо числа
// 		w, req := setupTestGinContext(http.MethodPost, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders", bytes.NewBuffer(invalidJsonBody))

// 		// Сервис не должен быть вызван
// 		mockService.On("CreateOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("service should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request из-за некорректного JSON")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Некорректные входные данные", responseBody.Error, "Ожидается сообщение об ошибке некорректных входных данных")

// 		mockService.AssertNotCalled(t, "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrInvalidServiceInput
// 	t.Run("Сервис возвращает ErrInvalidServiceInput", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validCreateRequest)
// 		w, req := setupTestGinContext(http.MethodPost, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders", bytes.NewBuffer(jsonBody))

// 		serviceError := order_service.ErrInvalidServiceInput
// 		mockService.On("CreateOrder", mock.Anything, authenticatedUserID, validCreateRequest).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request при ErrInvalidServiceInput")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, serviceError.Error(), responseBody.Error, "Ожидается сообщение об ошибке из сервиса")

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrServiceDatabaseError
// 	t.Run("Сервис возвращает ErrServiceDatabaseError", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validCreateRequest)
// 		w, req := setupTestGinContext(http.MethodPost, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders", bytes.NewBuffer(jsonBody))

// 		serviceError := order_service.ErrServiceDatabaseError
// 		mockService.On("CreateOrder", mock.Anything, authenticatedUserID, validCreateRequest).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при ErrServiceDatabaseError")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Ошибка при создании заказа", responseBody.Error, "Ожидается сообщение об ошибке создания заказа")

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает неизвестную ошибку
// 	t.Run("Сервис возвращает неизвестную ошибку", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validCreateRequest)
// 		w, req := setupTestGinContext(http.MethodPost, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders", bytes.NewBuffer(jsonBody))

// 		serviceError := errors.New("какая-то другая ошибка сервиса")
// 		mockService.On("CreateOrder", mock.Anything, authenticatedUserID, validCreateRequest).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при неизвестной ошибке")

// 		var responseBody common_handler.ErrorResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "Внутренняя ошибка сервера", responseBody.Error, "Ожидается сообщение о внутренней ошибке сервера")

// 		mockService.AssertExpectations(t)
// 	})
// }

// // TestGetOrderByID тестирует хендлер GetOrderByID
// func TestGetOrderByID(t *testing.T) {
// 	mockService := new(MockOrderService)
// 	realCommonHandler := common_handler.NewCommonHandler(logrus.New())
// 	mockLogger := logrus.New()
// 	handler := order_handler.NewOrderHandler(mockService, realCommonHandler, mockLogger)

// 	authenticatedUserID := uint(1)
// 	orderID := uint(10)
// 	expectedOrder := &order_model.Order{
// 		ID:          orderID,
// 		UserID:      authenticatedUserID,
// 		ProductName: "Test Product",
// 		Quantity:    1,
// 		Price:       100.0,
// 	}
// 	expectedResponse := order_model.OrderResponse{
// 		ID:          orderID,
// 		UserID:      authenticatedUserID,
// 		ProductName: "Test Product",
// 		Quantity:    1,
// 		Price:       100.0,
// 	}

// 	// Настраиваем роутер с middleware
// 	router := gin.New()
// 	router.Use(func(c *gin.Context) {
// 		c.Set("userID", authenticatedUserID)
// 		c.Next()
// 	})
// 	router.GET("/api/users/:id/orders/:orderID", handler.GetOrderByID)

// 	// Тестовый случай: Успешное получение заказа
// 	t.Run("Успешное получение заказа", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		// Настраиваем ожидание вызова сервиса
// 		mockService.On("GetOrderByID", mock.Anything, orderID, authenticatedUserID).Return(expectedOrder, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code, "Ожидается статус 200 OK")
// 		var actualResponse order_model.OrderResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedResponse, actualResponse, "Тело ответа не соответствует ожидаемому")

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: checkUserIDMatch возвращает false
// 	t.Run("checkUserIDMatch fails", func(t *testing.T) {
// 		routerFailCheck := gin.New()
// 		routerFailCheck.Use(func(c *gin.Context) {
// 			c.Set("userID", authenticatedUserID) // userID = 1
// 			c.Next()
// 		})
// 		routerFailCheck.GET("/api/users/:id/orders/:orderID", handler.GetOrderByID)

// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/999/orders/"+strconv.FormatUint(uint64(orderID), 10), nil) // URL id = 999

// 		mockService.On("GetOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

// 		routerFailCheck.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusForbidden, w.Code, "Ожидается статус 403 Forbidden от checkUserIDMatch")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Доступ запрещен", responseBody.Error)

// 		mockService.AssertNotCalled(t, "GetOrderByID", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Неверный формат userID в URL
// 	t.Run("Неверный формат userID в URL", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/abc/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		mockService.On("GetOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректный формат ID пользователя", responseBody.Error)

// 		mockService.AssertNotCalled(t, "GetOrderByID", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Неверный формат orderID в URL
// 	t.Run("Неверный формат orderID в URL", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/abc", nil)

// 		mockService.On("GetOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректный формат ID заказа", responseBody.Error)

// 		mockService.AssertNotCalled(t, "GetOrderByID", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrOrderNotFound
// 	t.Run("Сервис возвращает ErrOrderNotFound", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := order_service.ErrOrderNotFound
// 		mockService.On("GetOrderByID", mock.Anything, orderID, authenticatedUserID).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusNotFound, w.Code, "Ожидается статус 404 Not Found")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Заказ не найден", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrInvalidServiceInput
// 	t.Run("Сервис возвращает ErrInvalidServiceInput", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := order_service.ErrInvalidServiceInput
// 		mockService.On("GetOrderByID", mock.Anything, orderID, authenticatedUserID).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, serviceError.Error(), responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrServiceDatabaseError
// 	t.Run("Сервис возвращает ErrServiceDatabaseError", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := order_service.ErrServiceDatabaseError
// 		mockService.On("GetOrderByID", mock.Anything, orderID, authenticatedUserID).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Ошибка при получении заказа", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает неизвестную ошибку
// 	t.Run("Сервис возвращает неизвестную ошибку", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := errors.New("какая-то другая ошибка сервиса")
// 		mockService.On("GetOrderByID", mock.Anything, orderID, authenticatedUserID).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Внутренняя ошибка сервера", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})
// }

// // TestGetAllOrdersByUser тестирует хендлер GetAllOrdersByUser
// func TestGetAllOrdersByUser(t *testing.T) {
// 	mockService := new(MockOrderService)
// 	realCommonHandler := common_handler.NewCommonHandler(logrus.New())
// 	mockLogger := logrus.New()
// 	handler := order_handler.NewOrderHandler(mockService, realCommonHandler, mockLogger)

// 	authenticatedUserID := uint(1)
// 	expectedOrders := []order_model.Order{
// 		{ID: 1, UserID: authenticatedUserID, ProductName: "Product A", Quantity: 1, Price: 10.0},
// 		{ID: 2, UserID: authenticatedUserID, ProductName: "Product B", Quantity: 2, Price: 20.0},
// 	}
// 	expectedTotal := int64(15) // Общее количество заказов для пагинации
// 	expectedResponse := order_model.PaginatedOrdersResponse{
// 		Page:  1,
// 		Limit: 10,
// 		Total: expectedTotal,
// 		Orders: []order_model.OrderResponse{
// 			{ID: 1, UserID: authenticatedUserID, ProductName: "Product A", Quantity: 1, Price: 10.0},
// 			{ID: 2, UserID: authenticatedUserID, ProductName: "Product B", Quantity: 2, Price: 20.0},
// 		},
// 	}

// 	// Настраиваем роутер с middleware
// 	router := gin.New()
// 	router.Use(func(c *gin.Context) {
// 		c.Set("userID", authenticatedUserID)
// 		c.Next()
// 	})
// 	router.GET("/api/users/:id/orders", handler.GetAllOrdersByUser)

// 	// Тестовый случай: Успешное получение списка заказов
// 	t.Run("Успешное получение списка заказов", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders?page=1&limit=10", nil)

// 		// Настраиваем ожидание вызова сервиса
// 		// Обратите внимание на соответствие типов: expectedTotal int64
// 		mockService.On("GetAllOrdersByUser", mock.Anything, authenticatedUserID, 1, 10).Return(expectedOrders, expectedTotal, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code, "Ожидается статус 200 OK")
// 		var actualResponse order_model.PaginatedOrdersResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedResponse, actualResponse, "Тело ответа не соответствует ожидаемому")

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: checkUserIDMatch возвращает false
// 	t.Run("checkUserIDMatch fails", func(t *testing.T) {
// 		routerFailCheck := gin.New()
// 		routerFailCheck.Use(func(c *gin.Context) {
// 			c.Set("userID", authenticatedUserID) // userID = 1
// 			c.Next()
// 		})
// 		routerFailCheck.GET("/api/users/:id/orders", handler.GetAllOrdersByUser)

// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/999/orders", nil) // URL id = 999

// 		mockService.On("GetAllOrdersByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("should not be called")).Maybe()

// 		routerFailCheck.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusForbidden, w.Code, "Ожидается статус 403 Forbidden от checkUserIDMatch")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Доступ запрещен", responseBody.Error)

// 		mockService.AssertNotCalled(t, "GetAllOrdersByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Неверный формат userID в URL
// 	t.Run("Неверный формат userID в URL", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/abc/orders", nil)

// 		mockService.On("GetAllOrdersByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректный формат ID пользователя", responseBody.Error)

// 		mockService.AssertNotCalled(t, "GetAllOrdersByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Неверные параметры пагинации (например, limit = 0)
// 	t.Run("Неверные параметры пагинации", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders?page=1&limit=0", nil)

// 		mockService.On("GetAllOrdersByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request из-за некорректной пагинации")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Неверные параметры пагинации", responseBody.Error) // Сообщение из кода хендлера

// 		mockService.AssertNotCalled(t, "GetAllOrdersByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrInvalidServiceInput
// 	t.Run("Сервис возвращает ErrInvalidServiceInput", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders?page=1&limit=10", nil)

// 		serviceError := order_service.ErrInvalidServiceInput
// 		mockService.On("GetAllOrdersByUser", mock.Anything, authenticatedUserID, mock.Anything, mock.Anything).Return(nil, int64(0), serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request при ErrInvalidServiceInput")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Неверные параметры запроса", responseBody.Error) // Сообщение из кода хендлера

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrServiceDatabaseError
// 	t.Run("Сервис возвращает ErrServiceDatabaseError", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders?page=1&limit=10", nil)

// 		serviceError := order_service.ErrServiceDatabaseError
// 		mockService.On("GetAllOrdersByUser", mock.Anything, authenticatedUserID, mock.Anything, mock.Anything).Return(nil, int64(0), serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при ErrServiceDatabaseError")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Ошибка при получении списка заказов", responseBody.Error) // Сообщение из кода хендлера

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает неизвестную ошибку
// 	t.Run("Сервис возвращает неизвестную ошибку", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders?page=1&limit=10", nil)

// 		serviceError := errors.New("какая-то другая ошибка сервиса")
// 		mockService.On("GetAllOrdersByUser", mock.Anything, authenticatedUserID, mock.Anything, mock.Anything).Return(nil, int64(0), serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при неизвестной ошибке")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Внутренняя ошибка сервера", responseBody.Error) // Сообщение из кода хендлера

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает пустой список заказов
// 	t.Run("Сервис возвращает пустой список заказов", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders?page=1&limit=10", nil)

// 		emptyOrders := []order_model.Order{}
// 		zeroTotal := int64(0)

// 		mockService.On("GetAllOrdersByUser", mock.Anything, authenticatedUserID, 1, 10).Return(emptyOrders, zeroTotal, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code, "Ожидается статус 200 OK")

// 		var actualResponse order_model.PaginatedOrdersResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 1, actualResponse.Page)
// 		assert.Equal(t, 10, actualResponse.Limit)
// 		assert.Equal(t, zeroTotal, actualResponse.Total)
// 		assert.Empty(t, actualResponse.Orders, "Список заказов должен быть пустым")

// 		mockService.AssertExpectations(t)
// 	})
// }

// // TestUpdateOrder тестирует хендлер UpdateOrder
// func TestUpdateOrder(t *testing.T) {
// 	mockService := new(MockOrderService)
// 	realCommonHandler := common_handler.NewCommonHandler(logrus.New())
// 	mockLogger := logrus.New()
// 	handler := order_handler.NewOrderHandler(mockService, realCommonHandler, mockLogger)

// 	authenticatedUserID := uint(1)
// 	orderID := uint(10)

// 	// Helper для получения указателя на строку, инт, флоат64 - ПЕРЕМЕЩЕНЫ ВНУТРЬ ТЕСТА
// 	// Эти функции нужны, если поля в UpdateOrderRequest являются указателями.
// 	stringPtr := func(s string) *string { return &s }
// 	intPtr := func(i int) *int { return &i }
// 	floatPtr := func(f float64) *float64 { return &f }

// 	validUpdateRequest := order_model.UpdateOrderRequest{
// 		ProductName: stringPtr("Updated Product Name"),
// 		Quantity:    intPtr(5),
// 		Price:       floatPtr(25.5),
// 	}
// 	updatedOrder := &order_model.Order{
// 		ID:          orderID,
// 		UserID:      authenticatedUserID,
// 		ProductName: "Updated Product Name",
// 		Quantity:    5,
// 		Price:       25.5,
// 	}
// 	expectedOrderResponse := order_model.OrderResponse{
// 		ID:          orderID,
// 		UserID:      authenticatedUserID,
// 		ProductName: "Updated Product Name",
// 		Quantity:    5,
// 		Price:       25.5,
// 	}

// 	// Настраиваем роутер с middleware
// 	router := gin.New()
// 	router.Use(func(c *gin.Context) {
// 		c.Set("userID", authenticatedUserID)
// 		c.Next()
// 	})
// 	router.PUT("/api/users/:id/orders/:orderID", handler.UpdateOrder)

// 	// Тестовый случай: Успешное обновление заказа
// 	t.Run("Успешное обновление заказа", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody))

// 		// Настраиваем ожидание вызова сервиса
// 		mockService.On("UpdateOrder", mock.Anything, orderID, authenticatedUserID, validUpdateRequest).Return(updatedOrder, nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code, "Ожидается статус 200 OK")

// 		var actualResponse order_model.OrderResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedOrderResponse, actualResponse, "Тело ответа не соответствует ожидаемому")

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: checkUserIDMatch возвращает false
// 	t.Run("checkUserIDMatch fails", func(t *testing.T) {
// 		routerFailCheck := gin.New()
// 		routerFailCheck.Use(func(c *gin.Context) {
// 			c.Set("userID", authenticatedUserID) // userID = 1
// 			c.Next()
// 		})
// 		routerFailCheck.PUT("/api/users/:id/orders/:orderID", handler.UpdateOrder)

// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/999/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody)) // URL id = 999

// 		mockService.On("UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

// 		routerFailCheck.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusForbidden, w.Code, "Ожидается статус 403 Forbidden от checkUserIDMatch")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Доступ запрещен", responseBody.Error)

// 		mockService.AssertNotCalled(t, "UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Некорректный формат userID в URL
// 	t.Run("Некорректный формат userID в URL", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/abc/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody))

// 		mockService.On("UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request из-за некорректного userID")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректный формат ID пользователя", responseBody.Error)

// 		mockService.AssertNotCalled(t, "UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Некорректный формат orderID в URL
// 	t.Run("Некорректный формат orderID в URL", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/abc", bytes.NewBuffer(jsonBody))

// 		mockService.On("UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request из-за некорректного orderID")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректный формат ID заказа", responseBody.Error)

// 		mockService.AssertNotCalled(t, "UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Некорректный JSON в теле запроса
// 	t.Run("Некорректный JSON в теле запроса", func(t *testing.T) {
// 		invalidJsonBody := []byte(`{"product_name": 123}`) // Неверный тип для product_name
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(invalidJsonBody))

// 		mockService.On("UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request из-за некорректного JSON")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректные данные для обновления", responseBody.Error)

// 		mockService.AssertNotCalled(t, "UpdateOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrOrderNotFound
// 	t.Run("Сервис возвращает ErrOrderNotFound", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody))

// 		serviceError := order_service.ErrOrderNotFound
// 		mockService.On("UpdateOrder", mock.Anything, orderID, authenticatedUserID, validUpdateRequest).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusNotFound, w.Code, "Ожидается статус 404 Not Found при ErrOrderNotFound")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Заказ не найден", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrNoUpdateFields
// 	t.Run("Сервис возвращает ErrNoUpdateFields", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest) // Используем тот же запрос, но сервис вернет NoUpdate
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody))

// 		// Для ErrNoUpdateFields сервис также возвращает текущее состояние объекта
// 		serviceError := order_service.ErrNoUpdateFields
// 		// Возвращаем текущее состояние заказа, как это делает сервис при ErrNoUpdateFields
// 		orderBeforeUpdate := &order_model.Order{
// 			ID:          orderID,
// 			UserID:      authenticatedUserID,
// 			ProductName: "Original Product", // Отличается от Updated
// 			Quantity:    1,                  // Отличается от 5
// 			Price:       5.99,               // Отличается от 25.5
// 		}
// 		expectedResponseForNoUpdate := order_model.OrderResponse{
// 			ID:          orderID,
// 			UserID:      authenticatedUserID,
// 			ProductName: "Original Product",
// 			Quantity:    1,
// 			Price:       5.99,
// 		}

// 		mockService.On("UpdateOrder", mock.Anything, orderID, authenticatedUserID, validUpdateRequest).Return(orderBeforeUpdate, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		// В случае ErrNoUpdateFields ожидается статус 200 OK и возвращается текущий объект
// 		assert.Equal(t, http.StatusOK, w.Code, "Ожидается статус 200 OK при ErrNoUpdateFields")
// 		var actualResponse order_model.OrderResponse
// 		err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedResponseForNoUpdate, actualResponse, "Тело ответа должно содержать текущее состояние заказа при ErrNoUpdateFields")

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrInvalidServiceInput
// 	t.Run("Сервис возвращает ErrInvalidServiceInput", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody))

// 		serviceError := order_service.ErrInvalidServiceInput
// 		mockService.On("UpdateOrder", mock.Anything, orderID, authenticatedUserID, validUpdateRequest).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request при ErrInvalidServiceInput")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, serviceError.Error(), responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrServiceDatabaseError
// 	t.Run("Сервис возвращает ErrServiceDatabaseError", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody))

// 		serviceError := order_service.ErrServiceDatabaseError
// 		mockService.On("UpdateOrder", mock.Anything, orderID, authenticatedUserID, validUpdateRequest).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при ErrServiceDatabaseError")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Ошибка при обновлении заказа", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает неизвестную ошибку
// 	t.Run("Сервис возвращает неизвестную ошибку", func(t *testing.T) {
// 		jsonBody, _ := json.Marshal(validUpdateRequest)
// 		w, req := setupTestGinContext(http.MethodPut, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), bytes.NewBuffer(jsonBody))

// 		serviceError := errors.New("какая-то другая ошибка сервиса")
// 		mockService.On("UpdateOrder", mock.Anything, orderID, authenticatedUserID, validUpdateRequest).Return(nil, serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при неизвестной ошибке")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Внутренняя ошибка сервера", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})
// }

// // TestDeleteOrder тестирует хендлер DeleteOrder
// func TestDeleteOrder(t *testing.T) {
// 	mockService := new(MockOrderService)
// 	realCommonHandler := common_handler.NewCommonHandler(logrus.New())
// 	mockLogger := logrus.New()
// 	handler := order_handler.NewOrderHandler(mockService, realCommonHandler, mockLogger)

// 	authenticatedUserID := uint(1)
// 	orderID := uint(10)

// 	// Настраиваем роутер с middleware
// 	router := gin.New()
// 	router.Use(func(c *gin.Context) {
// 		c.Set("userID", authenticatedUserID)
// 		c.Next()
// 	})
// 	router.DELETE("/api/users/:id/orders/:orderID", handler.DeleteOrder)

// 	// Тестовый случай: Успешное удаление заказа
// 	t.Run("Успешное удаление заказа", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		// Настраиваем ожидание вызова сервиса
// 		mockService.On("DeleteOrder", mock.Anything, orderID, authenticatedUserID).Return(nil).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusNoContent, w.Code, "Ожидается статус 204 No Content")
// 		assert.Empty(t, w.Body.Bytes(), "Тело ответа должно быть пустым")

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: checkUserIDMatch возвращает false
// 	t.Run("checkUserIDMatch fails", func(t *testing.T) {
// 		routerFailCheck := gin.New()
// 		routerFailCheck.Use(func(c *gin.Context) {
// 			c.Set("userID", authenticatedUserID) // userID = 1
// 			c.Next()
// 		})
// 		routerFailCheck.DELETE("/api/users/:id/orders/:orderID", handler.DeleteOrder)

// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/999/orders/"+strconv.FormatUint(uint64(orderID), 10), nil) // URL id = 999

// 		mockService.On("DeleteOrder", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("should not be called")).Maybe()

// 		routerFailCheck.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusForbidden, w.Code, "Ожидается статус 403 Forbidden от checkUserIDMatch")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Доступ запрещен", responseBody.Error)

// 		mockService.AssertNotCalled(t, "DeleteOrder", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Неверный формат userID в URL
// 	t.Run("Неверный формат userID в URL", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/abc/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		mockService.On("DeleteOrder", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request из-за некорректного userID")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректный формат ID пользователя", responseBody.Error)

// 		mockService.AssertNotCalled(t, "DeleteOrder", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Неверный формат orderID в URL
// 	t.Run("Неверный формат orderID в URL", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/abc", nil)

// 		mockService.On("DeleteOrder", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("should not be called")).Maybe()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request из-за некорректного orderID")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректный формат ID заказа", responseBody.Error)

// 		mockService.AssertNotCalled(t, "DeleteOrder", mock.Anything, mock.Anything, mock.Anything)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrOrderNotFound
// 	t.Run("Сервис возвращает ErrOrderNotFound", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := order_service.ErrOrderNotFound
// 		mockService.On("DeleteOrder", mock.Anything, orderID, authenticatedUserID).Return(serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusNotFound, w.Code, "Ожидается статус 404 Not Found при ErrOrderNotFound")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Заказ не найден", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrInvalidServiceInput
// 	t.Run("Сервис возвращает ErrInvalidServiceInput", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := order_service.ErrInvalidServiceInput
// 		mockService.On("DeleteOrder", mock.Anything, orderID, authenticatedUserID).Return(serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидается статус 400 Bad Request при ErrInvalidServiceInput")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Некорректные данные запроса", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает ErrServiceDatabaseError
// 	t.Run("Сервис возвращает ErrServiceDatabaseError", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := order_service.ErrServiceDatabaseError
// 		mockService.On("DeleteOrder", mock.Anything, orderID, authenticatedUserID).Return(serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при ErrServiceDatabaseError")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Ошибка при удалении заказа", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})

// 	// Тестовый случай: Сервис возвращает неизвестную ошибку
// 	t.Run("Сервис возвращает неизвестную ошибку", func(t *testing.T) {
// 		w, req := setupTestGinContext(http.MethodDelete, "/api/users/"+strconv.FormatUint(uint64(authenticatedUserID), 10)+"/orders/"+strconv.FormatUint(uint64(orderID), 10), nil)

// 		serviceError := errors.New("какая-то другая ошибка сервиса")
// 		mockService.On("DeleteOrder", mock.Anything, orderID, authenticatedUserID).Return(serviceError).Once()

// 		router.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code, "Ожидается статус 500 Internal Server Error при неизвестной ошибке")
// 		var responseBody common_handler.ErrorResponse
// 		json.Unmarshal(w.Body.Bytes(), &responseBody)
// 		assert.Equal(t, "Внутренняя ошибка сервера", responseBody.Error)

// 		mockService.AssertExpectations(t)
// 	})
// }
