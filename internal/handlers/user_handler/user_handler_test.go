package user_handler_test

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io" // Импортируем io для io.Discard
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
// 	"github.com/IlyushinDM/user-order-api/internal/handlers/user_handler"
// 	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
// 	"github.com/IlyushinDM/user-order-api/internal/services/user_service"
// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// // --- Моки Зависимостей ---

// // MockUserService мок структуры user_service.UserService
// // Теперь возвращает *user_model.User в CreateUser
// type MockUserService struct {
// 	mock.Mock
// }

// // CreateUser имитация вызова сервиса для создания пользователя
// // Сигнатура изменена для соответствия user_service.UserService
// func (m *MockUserService) CreateUser(ctx context.Context, req user_model.CreateUserRequest) (*user_model.User, error) {
// 	args := m.Called(ctx, req)
// 	// Проверяем тип возвращаемого значения перед приведением
// 	ret0 := args.Get(0)
// 	if ret0 == nil {
// 		return nil, args.Error(1)
// 	}
// 	return ret0.(*user_model.User), args.Error(1)
// }

// // GetUserByID имитация вызова сервиса для получения пользователя по ID
// func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*user_model.User, error) {
// 	args := m.Called(ctx, id)
// 	ret0 := args.Get(0)
// 	if ret0 == nil {
// 		return nil, args.Error(1)
// 	}
// 	return ret0.(*user_model.User), args.Error(1)
// }

// // GetAllUsers имитация вызова сервиса для получения всех пользователей с пагинацией и фильтрацией
// func (m *MockUserService) GetAllUsers(ctx context.Context, page, limit int, filters map[string]interface{}) ([]user_model.User, int, error) {
// 	args := m.Called(ctx, page, limit, filters)
// 	ret0 := args.Get(0)
// 	if ret0 == nil {
// 		return nil, args.Get(1).(int), args.Error(2)
// 	}
// 	return ret0.([]user_model.User), args.Get(1).(int), args.Error(2)
// }

// // UpdateUser имитация вызова сервиса для обновления пользователя
// func (m *MockUserService) UpdateUser(ctx context.Context, id uint, req user_model.UpdateUserRequest) (*user_model.User, error) {
// 	args := m.Called(ctx, id, req)
// 	ret0 := args.Get(0)
// 	if ret0 == nil {
// 		return nil, args.Error(1)
// 	}
// 	return ret0.(*user_model.User), args.Error(1)
// }

// // DeleteUser имитация вызова сервиса для удаления пользователя
// func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }

// // LoginUser имитация вызова сервиса для входа пользователя
// func (m *MockUserService) LoginUser(ctx context.Context, req user_model.LoginRequest) (string, error) {
// 	args := m.Called(ctx, req)
// 	return args.Get(0).(string), args.Error(1)
// }

// // MockCommonHandler - этот мок больше не используется для создания NewUserHandler,
// // но сохраняется как пример мокирования, если бы CommonHandler был интерфейсом.
// // Для NewUserHandler мы теперь передаем реальный *common_handler.CommonHandler.
// /*
// type MockCommonHandler struct {
// 	mock.Mock
// }

// func (m *MockCommonHandler) GetPaginationParams(c *gin.Context) (int, int, error) {
// 	args := m.Called(c)
// 	return args.Get(0).(int), args.Get(1).(int), args.Error(2)
// }

// func (m *MockCommonHandler) GetFilteringParams(c *gin.Context) (map[string]interface{}, error) {
// 	args := m.Called(c)
// 	return args.Get(0).(map[string]interface{}), args.Error(1)
// }
// */

// // MockLogger - этот мок больше не используется для создания NewUserHandler.
// // Мы используем реальный logrus.Logger, настроенный для тестов.
// // Методы оставлены, но без лишнего args := m.Called()
// /*
// type MockLogger struct {
// 	mock.Mock
// }

// func (m *MockLogger) WithContext(ctx context.Context) *logrus.Entry {
// 	// args := m.Called(ctx) // Удалено, т.к. args не используется
// 	entry := logrus.NewEntry(logrus.New())
// 	entry.Logger.SetOutput(io.Discard) // Отключаем вывод в тестах
// 	entry.Logger.SetLevel(logrus.PanicLevel) // Подавляем логирование
// 	return entry
// }

// func (m *MockLogger) WithField(key string, value interface{}) *logrus.Entry {
// 	// args := m.Called(key, value) // Удалено
// 	entry := logrus.NewEntry(logrus.New())
// 	entry.Logger.SetOutput(io.Discard)
// 	entry.Logger.SetLevel(logrus.PanicLevel)
// 	return entry
// }

// func (m *MockLogger) WithFields(fields logrus.Fields) *logrus.Entry {
// 	// args := m.Called(fields) // Удалено
// 	entry := logrus.NewEntry(logrus.New())
// 	entry.Logger.SetOutput(io.Discard)
// 	entry.Logger.SetLevel(logrus.PanicLevel)
// 	return entry
// }

// // Реализация методов логгирования - просто заглушки
// func (m *MockLogger) Debugf(format string, args ...interface{}) {}
// func (m *MockLogger) Infof(format string, args ...interface{})  {}
// func (m *MockLogger) Printf(format string, args ...interface{}) {}
// func (m *MockLogger) Warnf(format string, args ...interface{})  {}
// func (m *MockLogger) Errorf(format string, args ...interface{}) {}
// func (m *MockLogger) Fatalf(format string, args ...interface{}) {} // Note: Fatalf will still os.Exit in real logrus
// func (m *MockLogger) Panicf(format string, args ...interface{}) {} // Note: Panicf will still panic in real logrus
// func (m *MockLogger) Debug(args ...interface{}) {}
// func (m *MockLogger) Info(args ...interface{})  {}
// func (m *MockLogger) Print(args ...interface{}) {}
// func (m *MockLogger) Warn(args ...interface{})  {}
// func (m *MockLogger) Error(args ...interface{}) {}
// func (m *MockLogger) Fatal(args ...interface{}) {} // Note: Fatal will still os.Exit in real logrus
// func (m *MockLogger) Panic(args ...interface{}) {} // Note: Panic will still panic in real logrus
// func (m *MockLogger) Debugln(args ...interface{}) {}
// func (m *MockLogger) Infoln(args ...interface{})  {}
// func (m *MockLogger) Println(args ...interface{}) {}
// func (m *MockLogger) Warnln(args ...interface{})  {}
// func (m *MockLogger) Errorln(args ...interface{}) {}
// func (m *MockLogger) Fatalln(args ...interface{}) {} // Note: Fatalln will still os.Exit in real logrus
// func (m *MockLogger) Panicln(args ...interface{}) {} // Note: Panicln will still panic in real logrus

