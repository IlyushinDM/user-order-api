package auth_middleware

import (
	"errors"
	"net/http"

	"github.com/IlyushinDM/user-order-api/internal/utils/jwt_util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware создает middleware для аутентификации JWT токенов
func AuthMiddleware(log *logrus.Logger, jwtSecret string) gin.HandlerFunc {
	return AuthMiddlewareWithValidator(log, jwtSecret, jwt_util.ValidateJWT)
}

func AuthMiddlewareWithValidator(
	log *logrus.Logger,
	jwtSecret string,
	validateJWT func(tokenString, secret string) (*jwt_util.Claims, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && c.Request.URL.Path == "/api/users" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.String(http.StatusUnauthorized, "Требуется заголовок Authorization")
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
			c.String(http.StatusUnauthorized, "Формат заголовка должен быть Bearer {token}")
			c.Abort()
			return
		}

		tokenString := authHeader[len(prefix):]
		if jwtSecret == "" {
			c.String(http.StatusInternalServerError, "Ошибка конфигурации сервера")
			c.Abort()
			return
		}

		claims, err := jwt_util.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			log.Warnf("Не удалось выполнить проверку JWT: %v", err)
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.String(http.StatusUnauthorized, "Токен просрочен")
			} else {
				c.String(http.StatusUnauthorized, "Неверный или просроченный токен")
			}
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}
