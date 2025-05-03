package handlers

import (
	"net/http"
	"strconv"

	"github.com/IlyushinDM/user-order-api/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Age      int    `json:"age" binding:"required,gte=0"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(input.Name, input.Email, input.Password, input.Age)
	if err != nil {
		// Check error type to return 400 for duplicate email, 500 for others
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"}) // Or appropriate error status
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"age":   user.Age,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		// Check if user not found error to return 404
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"}) // Or appropriate error status
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"age":   user.Age,
	})
}

// Implement other handlers: ListUsers, UpdateUser, DeleteUser, CreateOrder, ListOrders, Login