// func (m *MockLogger) WithError(err error) *logrus.Entry {
// 	// args := m.Called(err) // Удалено
// 	entry := logrus.NewEntry(logrus.New())
// 	entry.Logger.SetOutput(io.Discard)
// 	entry.Logger.SetLevel(logrus.PanicLevel)
// 	return entry
// }
// */

// // --- Вспомогательные Функции для Тестов ---

// // setupTestRouter настраивает роутер Gin с заданным хендлером и методом/путем.
// // Возвращает контекст Gin и ResponseRecorder.
// // Теперь принимает *real* common_handler.CommonHandler и *real* logrus.Logger.
// func setupTestRouter(method, path string, handlerFunc gin.HandlerFunc, body interface{}, authUserID uint, common *common_handler.CommonHandler, logger *logrus.Logger) (*gin.Context, *httptest.ResponseRecorder) {
// 	gin.SetMode(gin.ReleaseMode)

// 	w := httptest.NewRecorder()
// 	c, router := gin.CreateTestContext(w)

// 	var reqBody *bytes.Buffer
// 	if body != nil {
// 		reqBodyBytes, _ := json.Marshal(body)
// 		reqBody = bytes.NewBuffer(reqBodyBytes)
// 		c.Request, _ = http.NewRequest(method, path, reqBody)
// 		c.Request.Header.Set("Content-Type", "application/json")
// 	} else {
// 		c.Request, _ = http.NewRequest(method, path, nil)
// 	}

// 	c.Request = c.Request.WithContext(context.Background())

// 	if authUserID != 0 {
// 		c.Set("userID", authUserID)
// 	}

// 	// Добавляем хендлер к роутеру
// 	testPath := path
// 	// Более надежный способ обработки параметров пути для тестов
// 	if idStr, ok := getIDFromPath(path); ok {
// 		// Если в пути есть ID, добавляем параметр вручную
// 		c.Params = append(c.Params, gin.Param{Key: "id", Value: idStr})
// 		// И устанавливаем путь роута с плейсхолдером :id
// 		// Определяем базовый путь без ID
// 		parts := splitPath(path)
// 		if len(parts) > 0 {
// 			testPath = "/" + parts[0] // Добавляем первый сегмент, например "api"
// 			if len(parts) > 1 {
// 				testPath += "/" + parts[1] // Добавляем второй сегмент, например "users"
// 			}
// 		}
// 		testPath += "/:id" // Добавляем плейсхолдер
// 	}

// 	router.Handle(method, testPath, handlerFunc)

// 	// Выполняем запрос
// 	router.ServeHTTP(w, c.Request)

// 	return c, w
// }

// // getIDFromPath извлекает ID из пути, если он есть (скопировано из оригинала)
// func getIDFromPath(path string) (string, bool) {
// 	parts := splitPath(path)
// 	if len(parts) > 0 {
// 		lastSegment := parts[len(parts)-1]
// 		if _, err := parseUint(lastSegment); err == nil {
// 			return lastSegment, true
// 		}
// 	}
// 	return "", false
// }

// // splitPath разделяет путь на сегменты (скопировано из оригинала)
// func splitPath(path string) []string {
// 	parts := []string{}
// 	current := ""
// 	for _, r := range path {
// 		if r == '/' {
// 			if current != "" {
// 				parts = append(parts, current)
// 			}
// 			current = ""
// 		} else {
// 			current += string(r)
// 		}
// 	}
// 	if current != "" {
// 		parts = append(parts, current)
// 	}
// 	return parts
// }

// // parseUint безопасное преобразование строки в uint (скопировано из оригинала)
// func parseUint(s string) (uint, error) {
// 	var i uint64
// 	_, err := fmt.Sscan(s, &i)
// 	if err != nil {
// 		return 0, err
// 	}
// 	// Дополнительная проверка, что значение умещается в uint (uint32 на 32-бит системах)
// 	if i > uint62 { // Используем uint62 как константу для максимального uint, чтобы избежать проблем с размером int на разных архитектурах
// 		return 0, errors.New("value out of range for uint")
// 	}
// 	return uint(i), nil
// }

// const uint62 = (^uint(0) >> 1) // Максимальное значение для uint

// // Вспомогательные функции для получения указателей на примитивные типы
// func strPtr(s string) *string { return &s }
// func intPtr(i int) *int       { return &i }

// // --- Тесты для NewUserHandler ---

// func TestNewUserHandler(t *testing.T) {
// 	mockService := new(MockUserService)
// 	// Создаем реальные экземпляры для commonHandler и logger
// 	// В реальных тестах commonHandler может иметь свои зависимости, которые нужно мокать.
// 	// Здесь мы просто создаем его с минимальными зависимостями (например, логгером).
// 	testLogger := logrus.New()
// 	testLogger.SetOutput(io.Discard)       // Отправляем вывод логгера в никуда во время тестов
// 	testLogger.SetLevel(logrus.PanicLevel) // Отключаем все логи, кроме паники, чтобы тесты не висели

// 	// Если NewCommonHandler требует логгер, его нужно передать
// 	// Т.к. определения common_handler нет, предполагаем, что можно создать пустой
// 	// Если CommonHandler имеет сложную инициализацию, нужно будет создать мок ИНТЕРФЕЙСА CommonHandler,
// 	// или создать реальный экземпляр с моками его зависимостей.
// 	// Исходя из ошибки, NewUserHandler ожидает *common_handler.CommonHandler (конкретный тип).
// 	// Значит, нужно передать реальный *common_handler.CommonHandler.
// 	// Предполагаем, что common_handler.NewCommonHandler существует и принимает логгер.
// 	// Если NewCommonHandler не принимает логгер, просто создаем &common_handler.CommonHandler{}
// 	// commonHandler := &common_handler.CommonHandler{} // Если нет конструктора или он простой
// 	commonHandler := common_handler.NewCommonHandler(testLogger) // Предполагаем такой конструктор

