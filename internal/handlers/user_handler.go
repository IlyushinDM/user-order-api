package handlers

import (
	"net/http"
	"strconv"

	"errors"

	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService services.UserService
	log         *logrus.Logger
}

func NewUserHandler(userService services.UserService, log *logrus.Logger) *UserHandler {
	return &UserHandler{userService: userService, log: log}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Register a new user with name, email, age, and password.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User data"
// @Success 201 {object} models.UserResponse "User created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input data"
// @Failure 409 {object} ErrorResponse "User with this email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("CreateUser: Bad request format")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			h.log.WithError(err).Warn("CreateUser: Conflict - email exists")
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		} else {
			h.log.WithError(err).Error("CreateUser: Failed to create user")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Retrieve details of a specific user by their ID. Requires authentication.
// @Tags Users
// @Produce json
// @Param id path int true "User ID" Format(uint)
// @Success 200 {object} models.UserResponse "User details"
// @Failure 400 {object} ErrorResponse "Invalid user ID format"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("GetUserByID: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Warnf("GetUserByID: User not found, ID: %d", id)
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else {
			h.log.WithError(err).Errorf("GetUserByID: Failed to get user, ID: %d", id)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve user"})
		}
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	})
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Retrieve a paginated and filtered list of users. Requires authentication.
// @Tags Users
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Param min_age query int false "Minimum age filter" minimum(1)
// @Param max_age query int false "Maximum age filter" minimum(1)
// @Param name query string false "Name filter (case-insensitive, partial match)"
// @Success 200 {object} models.PaginatedUsersResponse "List of users"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page, limit := getPaginationParams(c)
	filters := getFilteringParams(c)

	users, total, err := h.userService.GetAllUsers(c.Request.Context(), page, limit, filters)
	if err != nil {
		h.log.WithError(err).Error("GetAllUsers: Failed to retrieve users")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve users"})
		return
	}

	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = models.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Age:   user.Age,
		}
	}

	response := models.PaginatedUsersResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Users: userResponses,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update details of an existing user by ID. Requires authentication. User can only update their own details unless they are an admin (admin logic not implemented here).
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" Format(uint)
// @Param user body models.UpdateUserRequest true "User data to update"
// @Success 200 {object} models.UserResponse "User updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input data or user ID format"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (trying to update another user - simplistic check)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 409 {object} ErrorResponse "Email already taken by another user"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("UpdateUser: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	// --- Authorization Check ---
	// Get user ID from JWT token (set by middleware)
	authUserID, exists := c.Get("userID")
	if !exists {
		h.log.Error("UpdateUser: userID not found in context (middleware error?)")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Authentication context error"})
		return
	}
	// Basic check: user can only update themselves
	if authUserID.(uint) != uint(id) {
		h.log.Warnf("UpdateUser: Forbidden attempt by user %d to update user %d", authUserID.(uint), id)
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden: You can only update your own profile"})
		return
	}
	// --- End Authorization Check ---

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warnf("UpdateUser: Bad request format for ID %d", id)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Warnf("UpdateUser: User not found, ID: %d", id)
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else if err.Error() == "email already taken by another user" {
			h.log.WithError(err).Warnf("UpdateUser: Conflict - email exists for ID %d", id)
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		} else {
			h.log.WithError(err).Errorf("UpdateUser: Failed to update user, ID: %d", id)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user"})
		}
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by their ID. Requires authentication. User can only delete their own account unless they are an admin (admin logic not implemented here).
// @Tags Users
// @Produce json
// @Param id path int true "User ID" Format(uint)
// @Success 204 "User deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID format"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (trying to delete another user)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("DeleteUser: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	// --- Authorization Check ---
	authUserID, exists := c.Get("userID")
	if !exists {
		h.log.Error("DeleteUser: userID not found in context")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Authentication context error"})
		return
	}
	if authUserID.(uint) != uint(id) {
		h.log.Warnf("DeleteUser: Forbidden attempt by user %d to delete user %d", authUserID.(uint), id)
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden: You can only delete your own account"})
		return
	}
	// --- End Authorization Check ---

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Warnf("DeleteUser: User not found, ID: %d", id)
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		} else {
			h.log.WithError(err).Errorf("DeleteUser: Failed to delete user, ID: %d", id)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete user"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// LoginUser godoc
// @Summary User login
// @Description Authenticate a user with email and password, returns a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse "Login successful, includes JWT token"
// @Failure 400 {object} ErrorResponse "Invalid input data"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("LoginUser: Bad request format")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	token, err := h.userService.LoginUser(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			h.log.Warnf("LoginUser: Invalid credentials for email %s", req.Email)
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		} else {
			h.log.WithError(err).Error("LoginUser: Failed to process login")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{Token: token})
}

