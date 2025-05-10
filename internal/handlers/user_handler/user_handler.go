package user_handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/services/user_service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	userService   user_service.UserService
	commonHandler *common_handler.CommonHandler
	log           *logrus.Logger
}

// NewUserHandler инициализирует UserHandler с проверкой зависимостей
func NewUserHandler(userService user_service.UserService, commonHandler *common_handler.CommonHandler, log *logrus.Logger) *UserHandler {
	if userService == nil {
		logrus.Fatal("UserService не может быть nil")
	}
	if commonHandler == nil {
		logrus.Fatal("CommonHandler не может быть nil")
	}
	if log == nil {
		defaultLog := logrus.New()
		defaultLog.SetLevel(logrus.InfoLevel)
		defaultLog.Warn("Logger не указан, используется logger по умолчанию")
		log = defaultLog
	}
	return &UserHandler{userService: userService, commonHandler: commonHandler, log: log}
}

// CreateUser godoc
// @Summary Создание нового пользователя
// @Description Регистрация нового пользователя с именем, email, возрастом и паролем.
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param user body user_model.CreateUserRequest true "Данные пользователя"
// @Success 201 {object} user_model.UserResponse "Пользователь успешно создан"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректные входные данные"
// @Failure 409 {object} common_handler.ErrorResponse "Пользователь с таким email уже существует"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	logger := h.log.WithContext(c.Request.Context()).WithField("method", "UserHandler.CreateUser")
	var req user_model.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Warn("Неправильный формат запроса")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые входные данные", Details: err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req)
	// Обработка ошибок сервисного слоя
	if err != nil {
		logger.WithError(err).Error("Сервис вернул ошибку при создании пользователя")
		switch {
		case errors.Is(err, user_service.ErrInvalidServiceInput):
			// Ошибка валидации на уровне сервиса
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые входные данные", Details: err.Error()})
		case errors.Is(err, user_service.ErrUserAlreadyExists):
			// Пользователь с таким email уже существует
			c.JSON(http.StatusConflict, common_handler.ErrorResponse{Error: "Пользователь с таким email уже существует"})
		case errors.Is(err, user_service.ErrInternalServiceError):
			// Внутренняя ошибка сервиса
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Внутренняя ошибка сервера"})
		case errors.Is(err, user_service.ErrServiceDatabaseError):
			// Ошибка, связанная с БД
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Сбой операции с базой данных"})
		default:
			// Неизвестная ошибка сервиса
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Не удалось создать пользователя"})
		}
		return
	}

	logger.WithField("user_id", user.ID).Info("Пользователь успешно создан")
	c.JSON(http.StatusCreated, user_model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	})
}

// GetUserByID godoc
// @Summary Получение пользователя по ID
// @Description Получение информации о конкретном пользователе по его ID. Требуется аутентификация.
// @Tags Пользователи
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Success 200 {object} user_model.UserResponse "Информация о пользователе"
// @Failure 400 {object} common_handler.ErrorResponse "Неверный формат ID пользователя"
// @Failure 401 {object} common_handler.ErrorResponse "Неавторизован"
// @Failure 404 {object} common_handler.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	logger := h.log.WithContext(c.Request.Context()).WithField("method", "UserHandler.GetUserByID")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.WithError(err).Warnf("Недопустимый формат идентификатора '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Неверный формат идентификатора пользователя"})
		return
	}
	logger = logger.WithField("user_id", uint(id))

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	// Обработка ошибок сервисного слоя
	if err != nil {
		logger.WithError(err).Error("Служба вернула ошибку при получении пользователя по идентификатору")
		switch {
		case errors.Is(err, user_service.ErrUserNotFound):
			// Пользователь не найден
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Пользователь не найден"})
		case errors.Is(err, user_service.ErrInvalidServiceInput):
			// Ошибка валидации ID на уровне сервиса (например, ID=0) - маппим к 404, т.к. 0 не существует.
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{
				Error: "Пользователь не найден: " + user_service.ErrUserNotFound.Error(),
			})
		case errors.Is(err, user_service.ErrServiceDatabaseError):
			// Ошибка, связанная с БД
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Сбой операции с базой данных"})
		default:
			// Неизвестная ошибка сервиса
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Неизвестная ошибка сервиса"})
		}
		return
	}

	logger.Info("Пользователь успешно восстановлен по ID")
	c.JSON(http.StatusOK, user_model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	})
}