// 	t.Run("Успешное создание с валидными зависимостями", func(t *testing.T) {
// 		handler := user_handler.NewUserHandler(mockService, commonHandler, testLogger)
// 		assert.NotNil(t, handler, "Хендлер должен быть создан")
// 	})

// 	t.Run("Паника при nil UserService", func(t *testing.T) {
// 		defer func() {
// 			if r := recover(); r == nil {
// 				t.Errorf("Ожидалась паника при nil UserService")
// 			}
// 		}()
// 		// Используем реальные, но пустые или тестовые экземпляры для других зависимостей
// 		user_handler.NewUserHandler(nil, commonHandler, testLogger)
// 	})

// 	t.Run("Паника при nil CommonHandler", func(t *testing.T) {
// 		defer func() {
// 			if r := recover(); r == nil {
// 				t.Errorf("Ожидалась паника при nil CommonHandler")
// 			}
// 		}()
// 		user_handler.NewUserHandler(mockService, nil, testLogger)
// 	})

// 	t.Run("Использование дефолтного логгера при nil Logger", func(t *testing.T) {
// 		// Этот тест проверяет, что не происходит паники и возвращается не nil хендлер.
// 		// Проверить, что используется именно дефолтный логгер, сложнее без мокирования logrus.New().
// 		handler := user_handler.NewUserHandler(mockService, commonHandler, nil)
// 		assert.NotNil(t, handler, "Хендлер должен быть создан даже с nil логгером")
// 	})
// }

// // --- Тесты для CreateUser ---

// func TestCreateUser(t *testing.T) {
// 	gin.SetMode(gin.ReleaseMode)

// 	mockService := new(MockUserService)
// 	testLogger := logrus.New()
// 	testLogger.SetOutput(io.Discard)
// 	testLogger.SetLevel(logrus.PanicLevel)
// 	commonHandler := common_handler.NewCommonHandler(testLogger)
// 	handler := user_handler.NewUserHandler(mockService, commonHandler, testLogger)

// 	tests := []struct {
// 		name               string
// 		requestBody        user_model.CreateUserRequest
// 		mockServiceReturn  *user_model.User // Теперь ожидаем указатель
// 		mockServiceError   error
// 		expectedStatusCode int
// 		expectedBody       interface{}
// 		expectServiceCall  bool
// 	}{
// 		{
// 			name: "Успешное создание пользователя",
// 			requestBody: user_model.CreateUserRequest{
// 				Name:     "Test User",
// 				Email:    "test@example.com",
// 				Age:      30,
// 				Password: "password123",
// 			},
// 			mockServiceReturn: &user_model.User{ // Передаем указатель
// 				ID:    1,
// 				Name:  "Test User",
// 				Email: "test@example.com",
// 				Age:   30,
// 			},
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusCreated,
// 			expectedBody: user_model.UserResponse{
// 				ID:    1,
// 				Name:  "Test User",
// 				Email: "test@example.com",
// 				Age:   30,
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Некорректный формат запроса (JSON bind error)",
// 			requestBody: user_model.CreateUserRequest{
// 				Name:     "Test User",
// 				Email:    "test@example.com",
// 				Age:      30,
// 				Password: "password123",
// 			},
// 			mockServiceReturn:  nil, // Сервис не будет вызван
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Invalid input data",
// 			},
// 			expectServiceCall: false,
// 		},
// 		{
// 			name: "Ошибка сервиса: некорректные входные данные",
// 			requestBody: user_model.CreateUserRequest{
// 				Name:     "",
// 				Email:    "invalid-email",
// 				Age:      -5,
// 				Password: "",
// 			},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrInvalidServiceInput,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error:   "Invalid input data",
// 				Details: user_service.ErrInvalidServiceInput.Error(),
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: пользователь уже существует",
// 			requestBody: user_model.CreateUserRequest{
// 				Name:     "Existing User",
// 				Email:    "existing@example.com",
// 				Age:      25,
// 				Password: "password",
// 			},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrUserAlreadyExists,
// 			expectedStatusCode: http.StatusConflict,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: user_service.ErrUserAlreadyExists.Error(),
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: внутренняя ошибка",
// 			requestBody: user_model.CreateUserRequest{
// 				Name:     "Internal Error User",
// 				Email:    "internal@example.com",
// 				Age:      40,
// 				Password: "password",
// 			},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrInternalServiceError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Internal server error",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: ошибка базы данных",
// 			requestBody: user_model.CreateUserRequest{
// 				Name:     "Database Error User",
// 				Email:    "db@example.com",
// 				Age:      50,
// 				Password: "password",
// 			},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrServiceDatabaseError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Database operation failed",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: неизвестная ошибка",
// 			requestBody: user_model.CreateUserRequest{
// 				Name:     "Unknown Error User",
// 				Email:    "unknown@example.com",
// 				Age:      60,
// 				Password: "password",
// 			},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   errors.New("some unexpected error"),
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Failed to create user",
// 			},
// 			expectServiceCall: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockService.Calls = []mock.Call{}

// 			if tt.expectServiceCall {
// 				// Передаем указатель в Return
// 				mockService.On("CreateUser", mock.Anything, tt.requestBody).Return(tt.mockServiceReturn, tt.mockServiceError).Once()
// 			}

// 			var reqBody interface{} = tt.requestBody
// 			if tt.name == "Некорректный формат запроса (JSON bind error)" {
// 				reqBody = "invalid json string"
// 			}

// 			_, w := setupTestRouter(http.MethodPost, "/api/users", handler.CreateUser, reqBody, 0, commonHandler, testLogger) // Передаем real commonHandler и testLogger

// 			assert.Equal(t, tt.expectedStatusCode, w.Code, "Неверный статус код")