// import (
// 	"log"
// 	"net/http"
// 	"strconv"

// 	"github.com/IlyushinDM/user-order-api/internal/models"
// 	"github.com/gin-gonic/gin"
// 	"golang.org/x/crypto/bcrypt"
// 	"gorm.io/gorm"
// )

// type UserHandler struct {
// 	db *gorm.DB
// }

// func NewUserHandler(db *gorm.DB) *UserHandler {
// 	return &UserHandler{db: db}
// }

// // --- Структуры для запросов/ответов ---

// type CreateUserRequest struct {
// 	Name     string `json:"name" binding:"required"`
// 	Email    string `json:"email" binding:"required,email"`
// 	Age      int    `json:"age" binding:"required,gt=0"`       // Добавлена валидация возраста
// 	Password string `json:"password" binding:"required,min=6"` // Добавлена валидация пароля
// }

// type UpdateUserRequest struct {
// 	Name  string `json:"name"`                            // Поля опциональны для обновления
// 	Email string `json:"email" binding:"omitempty,email"` // Валидация email, если передан
// 	Age   int    `json:"age" binding:"omitempty,gt=0"`    // Валидация возраста, если передан
// }

// type PaginatedUserResponse struct {
// 	Page  int                   `json:"page"`
// 	Limit int                   `json:"limit"`
// 	Total int64                 `json:"total"` // Используем int64 для совместимости с Count()
// 	Users []models.UserResponse `json:"users"`
// }

// // --- Обработчики ---

// // Create User
// func (h *UserHandler) Create(c *gin.Context) {
// 	var req CreateUserRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
// 		return
// 	}

// 	// Check if user exists
// 	var existingUser models.User
// 	// Используем First() вместо Where().First(), чтобы избежать ошибки, если пользователя нет
// 	result := h.db.Where("email = ?", req.Email).Limit(1).Find(&existingUser)
// 	// Проверяем, была ли найдена запись (а не ошибку запроса)
// 	if result.RowsAffected > 0 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
// 		return
// 	}
// 	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
// 		log.Printf("Database error checking user existence: %v\n", result.Error)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
// 		return
// 	}

// 	// Hash password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		log.Printf("Failed to hash password: %v\n", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
// 		return
// 	}

// 	user := models.User{
// 		Name:         req.Name,
// 		Email:        req.Email,
// 		Age:          req.Age,
// 		PasswordHash: string(hashedPassword),
// 		// Orders не указываем здесь, они будут связаны отдельно
// 	}

// 	if result := h.db.Create(&user); result.Error != nil {
// 		log.Printf("Failed to create user: %v\n", result.Error)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, models.UserResponse{
// 		ID:    user.ID,
// 		Name:  user.Name,
// 		Email: user.Email,
// 		Age:   user.Age,
// 	})
// }

// // Get User By ID
// func (h *UserHandler) GetByID(c *gin.Context) {
// 	idStr := c.Param("id")
// 	id, err := strconv.ParseUint(idStr, 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
// 		return
// 	}

// 	var user models.User
// 	// Используем Preload("Orders"), если нужно загрузить связанные заказы
// 	result := h.db.First(&user, uint(id))

// 	if result.Error != nil {
// 		if result.Error == gorm.ErrRecordNotFound {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 			return
// 		}
// 		log.Printf("Failed to fetch user by ID %d: %v\n", id, result.Error)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, models.UserResponse{
// 		ID:    user.ID,
// 		Name:  user.Name,
// 		Email: user.Email,
// 		Age:   user.Age,
// 	})
// }

// // Get All Users (with Pagination and Filtering)
// func (h *UserHandler) GetAll(c *gin.Context) {
// 	// Параметры пагинации
// 	pageStr := c.DefaultQuery("page", "1")
// 	limitStr := c.DefaultQuery("limit", "10")
// 	page, err := strconv.Atoi(pageStr)
// 	if err != nil || page < 1 {
// 		page = 1
// 	}
// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit < 1 {
// 		limit = 10
// 	}
// 	offset := (page - 1) * limit

// 	// Параметры фильтрации
// 	minAgeStr := c.Query("min_age")
// 	maxAgeStr := c.Query("max_age")

// 	// Создаем запрос к БД
// 	query := h.db.Model(&models.User{})

// 	// Применяем фильтры
// 	if minAgeStr != "" {
// 		minAge, err := strconv.Atoi(minAgeStr)
// 		if err == nil && minAge > 0 {
// 			query = query.Where("age >= ?", minAge)
// 		} else {
// 			log.Printf("Invalid min_age parameter: %s\n", minAgeStr)
// 			// Можно вернуть ошибку 400, если параметр невалиден
// 			// c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_age parameter"})
// 			// return
// 		}
// 	}
// 	if maxAgeStr != "" {
// 		maxAge, err := strconv.Atoi(maxAgeStr)
// 		if err == nil && maxAge > 0 {
// 			query = query.Where("age <= ?", maxAge)
// 		} else {
// 			log.Printf("Invalid max_age parameter: %s\n", maxAgeStr)
// 			// Можно вернуть ошибку 400
// 		}
// 	}