// GetAllUsers godoc
// @Summary Получение всех пользователей
// @Description Получение списка пользователей с пагинацией и фильтрацией. Требуется аутентификация.
// @Tags Пользователи
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(10) minimum(1) maximum(100)
// @Param min_age query int false "Минимальный возраст для фильтрации" minimum(1)
// @Param max_age query int false "Максимальный возраст для фильтрации" minimum(1)
// @Param name query string false "Фильтр по имени (без учета регистра, частичное совпадение)"
// @Success 200 {object} user_model.PaginatedUsersResponse "Список пользователей"
// @Failure 400 {object} common_handler.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} common_handler.ErrorResponse "Неавторизован"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	logger := h.log.WithContext(c.Request.Context()).WithField("method", "UserHandler.GetAllUsers")

	page, limit, err := h.commonHandler.GetPaginationParams(c)
	if err != nil {
		logger.WithError(err).Warn("Недопустимые параметры разбивки на страницы")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые параметры разбивки на страницы", Details: err.Error()})
		return
	}

	filters, err := h.commonHandler.GetFilteringParams(c)
	if err != nil {
		logger.WithError(err).Warn("Недопустимые параметры фильтрации")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые параметры фильтрации", Details: err.Error()})
		return
	}

	logger = logger.WithFields(logrus.Fields{"page": page, "limit": limit, "filters": filters})

	users, total, err := h.userService.GetAllUsers(c.Request.Context(), page, limit, filters)
	// Обработка ошибок сервисного слоя
	if err != nil {
		logger.WithError(err).Error("Ошибка во время получения всех пользователей")
		switch {
		case errors.Is(err, user_service.ErrInvalidServiceInput):
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые параметры запроса", Details: err.Error()})
		case errors.Is(err, user_service.ErrServiceDatabaseError):
			// Ошибка, связанная с БД
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Сбой операции с базой данных"})
		default:
			// Неизвестная ошибка сервиса
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Неизвестная ошибка сервиса"})
		}
		return
	}

	logger.WithField("count", len(users)).Info("Пользователи были успешно восстановлены")
	userResponses := make([]user_model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user_model.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Age:   user.Age,
		}
	}

	response := user_model.PaginatedUsersResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Users: userResponses,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser godoc
// @Summary Обновление пользователя
// @Description Обновление информации о существующем пользователе по ID. Требуется аутентификация. Пользователь может обновлять только свои данные, если он не является администратором (логика администратора здесь не реализована).
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Param user body user_model.UpdateUserRequest true "Данные пользователя для обновления"
// @Success 200 {object} user_model.UserResponse "Пользователь успешно обновлен"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректные входные данные или неверный формат ID пользователя"
// @Failure 401 {object} common_handler.ErrorResponse "Неавторизован"
// @Failure 403 {object} common_handler.ErrorResponse "Запрещено (попытка обновить другого пользователя - упрощенная проверка)"
// @Failure 404 {object} common_handler.ErrorResponse "Пользователь не найден"
// @Failure 409 {object} common_handler.ErrorResponse "Email уже используется другим пользователем"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	logger := h.log.WithContext(c.Request.Context()).WithField("method", "UserHandler.UpdateUser")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.WithError(err).Warnf("Недопустимый формат идентификатора '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Неверный формат идентификатора пользователя"})
		return
	}
	logger = logger.WithField("user_id", uint(id))

	// Проверка авторизованного пользователя
	authUserID, exists := c.Get("userID")
	if !exists {
		logger.Error("userID не найден в context (Возможна ошибка в middleware)")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка context аутентификации"})
		return
	}
	if authUserID.(uint) != uint(id) {
		logger.Warnf("Forbidden attempt by user %d to update user %d", authUserID.(uint), id)
		c.JSON(http.StatusForbidden, common_handler.ErrorResponse{Error: "Forbidden: You can only update your own profile"})
		return
	}

	var req user_model.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Warn("Bad request format")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые входные данные", Details: err.Error()})
		return
	}

	// Call service to update user
	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), req)
	// Обработка ошибок сервисного слоя
	if err != nil {
		logger.WithError(err).Error("Service returned error during user update")
		switch {
		case errors.Is(err, user_service.ErrInvalidServiceInput):
			// Ошибка валидации на уровне сервиса
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые входные данные", Details: err.Error()})
		case errors.Is(err, user_service.ErrUserNotFound):
			// Пользователь не найден (либо по ID, либо конкурентное удаление)
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: err.Error()}) // Возвращаем сообщение "User not found"
		case errors.Is(err, user_service.ErrEmailAlreadyTaken):
			// Email уже занят другим пользователем
			c.JSON(http.StatusConflict, common_handler.ErrorResponse{Error: err.Error()})
		case errors.Is(err, user_service.ErrNoUpdateFields):
			// Нет полей для обновления (это может быть 200 OK)
			logger.Info("Update called with no fields to update, returning existing user data")
			// Сервис в этом случае возвращает текущие данные пользователя
			c.JSON(http.StatusOK, user_model.UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Age:   user.Age,
			})
			return // Важно выйти после успешного ответа
		case errors.Is(err, user_service.ErrServiceDatabaseError):
			// Ошибка, связанная с БД
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Сбой операции с базой данных"})
		default:
			// Неизвестная ошибка сервиса
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to update user"})
		}
		return
	}

	logger.Info("User updated successfully")
	// Respond with success
	c.JSON(http.StatusOK, user_model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	})
}