// 			var actualBody interface{}
// 			if w.Body.Len() > 0 {
// 				if tt.expectedStatusCode >= 400 {
// 					actualBody = &common_handler.ErrorResponse{}
// 				} else {
// 					actualBody = &user_model.UserResponse{}
// 				}
// 				err := json.Unmarshal(w.Body.Bytes(), actualBody)
// 				assert.NoError(t, err, "Ошибка десериализации тела ответа")
// 			} else {
// 				actualBody = nil
// 			}

// 			if tt.name == "Некорректный формат запроса (JSON bind error)" && tt.expectedStatusCode == http.StatusBadRequest {
// 				expectedErrResp, ok := tt.expectedBody.(common_handler.ErrorResponse)
// 				assert.True(t, ok, "Ожидалось тело типа ErrorResponse")
// 				actualErrResp, ok := actualBody.(*common_handler.ErrorResponse)
// 				assert.True(t, ok, "Фактическое тело имеет тип ErrorResponse")
// 				assert.Equal(t, expectedErrResp.Error, actualErrResp.Error, "Неверное сообщение об ошибке")
// 				assert.NotEmpty(t, actualErrResp.Details, "Поле Details не должно быть пустым при ошибке BindJSON")
// 			} else {
// 				assert.Equal(t, tt.expectedBody, actualBody, "Неверное тело ответа")
// 			}

// 			mockService.AssertExpectations(t)
// 		})
// 	}
// }

// // --- Тесты для GetUserByID ---

// func TestGetUserByID(t *testing.T) {
// 	gin.SetMode(gin.ReleaseMode)

// 	mockService := new(MockUserService)
// 	testLogger := logrus.New()
// 	testLogger.SetOutput(io.Discard)
// 	testLogger.SetLevel(logrus.PanicLevel)
// 	commonHandler := common_handler.NewCommonHandler(testLogger)
// 	handler := user_handler.NewUserHandler(mockService, commonHandler, testLogger)

// 	tests := []struct {
// 		name               string
// 		userIDParam        string
// 		mockServiceReturn  *user_model.User // Теперь ожидаем указатель
// 		mockServiceError   error
// 		expectedStatusCode int
// 		expectedBody       interface{}
// 		expectServiceCall  bool
// 	}{
// 		{
// 			name:               "Успешное получение пользователя по ID",
// 			userIDParam:        "123",
// 			mockServiceReturn:  &user_model.User{ID: 123, Name: "Fetched User", Email: "fetched@example.com", Age: 28}, // Передаем указатель
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusOK,
// 			expectedBody: user_model.UserResponse{
// 				ID:    123,
// 				Name:  "Fetched User",
// 				Email: "fetched@example.com",
// 				Age:   28,
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name:               "Некорректный формат ID в пути",
// 			userIDParam:        "abc",
// 			mockServiceReturn:  nil,
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Invalid user ID format",
// 			},
// 			expectServiceCall: false,
// 		},
// 		{
// 			name:               "Ошибка сервиса: пользователь не найден",
// 			userIDParam:        "404",
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrUserNotFound,
// 			expectedStatusCode: http.StatusNotFound,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: user_service.ErrUserNotFound.Error(),
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: некорректный ID (например, 0) -> маппится в 404",
// 			userIDParam:        "0",
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrInvalidServiceInput,
// 			expectedStatusCode: http.StatusNotFound,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: user_service.ErrUserNotFound.Error(),
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: ошибка базы данных",
// 			userIDParam:        "500",
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrServiceDatabaseError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Database operation failed",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: неизвестная ошибка",
// 			userIDParam:        "999",
// 			mockServiceReturn:  nil,
// 			mockServiceError:   errors.New("some other service error"),
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Failed to retrieve user",
// 			},
// 			expectServiceCall: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockService.Calls = []mock.Call{}

// 			userIDUint, parseErr := parseUint(tt.userIDParam)

// 			if tt.expectServiceCall && parseErr == nil {
// 				// Передаем указатель в Return
// 				mockService.On("GetUserByID", mock.Anything, userIDUint).Return(tt.mockServiceReturn, tt.mockServiceError).Once()
// 			}

// 			path := "/api/users/" + tt.userIDParam
// 			_, w := setupTestRouter(http.MethodGet, path, handler.GetUserByID, nil, 0, commonHandler, testLogger) // Передаем real commonHandler и testLogger

// 			assert.Equal(t, tt.expectedStatusCode, w.Code, "Неверный статус код")

// 			var actualBody interface{}
// 			if w.Body.Len() > 0 {
// 				if tt.expectedStatusCode >= 400 {
// 					actualBody = &common_handler.ErrorResponse{}
// 				} else {
// 					actualBody = &user_model.UserResponse{}
// 				}
// 				err := json.Unmarshal(w.Body.Bytes(), actualBody)
// 				assert.NoError(t, err, "Ошибка десериализации тела ответа")
// 			} else {
// 				actualBody = nil
// 			}

// 			assert.Equal(t, tt.expectedBody, actualBody, "Неверное тело ответа")

// 			mockService.AssertExpectations(t)
// 		})
// 	}
// }

// // --- Тесты для GetAllUsers ---

// func TestGetAllUsers(t *testing.T) {
// 	gin.SetMode(gin.ReleaseMode)

// 	mockService := new(MockUserService)
// 	testLogger := logrus.New()
// 	testLogger.SetOutput(io.Discard)
// 	testLogger.SetLevel(logrus.PanicLevel)
// 	// Для этого теста commonHandler используется для GetPaginationParams и GetFilteringParams.
// 	// Поскольку NewUserHandler ожидает *real* common_handler.CommonHandler, мы передаем его.
// 	// Методы GetPaginationParams и GetFilteringParams в реальном commonHandler должны быть реализованы
// 	// или в CommonHandlerTest вы должны мокать зависимости CommonHandler.
// 	// Здесь предполагается, что NewCommonHandler(logger) создает рабочую CommonHandler.
// 	mockCommon := common_handler.NewCommonHandler(testLogger) // Используем реальный CommonHandler

// 	handler := user_handler.NewUserHandler(mockService, mockCommon, testLogger)

