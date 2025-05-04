package middleware

import (
	"fmt"
	"net/http"

	"github.com/IlyushinDM/user-order-api/internal/services"
)

type AuthMiddleware struct {
	userService services.UserService
}

func NewAuthMiddleware(userService services.UserService) *AuthMiddleware {
	return &AuthMiddleware{userService: userService}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//  Пример: Проверка токена авторизации в заголовке
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		//  Валидация токена (здесь должна быть реальная логика валидации)
		//  Допустим, проверяем, что токен соответствует какому-то пользователю
		// user, err := m.userService.GetUserByToken(token)
		// if err != nil {
		//     http.Error(w, "Unauthorized", http.StatusUnauthorized)
		//     return
		// }

		//  Если авторизация успешна, передаем управление следующему обработчику
		fmt.Println("Authentication successful")
		next(w, r)
	}
}


import "github.com/gin-gonic/gin"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// JWT validation logic goes here
		c.Next() // Call next handler if authenticated
	}
}
