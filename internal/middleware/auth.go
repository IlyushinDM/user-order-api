package middleware

import "github.com/gin-gonic/gin"

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Реализация JWT аутентификации
	}
}