// 	// Получаем общее количество записей (до применения Limit/Offset)
// 	var total int64
// 	if err := query.Count(&total).Error; err != nil {
// 		log.Printf("Failed to count users: %v\n", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error counting users"})
// 		return
// 	}

// 	// Получаем список пользователей с пагинацией
// 	var users []models.User
// 	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
// 		log.Printf("Failed to fetch users: %v\n", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching users"})
// 		return
// 	}

// 	// Преобразуем в формат UserResponse
// 	userResponses := make([]models.UserResponse, len(users))
// 	for i, user := range users {
// 		userResponses[i] = models.UserResponse{
// 			ID:    user.ID,
// 			Name:  user.Name,
// 			Email: user.Email,
// 			Age:   user.Age,
// 		}
// 	}

// 	// Формируем ответ
// 	response := PaginatedUserResponse{
// 		Page:  page,
// 		Limit: limit,
// 		Total: total,
// 		Users: userResponses,
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // Update User
// func (h *UserHandler) Update(c *gin.Context) {
// 	idStr := c.Param("id")
// 	id, err := strconv.ParseUint(idStr, 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
// 		return
// 	}

// 	// Находим пользователя
// 	var user models.User
// 	if result := h.db.First(&user, uint(id)); result.Error != nil {
// 		if result.Error == gorm.ErrRecordNotFound {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 			return
// 		}
// 		log.Printf("Database error finding user %d for update: %v\n", id, result.Error)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error finding user"})
// 		return
// 	}

// 	// Биндим тело запроса
// 	var req UpdateUserRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data", "details": err.Error()})
// 		return
// 	}

// 	// Обновляем поля, если они переданы в запросе
// 	updateData := make(map[string]interface{})
// 	if req.Name != "" {
// 		updateData["name"] = req.Name
// 	}
// 	if req.Email != "" {
// 		// Дополнительно проверяем, не занят ли новый email другим пользователем
// 		var existingUser models.User
// 		result := h.db.Where("email = ? AND id != ?", req.Email, user.ID).Limit(1).Find(&existingUser)
// 		if result.RowsAffected > 0 {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already taken by another user"})
// 			return
// 		}
// 		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
// 			log.Printf("Database error checking email uniqueness during update: %v\n", result.Error)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking email"})
// 			return
// 		}
// 		updateData["email"] = req.Email
// 	}
// 	if req.Age > 0 { // Используем > 0, так как age - число
// 		updateData["age"] = req.Age
// 	}

// 	// Применяем обновления, если есть что обновлять
// 	if len(updateData) > 0 {
// 		if result := h.db.Model(&user).Updates(updateData); result.Error != nil {
// 			log.Printf("Failed to update user %d: %v\n", id, result.Error)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
// 			return
// 		}
// 	} else {
// 		// Если не передано ни одного поля для обновления
// 		log.Printf("Update request for user %d received with no fields to update.\n", id)
// 		// Можно вернуть 200 OK с текущими данными или 304 Not Modified
// 		// Возвращаем 200 OK с текущими данными для простоты
// 	}

// 	// Возвращаем обновленного пользователя (даже если ничего не изменилось)
// 	c.JSON(http.StatusOK, models.UserResponse{
// 		ID:    user.ID,    // ID остается тем же
// 		Name:  user.Name,  // Имя могло обновиться
// 		Email: user.Email, // Email мог обновиться
// 		Age:   user.Age,   // Возраст мог обновиться
// 	})
// }

// // Delete User
// func (h *UserHandler) Delete(c *gin.Context) {
// 	idStr := c.Param("id")
// 	id, err := strconv.ParseUint(idStr, 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
// 		return
// 	}

// 	// Пытаемся удалить пользователя
// 	// GORM вернет ошибку, если запись не найдена при использовании Delete со структурой или срезом,
// 	// но при удалении по ID он не вернет ошибку, если записи нет.
// 	// Поэтому проверяем RowsAffected.
// 	result := h.db.Delete(&models.User{}, uint(id))

// 	if result.Error != nil {
// 		log.Printf("Failed to delete user %d: %v\n", id, result.Error)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error deleting user"})
// 		return
// 	}

// 	// Проверяем, была ли запись реально удалена
// 	if result.RowsAffected == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}

// 	// Успешное удаление
// 	c.Status(http.StatusNoContent) // Возвращаем 204 No Content
// }