// 	tests := []struct {
// 		name               string
// 		requestPath        string // /api/users?page=...&limit=...&min_age=... etc.
// 		mockServiceReturn  []user_model.User
// 		mockServiceTotal   int
// 		mockServiceError   error
// 		expectedStatusCode int
// 		expectedBody       interface{}
// 		expectServiceCall  bool
// 	}{
// 		{
// 			name:               "Успешное получение всех пользователей с пагинацией",
// 			requestPath:        "/api/users?page=1&limit=10",
// 			mockServiceReturn:  []user_model.User{{ID: 1, Name: "User 1"}, {ID: 2, Name: "User 2"}},
// 			mockServiceTotal:   100,
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusOK,
// 			expectedBody: user_model.PaginatedUsersResponse{
// 				Page:  1,
// 				Limit: 10,
// 				Total: 100,
// 				Users: []user_model.UserResponse{{ID: 1, Name: "User 1"}, {ID: 2, Name: "User 2"}},
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name:               "Успешное получение всех пользователей с пагинацией и фильтром по возрасту",
// 			requestPath:        "/api/users?page=2&limit=5&min_age=20&max_age=40",
// 			mockServiceReturn:  []user_model.User{{ID: 3, Name: "User 3", Age: 25}},
// 			mockServiceTotal:   5,
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusOK,
// 			expectedBody: user_model.PaginatedUsersResponse{
// 				Page:  2,
// 				Limit: 5,
// 				Total: 5,
// 				Users: []user_model.UserResponse{{ID: 3, Name: "User 3", Age: 25}},
// 			},
// 			expectServiceCall: true,
// 		},
// 		// CommonHandler.GetPaginationParams/GetFilteringParams обрабатывают ошибки парсинга query
// 		// Но сервис может вернуть ErrInvalidServiceInput для логических ошибок (например, limit > max_allowed)
// 		{
// 			name:               "Ошибка сервиса: некорректные параметры запроса",
// 			requestPath:        "/api/users?page=1&limit=10", // Предположим, сервис считает limit=10 некорректным в этом тесте
// 			mockServiceReturn:  nil,
// 			mockServiceTotal:   0,
// 			mockServiceError:   user_service.ErrInvalidServiceInput,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error:   "Invalid query parameters",
// 				Details: user_service.ErrInvalidServiceInput.Error(),
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: ошибка базы данных",
// 			requestPath:        "/api/users?page=1&limit=10",
// 			mockServiceReturn:  nil,
// 			mockServiceTotal:   0,
// 			mockServiceError:   user_service.ErrServiceDatabaseError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Database operation failed",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: неизвестная ошибка",
// 			requestPath:        "/api/users?page=1&limit=10",
// 			mockServiceReturn:  nil,
// 			mockServiceTotal:   0,
// 			mockServiceError:   errors.New("some unknown get all error"),
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Failed to retrieve users",
// 			},
// 			expectServiceCall: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockService.Calls = []mock.Call{}

// 			// GetPaginationParams и GetFilteringParams вызываются внутри хендлера.
// 			// Так как мы используем реальный CommonHandler, мы полагаемся на его логику парсинга query params.
// 			// Мок сервиса настраивается с ожидаемыми параметрами, которые CommonHandler должен распарсить.
// 			page, limit, _ := mockCommon.GetPaginationParams(&gin.Context{Request: httptest.NewRequest(http.MethodGet, tt.requestPath, nil)})
// 			filters, _ := mockCommon.GetFilteringParams(&gin.Context{Request: httptest.NewRequest(http.MethodGet, tt.requestPath, nil)})

// 			if tt.expectServiceCall {
// 				mockService.On("GetAllUsers", mock.Anything, page, limit, filters).Return(tt.mockServiceReturn, tt.mockServiceTotal, tt.mockServiceError).Once()
// 			}

// 			_, w := setupTestRouter(http.MethodGet, tt.requestPath, handler.GetAllUsers, nil, 0, mockCommon, testLogger) // Передаем real commonHandler и testLogger

// 			assert.Equal(t, tt.expectedStatusCode, w.Code, "Неверный статус код")

// 			var actualBody interface{}
// 			if w.Body.Len() > 0 {
// 				if tt.expectedStatusCode >= 400 {
// 					actualBody = &common_handler.ErrorResponse{}
// 				} else {
// 					actualBody = &user_model.PaginatedUsersResponse{}
// 				}
// 				err := json.Unmarshal(w.Body.Bytes(), actualBody)
// 				assert.NoError(t, err, "Ошибка десериализации тела ответа")
// 			} else {
// 				actualBody = nil
// 			}

// 			assert.Equal(t, tt.expectedBody, actualBody, "Неверное тело ответа")

// 			mockService.AssertExpectations(t)
// 		})
// 	}
// }

// // --- Тесты для UpdateUser ---

// func TestUpdateUser(t *testing.T) {
// 	gin.SetMode(gin.ReleaseMode)

// 	mockService := new(MockUserService)
// 	testLogger := logrus.New()
// 	testLogger.SetOutput(io.Discard)
// 	testLogger.SetLevel(logrus.PanicLevel)
// 	commonHandler := common_handler.NewCommonHandler(testLogger)
// 	handler := user_handler.NewUserHandler(mockService, commonHandler, testLogger)

