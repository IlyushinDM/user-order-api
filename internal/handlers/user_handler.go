package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IlyushinDM/user-order-api/internal/models"
	"github.com/IlyushinDM/user-order-api/internal/services"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdUser, err := h.userService.CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdUser)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}




import (
	"github.com/gin-gonic/gin"
	"project/internal/services"
)

type UserHandler interface {
	CreateUser(c *gin.Context)
	GetUsers(c *gin.Context)
	GetUserByID(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
}

type AuthHandler interface {
	Login(c *gin.Context)
}

type OrderHandler interface {
	CreateOrder(c *gin.Context)
	GetOrdersByUserID(c *gin.Context)
}

func NewUserHandler(service services.UserService) UserHandler {
	return &userHandler{userService: service}
}

type userHandler struct {
	userService services.UserService
}

func (h *userHandler) CreateUser(c *gin.Context) {}
func (h *userHandler) GetUsers(c *gin.Context)    {}
func (h *userHandler) GetUserByID(c *gin.Context) {}
func (h *userHandler) UpdateUser(c *gin.Context)  {}
func (h *userHandler) DeleteUser(c *gin.Context)  {}

func NewAuthHandler(service services.UserService) AuthHandler {
	return &authHandler{userService: service}
}

type authHandler struct {
	userService services.UserService
}

func (h *authHandler) Login(c *gin.Context) {}

func NewOrderHandler(service services.OrderService) OrderHandler {
	return &orderHandler{orderService: service}
}

type orderHandler struct {
	orderService services.OrderService
}

func (h *orderHandler) CreateOrder(c *gin.Context)    {}
func (h *orderHandler) GetOrdersByUserID(c *gin.Context) {}
