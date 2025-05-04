package services

import (
	"github.com/IlyushinDM/user-order-api/internal/repository"
)

type OrderService interface {
	// Define methods for order business logic
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

type orderService struct {
	orderRepo repository.OrderRepository
}