// 	tests := []struct {
// 		name               string
// 		userIDParam        string // ID пользователя в пути запроса
// 		authUserID         uint   // ID аутентифицированного пользователя из контекста
// 		requestBody        user_model.UpdateUserRequest
// 		mockServiceReturn  *user_model.User // Теперь ожидаем указатель
// 		mockServiceError   error
// 		expectedStatusCode int
// 		expectedBody       interface{}
// 		expectServiceCall  bool
// 		expectAuthCheck    bool // Нужно ли проверять аутентификацию/авторизацию в этом тесте?
// 	}{
// 		{
// 			name:        "Успешное обновление своего профиля",
// 			userIDParam: "1",
// 			authUserID:  1, // Аутентифицированный пользователь обновляет свой профиль
// 			requestBody: user_model.UpdateUserRequest{
// 				Name:  strPtr("Updated Name"),
// 				Age:   intPtr(35),
// 				Email: strPtr("updated@example.com"),
// 			},
// 			mockServiceReturn: &user_model.User{ // Передаем указатель
// 				ID:    1,
// 				Name:  "Updated Name",
// 				Email: "updated@example.com",
// 				Age:   35,
// 			},
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusOK,
// 			expectedBody: user_model.UserResponse{
// 				ID:    1,
// 				Name:  "Updated Name",
// 				Email: "updated@example.com",
// 				Age:   35,
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:              "Некорректный формат ID в пути",
// 			userIDParam:       "abc",
// 			authUserID:        1,
// 			requestBody:       user_model.UpdateUserRequest{Name: strPtr("Name")},
// 			mockServiceReturn: nil, mockServiceError: nil,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Invalid user ID format",
// 			},
// 			expectServiceCall: false,
// 			expectAuthCheck:   false, // Проверка ID происходит до проверки аутентификации
// 		},
// 		{
// 			name:              "Ошибка аутентификации: userID отсутствует в контексте",
// 			userIDParam:       "1",
// 			authUserID:        0, // userID не установлен в контексте
// 			requestBody:       user_model.UpdateUserRequest{Name: strPtr("Name")},
// 			mockServiceReturn: nil, mockServiceError: nil,
// 			expectedStatusCode: http.StatusInternalServerError, // Ошибка конфигурации/middleware
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Authentication context error",
// 			},
// 			expectServiceCall: false,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:              "Запрещено: попытка обновить другого пользователя",
// 			userIDParam:       "2",
// 			authUserID:        1, // Пользователь 1 пытается обновить пользователя 2
// 			requestBody:       user_model.UpdateUserRequest{Name: strPtr("Name")},
// 			mockServiceReturn: nil, mockServiceError: nil,
// 			expectedStatusCode: http.StatusForbidden,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Forbidden: You can only update your own profile",
// 			},
// 			expectServiceCall: false,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:              "Некорректный формат запроса (JSON bind error)",
// 			userIDParam:       "1",
// 			authUserID:        1,
// 			requestBody:       user_model.UpdateUserRequest{}, // Специально передаем некорректные данные
// 			mockServiceReturn: nil, mockServiceError: nil,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Invalid input data",
// 			},
// 			expectServiceCall: false,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:        "Ошибка сервиса: некорректные входные данные",
// 			userIDParam: "1",
// 			authUserID:  1,
// 			requestBody: user_model.UpdateUserRequest{
// 				Email: strPtr("invalid-email"),
// 				Age:   intPtr(-10),
// 			},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrInvalidServiceInput,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error:   "Invalid input data",
// 				Details: user_service.ErrInvalidServiceInput.Error(),
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: пользователь не найден",
// 			userIDParam:        "1",
// 			authUserID:         1,
// 			requestBody:        user_model.UpdateUserRequest{Name: strPtr("Name")},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrUserNotFound,
// 			expectedStatusCode: http.StatusNotFound,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: user_service.ErrUserNotFound.Error(),
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: email уже используется",
// 			userIDParam:        "1",
// 			authUserID:         1,
// 			requestBody:        user_model.UpdateUserRequest{Email: strPtr("existing@example.com")},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrEmailAlreadyTaken,
// 			expectedStatusCode: http.StatusConflict,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: user_service.ErrEmailAlreadyTaken.Error(),
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:        "Ошибка сервиса: нет полей для обновления -> возвращает 200 с текущими данными",
// 			userIDParam: "1",
// 			authUserID:  1,
// 			requestBody: user_model.UpdateUserRequest{}, // Пустой запрос на обновление
// 			mockServiceReturn: &user_model.User{ // Сервис должен вернуть текущие данные (указатель)
// 				ID: 1, Name: "Current Name", Email: "current@example.com", Age: 30,
// 			},
// 			mockServiceError:   user_service.ErrNoUpdateFields,
// 			expectedStatusCode: http.StatusOK,
// 			expectedBody: user_model.UserResponse{
// 				ID: 1, Name: "Current Name", Email: "current@example.com", Age: 30,
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: ошибка базы данных",
// 			userIDParam:        "1",
// 			authUserID:         1,
// 			requestBody:        user_model.UpdateUserRequest{Name: strPtr("Name")},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   user_service.ErrServiceDatabaseError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Database operation failed",
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: неизвестная ошибка",
// 			userIDParam:        "1",
// 			authUserID:         1,
// 			requestBody:        user_model.UpdateUserRequest{Name: strPtr("Name")},
// 			mockServiceReturn:  nil,
// 			mockServiceError:   errors.New("some unexpected update error"),
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Failed to update user",
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockService.Calls = []mock.Call{}

// 			userIDUint, parseErr := parseUint(tt.userIDParam)

// 			shouldCallService := tt.expectServiceCall && parseErr == nil && (tt.authUserID != 0 && tt.authUserID == userIDUint) // Сервис вызывается только если ID распарсен, пользователь аутентифицирован и авторизован
// 			if tt.name == "Ошибка аутентификации: userID отсутствует в контексте" || tt.name == "Запрещено: попытка обновить другого пользователя" || tt.name == "Некорректный формат ID в пути" || tt.name == "Некорректный формат запроса (JSON bind error)" {
// 				shouldCallService = false
// 			}

// 			if shouldCallService {
// 				// Передаем указатель в Return
// 				mockService.On("UpdateUser", mock.Anything, userIDUint, tt.requestBody).Return(tt.mockServiceReturn, tt.mockServiceError).Once()
// 			}

// 			var reqBody interface{} = tt.requestBody
// 			if tt.name == "Некорректный формат запроса (JSON bind error)" {
// 				reqBody = "invalid json string"
// 			}

// 			path := "/api/users/" + tt.userIDParam

// 			_, w := setupTestRouter(http.MethodPut, path, handler.UpdateUser, reqBody, tt.authUserID, commonHandler, testLogger) // Передаем real commonHandler и testLogger

// 			assert.Equal(t, tt.expectedStatusCode, w.Code, "Неверный статус код")