// DeleteUser godoc
// @Summary Удаление пользователя
// @Description Удаление пользователя по его ID. Требуется аутентификация. Пользователь может удалить только свою учетную запись, если он не является администратором (логика администратора здесь не реализована).
// @Tags Пользователи
// @Produce json
// @Param id path int true "ID пользователя" Format(uint)
// @Success 204 "Пользователь успешно удален"
// @Failure 400 {object} common_handler.ErrorResponse "Неверный формат ID пользователя"
// @Failure 401 {object} common_handler.ErrorResponse "Неавторизован"
// @Failure 403 {object} common_handler.ErrorResponse "Запрещено (попытка удалить другого пользователя)"
// @Failure 404 {object} common_handler.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	logger := h.log.WithContext(c.Request.Context()).WithField("method", "UserHandler.DeleteUser")
	idStr := c.Param("id")

	// Parse user ID from path parameter
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.WithError(err).Warnf("Недопустимый формат идентификатора '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Неверный формат идентификатора пользователя"})
		return
	}
	logger = logger.WithField("user_id", uint(id)) // Добавляем ID в лог

	// Проверка авторизованного пользователя (остается в обработчике)
	authUserID, exists := c.Get("userID") // userID должен быть установлен middleware аутентификации
	if !exists {
		logger.Error("userID not found in context")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	if authUserID.(uint) != uint(id) {
		logger.Warnf("Forbidden attempt by user %d to delete user %d", authUserID.(uint), id)
		c.JSON(http.StatusForbidden, common_handler.ErrorResponse{Error: "Forbidden: You can only delete your own account"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	// Обработка ошибок сервисного слоя
	if err != nil {
		logger.WithError(err).Error("Service returned error during user deletion")
		switch {
		case errors.Is(err, user_service.ErrInvalidServiceInput):
			// Ошибка валидации ID на уровне сервиса (например, ID=0) - маппим к 404, т.к. 0 не существует.
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: user_service.ErrUserNotFound.Error()})
		case errors.Is(err, user_service.ErrUserNotFound):
			// Пользователь не найден
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "Пользователь не найден"})
		case errors.Is(err, user_service.ErrServiceDatabaseError):
			// Ошибка, связанная с БД
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Сбой операции с базой данных"})
		default:
			// Неизвестная ошибка сервиса
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Не удалось удалить пользователяr"})
		}
		return
	}

	logger.Info("Пользователь успешно удален")
	c.Status(http.StatusNoContent)
}

// LoginUser godoc
// @Summary Вход пользователя
// @Description Аутентификация пользователя с использованием email и пароля, возвращает JWT токен.
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param credentials body user_model.LoginRequest true "Учетные данные для входа"
// @Success 200 {object} user_model.LoginResponse "Вход выполнен успешно, включает JWT токен"
// @Failure 400 {object} common_handler.ErrorResponse "Некорректные входные данные"
// @Failure 401 {object} common_handler.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} common_handler.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	logger := h.log.WithContext(c.Request.Context()).WithField("method", "UserHandler.LoginUser")
	var req user_model.LoginRequest

	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Warn("Неправильный формат запроса")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые входные данные", Details: err.Error()})
		return
	}
	logger = logger.WithField("email", req.Email)

	token, err := h.userService.LoginUser(c.Request.Context(), req)
	// Обработка ошибок сервисного слоя
	if err != nil {
		logger.WithError(err).Error("Служба вернула ошибку при входе в систему")
		switch {
		case errors.Is(err, user_service.ErrInvalidServiceInput):
			// Ошибка валидации на уровне сервиса (пустые email/пароль)
			c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Недопустимые входные данные", Details: err.Error()})
		case errors.Is(err, user_service.ErrInvalidCredentials):
			// Неверные учетные данные (пользователь не найден или неверный пароль)
			// Возвращаем generic 401 для безопасности.
			c.JSON(http.StatusUnauthorized, common_handler.ErrorResponse{Error: "Неверные учетные данные"})
		case errors.Is(err, user_service.ErrInternalServiceError):
			// Внутренняя ошибка сервиса (например, ошибка генерации JWT)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Внутренняя ошибка сервера"})
		case errors.Is(err, user_service.ErrServiceDatabaseError):
			// Ошибка, связанная с БД
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Сбой операции с базой данных"})
		default:
			// Неизвестная ошибка сервиса
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Ошибка входа"})
		}
		return
	}

	logger.Info("Пользователь успешно вошел в систему, токен сгенерирован")
	c.JSON(http.StatusOK, user_model.LoginResponse{Token: token})
}
