package user_handler

import (
	"net/http"
	"strconv"

	"errors"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/IlyushinDM/user-order-api/internal/services/user_service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService user_service.UserService
	log         *logrus.Logger
}

func NewUserHandler(userService user_service.UserService, log *logrus.Logger) *UserHandler {
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
// @Failure 400 {object} common_handler.ErrorResponse "Invalid input data"
// @Failure 409 {object} common_handler.ErrorResponse "User with this email already exists"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req user_model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("CreateUser: Bad request format")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "User with this email already exists" {
			h.log.WithError(err).Warn("CreateUser: Conflict - email exists")
			c.JSON(http.StatusConflict, common_handler.ErrorResponse{Error: err.Error()})
		} else {
			h.log.WithError(err).Error("CreateUser: Failed to create user")
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, user_model.UserResponse{
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
// @Failure 400 {object} common_handler.ErrorResponse "Invalid user ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} common_handler.ErrorResponse "User not found"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("GetUserByID: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Warnf("GetUserByID: User not found, ID: %d", id)
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "User not found"})
		} else {
			h.log.WithError(err).Errorf("GetUserByID: Failed to get user, ID: %d", id)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to retrieve user"})
		}
		return
	}

	c.JSON(http.StatusOK, user_model.UserResponse{
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
// @Failure 400 {object} common_handler.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page, limit := common_handler.GetPaginationParams(c)
	filters := common_handler.GetFilteringParams(c)

	users, total, err := h.userService.GetAllUsers(c.Request.Context(), page, limit, filters)
	if err != nil {
		h.log.WithError(err).Error("GetAllUsers: Failed to retrieve users")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to retrieve users"})
		return
	}

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
// @Summary Update a user
// @Description Update details of an existing user by ID. Requires authentication. User can only update their own details unless they are an admin (admin logic not implemented here).
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" Format(uint)
// @Param user body models.UpdateUserRequest true "User data to update"
// @Success 200 {object} models.UserResponse "User updated successfully"
// @Failure 400 {object} common_handler.ErrorResponse "Invalid input data or user ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} common_handler.ErrorResponse "Forbidden (trying to update another user - simplistic check)"
// @Failure 404 {object} common_handler.ErrorResponse "User not found"
// @Failure 409 {object} common_handler.ErrorResponse "Email already taken by another user"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("UpdateUser: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	// --- Authorization Check ---
	// Get user ID from JWT token (set by middleware)
	authUserID, exists := c.Get("userID")
	if !exists {
		h.log.Error("UpdateUser: userID not found in context (middleware error?)")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	// Basic check: user can only update themselves
	if authUserID.(uint) != uint(id) {
		h.log.Warnf("UpdateUser: Forbidden attempt by user %d to update user %d", authUserID.(uint), id)
		c.JSON(http.StatusForbidden, common_handler.ErrorResponse{Error: "Forbidden: You can only update your own profile"})
		return
	}
	// --- End Authorization Check ---

	var req user_model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warnf("UpdateUser: Bad request format for ID %d", id)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Warnf("UpdateUser: User not found, ID: %d", id)
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "User not found"})
		} else if err.Error() == "email already taken by another user" {
			h.log.WithError(err).Warnf("UpdateUser: Conflict - email exists for ID %d", id)
			c.JSON(http.StatusConflict, common_handler.ErrorResponse{Error: err.Error()})
		} else {
			h.log.WithError(err).Errorf("UpdateUser: Failed to update user, ID: %d", id)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to update user"})
		}
		return
	}

	c.JSON(http.StatusOK, user_model.UserResponse{
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
// @Failure 400 {object} common_handler.ErrorResponse "Invalid user ID format"
// @Failure 401 {object} common_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} common_handler.ErrorResponse "Forbidden (trying to delete another user)"
// @Failure 404 {object} common_handler.ErrorResponse "User not found"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.log.WithError(err).Warnf("DeleteUser: Invalid ID format '%s'", idStr)
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	// --- Authorization Check ---
	authUserID, exists := c.Get("userID")
	if !exists {
		h.log.Error("DeleteUser: userID not found in context")
		c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Authentication context error"})
		return
	}
	if authUserID.(uint) != uint(id) {
		h.log.Warnf("DeleteUser: Forbidden attempt by user %d to delete user %d", authUserID.(uint), id)
		c.JSON(http.StatusForbidden, common_handler.ErrorResponse{Error: "Forbidden: You can only delete your own account"})
		return
	}
	// --- End Authorization Check ---

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Warnf("DeleteUser: User not found, ID: %d", id)
			c.JSON(http.StatusNotFound, common_handler.ErrorResponse{Error: "User not found"})
		} else {
			h.log.WithError(err).Errorf("DeleteUser: Failed to delete user, ID: %d", id)
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Failed to delete user"})
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
// @Failure 400 {object} common_handler.ErrorResponse "Invalid input data"
// @Failure 401 {object} common_handler.ErrorResponse "Invalid credentials"
// @Failure 500 {object} common_handler.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	var req user_model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("LoginUser: Bad request format")
		c.JSON(http.StatusBadRequest, common_handler.ErrorResponse{Error: "Invalid input data", Details: err.Error()})
		return
	}

	token, err := h.userService.LoginUser(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			h.log.Warnf("LoginUser: Invalid credentials for email %s", req.Email)
			c.JSON(http.StatusUnauthorized, common_handler.ErrorResponse{Error: "Invalid credentials"})
		} else {
			h.log.WithError(err).Error("LoginUser: Failed to process login")
			c.JSON(http.StatusInternalServerError, common_handler.ErrorResponse{Error: "Login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, user_model.LoginResponse{Token: token})
}