// 			var actualBody interface{}
// 			if w.Body.Len() > 0 {
// 				// Определяем ожидаемый тип тела ответа на основе статуса
// 				switch tt.expectedStatusCode {
// 				case http.StatusOK: // Успех или ErrNoUpdateFields
// 					actualBody = &user_model.UserResponse{}
// 				case http.StatusCreated: // Не ожидается для PUT, но на всякий случай
// 					actualBody = &user_model.UserResponse{}
// 				case http.StatusNoContent: // Не ожидается для PUT
// 					actualBody = nil
// 				default: // Ошибки (400, 401, 403, 404, 409, 500)
// 					actualBody = &common_handler.ErrorResponse{}
// 				}

// 				if actualBody != nil {
// 					err := json.Unmarshal(w.Body.Bytes(), actualBody)
// 					assert.NoError(t, err, "Ошибка десериализации тела ответа")
// 				}
// 			} else {
// 				actualBody = nil
// 			}

// 			if tt.name == "Некорректный формат запроса (JSON bind error)" && tt.expectedStatusCode == http.StatusBadRequest {
// 				expectedErrResp, ok := tt.expectedBody.(common_handler.ErrorResponse)
// 				assert.True(t, ok, "Ожидалось тело типа ErrorResponse")
// 				actualErrResp, ok := actualBody.(*common_handler.ErrorResponse)
// 				assert.True(t, ok, "Фактическое тело имеет тип ErrorResponse")
// 				assert.Equal(t, expectedErrResp.Error, actualErrResp.Error, "Неверное сообщение об ошибке")
// 				assert.NotEmpty(t, actualErrResp.Details, "Поле Details не должно быть пустым при ошибке BindJSON")
// 			} else {
// 				assert.Equal(t, tt.expectedBody, actualBody, "Неверное тело ответа")
// 			}

// 			mockService.AssertExpectations(t)
// 		})
// 	}
// }

// // --- Тесты для DeleteUser ---

// func TestDeleteUser(t *testing.T) {
// 	gin.SetMode(gin.ReleaseMode)

// 	mockService := new(MockUserService)
// 	testLogger := logrus.New()
// 	testLogger.SetOutput(io.Discard)
// 	testLogger.SetLevel(logrus.PanicLevel)
// 	commonHandler := common_handler.NewCommonHandler(testLogger)
// 	handler := user_handler.NewUserHandler(mockService, commonHandler, testLogger)

// 	tests := []struct {
// 		name               string
// 		userIDParam        string // ID пользователя в пути запроса
// 		authUserID         uint   // ID аутентифицированного пользователя из контекста
// 		mockServiceError   error
// 		expectedStatusCode int
// 		expectedBody       interface{}
// 		expectServiceCall  bool
// 		expectAuthCheck    bool
// 	}{
// 		{
// 			name:               "Успешное удаление своего аккаунта",
// 			userIDParam:        "1",
// 			authUserID:         1, // Аутентифицированный пользователь удаляет свой аккаунт
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusNoContent,
// 			expectedBody:       nil, // Нет тела ответа для 204
// 			expectServiceCall:  true,
// 			expectAuthCheck:    true,
// 		},
// 		{
// 			name:               "Некорректный формат ID в пути",
// 			userIDParam:        "abc",
// 			authUserID:         1,
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Invalid user ID format",
// 			},
// 			expectServiceCall: false,
// 			expectAuthCheck:   false,
// 		},
// 		{
// 			name:               "Ошибка аутентификации: userID отсутствует в контексте",
// 			userIDParam:        "1",
// 			authUserID:         0, // userID не установлен в контексте
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusInternalServerError, // Ошибка конфигурации/middleware
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Authentication context error",
// 			},
// 			expectServiceCall: false,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Запрещено: попытка удалить другого пользователя",
// 			userIDParam:        "2",
// 			authUserID:         1, // Пользователь 1 пытается удалить аккаунт пользователя 2
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusForbidden,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Forbidden: You can only delete your own account",
// 			},
// 			expectServiceCall: false,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: пользователь не найден",
// 			userIDParam:        "1",
// 			authUserID:         1,
// 			mockServiceError:   user_service.ErrUserNotFound,
// 			expectedStatusCode: http.StatusNotFound,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: user_service.ErrUserNotFound.Error(),
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: некорректный ID (например, 0) -> маппится в 404",
// 			userIDParam:        "0",
// 			authUserID:         1,
// 			mockServiceError:   user_service.ErrInvalidServiceInput,
// 			expectedStatusCode: http.StatusNotFound,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: user_service.ErrUserNotFound.Error(),
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: ошибка базы данных",
// 			userIDParam:        "1",
// 			authUserID:         1,
// 			mockServiceError:   user_service.ErrServiceDatabaseError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Database operation failed",
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 		{
// 			name:               "Ошибка сервиса: неизвестная ошибка",
// 			userIDParam:        "1",
// 			authUserID:         1,
// 			mockServiceError:   errors.New("some unexpected delete error"),
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Failed to delete user",
// 			},
// 			expectServiceCall: true,
// 			expectAuthCheck:   true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockService.Calls = []mock.Call{}

// 			userIDUint, parseErr := parseUint(tt.userIDParam)

// 			shouldCallService := tt.expectServiceCall && parseErr == nil && (tt.authUserID != 0 && tt.authUserID == userIDUint) // Сервис вызывается только если ID распарсен, пользователь аутентифицирован и авторизован
// 			if tt.name == "Ошибка аутентификации: userID отсутствует в контексте" || tt.name == "Запрещено: попытка удалить другого пользователя" || tt.name == "Некорректный формат ID в пути" {
// 				shouldCallService = false
// 			}

// 			if shouldCallService {
// 				mockService.On("DeleteUser", mock.Anything, userIDUint).Return(tt.mockServiceError).Once()
// 			}

// 			path := "/api/users/" + tt.userIDParam

// 			_, w := setupTestRouter(http.MethodDelete, path, handler.DeleteUser, nil, tt.authUserID, commonHandler, testLogger) // Передаем real commonHandler и testLogger

// 			assert.Equal(t, tt.expectedStatusCode, w.Code, "Неверный статус код")

// 			var actualBody interface{}
// 			if w.Body.Len() > 0 {
// 				actualBody = &common_handler.ErrorResponse{}
// 				err := json.Unmarshal(w.Body.Bytes(), actualBody)
// 				assert.NoError(t, err, "Ошибка десериализации тела ответа")
// 			} else {
// 				actualBody = nil
// 			}

// 			assert.Equal(t, tt.expectedBody, actualBody, "Неверное тело ответа")

// 			mockService.AssertExpectations(t)
// 		})
// 	}
// }

// // --- Тесты для LoginUser ---

// func TestLoginUser(t *testing.T) {
// 	gin.SetMode(gin.ReleaseMode)

// 	mockService := new(MockUserService)
// 	testLogger := logrus.New()
// 	testLogger.SetOutput(io.Discard)
// 	testLogger.SetLevel(logrus.PanicLevel)
// 	commonHandler := common_handler.NewCommonHandler(testLogger)
// 	handler := user_handler.NewUserHandler(mockService, commonHandler, testLogger)

// 	tests := []struct {
// 		name               string
// 		requestBody        user_model.LoginRequest
// 		mockServiceReturn  string // Токен
// 		mockServiceError   error
// 		expectedStatusCode int
// 		expectedBody       interface{}
// 		expectServiceCall  bool
// 	}{
// 		{
// 			name: "Успешный вход пользователя",
// 			requestBody: user_model.LoginRequest{
// 				Email:    "user@example.com",
// 				Password: "correctpassword",
// 			},
// 			mockServiceReturn:  "fake-jwt-token",
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusOK,
// 			expectedBody: user_model.LoginResponse{
// 				Token: "fake-jwt-token",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Некорректный формат запроса (JSON bind error)",
// 			requestBody: user_model.LoginRequest{
// 				Email:    "user@example.com",
// 				Password: "correctpassword",
// 			},
// 			mockServiceReturn:  "",
// 			mockServiceError:   nil,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Invalid input data",
// 			},
// 			expectServiceCall: false,
// 		},
// 		{
// 			name: "Ошибка сервиса: некорректные входные данные (пустой email/пароль)",
// 			requestBody: user_model.LoginRequest{
// 				Email:    "",
// 				Password: "",
// 			},
// 			mockServiceReturn:  "",
// 			mockServiceError:   user_service.ErrInvalidServiceInput,
// 			expectedStatusCode: http.StatusBadRequest,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error:   "Invalid input data",
// 				Details: user_service.ErrInvalidServiceInput.Error(),
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: неверные учетные данные (пользователь не найден или неверный пароль)",
// 			requestBody: user_model.LoginRequest{
// 				Email:    "nonexistent@example.com",
// 				Password: "wrongpassword",
// 			},
// 			mockServiceReturn:  "",
// 			mockServiceError:   user_service.ErrInvalidCredentials,
// 			expectedStatusCode: http.StatusUnauthorized,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Invalid credentials",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: внутренняя ошибка (например, при генерации токена)",
// 			requestBody: user_model.LoginRequest{
// 				Email:    "user@example.com",
// 				Password: "correctpassword",
// 			},
// 			mockServiceReturn:  "",
// 			mockServiceError:   user_service.ErrInternalServiceError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Internal server error",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: ошибка базы данных",
// 			requestBody: user_model.LoginRequest{
// 				Email:    "user@example.com",
// 				Password: "correctpassword",
// 			},
// 			mockServiceReturn:  "",
// 			mockServiceError:   user_service.ErrServiceDatabaseError,
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Database operation failed",
// 			},
// 			expectServiceCall: true,
// 		},
// 		{
// 			name: "Ошибка сервиса: неизвестная ошибка",
// 			requestBody: user_model.LoginRequest{
// 				Email:    "user@example.com",
// 				Password: "correctpassword",
// 			},
// 			mockServiceReturn:  "",
// 			mockServiceError:   errors.New("some unexpected login error"),
// 			expectedStatusCode: http.StatusInternalServerError,
// 			expectedBody: common_handler.ErrorResponse{
// 				Error: "Login failed",
// 			},
// 			expectServiceCall: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockService.Calls = []mock.Call{}

// 			if tt.expectServiceCall {
// 				mockService.On("LoginUser", mock.Anything, tt.requestBody).Return(tt.mockServiceReturn, tt.mockServiceError).Once()
// 			}

// 			var reqBody interface{} = tt.requestBody
// 			if tt.name == "Некорректный формат запроса (JSON bind error)" {
// 				reqBody = "invalid json string"
// 			}

// 			_, w := setupTestRouter(http.MethodPost, "/auth/login", handler.LoginUser, reqBody, 0, commonHandler, testLogger) // Передаем real commonHandler и testLogger

// 			assert.Equal(t, tt.expectedStatusCode, w.Code, "Неверный статус код")

// 			var actualBody interface{}
// 			if w.Body.Len() > 0 {
// 				if tt.expectedStatusCode >= 400 {
// 					actualBody = &common_handler.ErrorResponse{}
// 				} else {
// 					actualBody = &user_model.LoginResponse{}
// 				}
// 				err := json.Unmarshal(w.Body.Bytes(), actualBody)
// 				assert.NoError(t, err, "Ошибка десериализации тела ответа")
// 			} else {
// 				actualBody = nil
// 			}

// 			if tt.name == "Некорректный формат запроса (JSON bind error)" && tt.expectedStatusCode == http.StatusBadRequest {
// 				expectedErrResp, ok := tt.expectedBody.(common_handler.ErrorResponse)
// 				assert.True(t, ok, "Ожидалось тело типа ErrorResponse")
// 				actualErrResp, ok := actualBody.(*common_handler.ErrorResponse)
// 				assert.True(t, ok, "Фактическое тело имеет тип ErrorResponse")
// 				assert.Equal(t, expectedErrResp.Error, actualErrResp.Error, "Неверное сообщение об ошибке")
// 				assert.NotEmpty(t, actualErrResp.Details, "Поле Details не должно быть пустым при ошибке BindJSON")
// 			} else {
// 				assert.Equal(t, tt.expectedBody, actualBody, "Неверное тело ответа")
// 			}

// 			mockService.AssertExpectations(t)
// 		})
// 	}
// }
