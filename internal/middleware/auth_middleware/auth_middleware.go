package auth_middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/utils/jwt_util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware создает middleware для аутентификации JWT токенов
func AuthMiddleware(log *logrus.Logger, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем запросы регистрации пользователей
		if c.Request.Method == http.MethodPost && c.Request.URL.Path == "/api/users" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Отсутствует заголовок Authorization")
			c.AbortWithStatusJSON(http.StatusUnauthorized, common_handler.ErrorResponse{
				Error: "Требуется заголовок Authorization",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Warn("Некорректный формат заголовка Authorization")
			c.AbortWithStatusJSON(http.StatusUnauthorized, common_handler.ErrorResponse{
				Error: "Формат заголовка должен быть Bearer {token}",
			})
			return
		}

		tokenString := parts[1]
		if jwtSecret == "" {
			log.Error("Не установлена переменная окружения JWT_SECRET")
			c.AbortWithStatusJSON(http.StatusInternalServerError, common_handler.ErrorResponse{
				Error: "Ошибка конфигурации сервера",
			})
			return
		}

		claims, err := jwt_util.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			log.WithError(err).Warn("Некорректный JWT токен")
			status := http.StatusUnauthorized
			message := "Неверный или просроченный токен"
			if errors.Is(err, jwt.ErrTokenExpired) {
				message = "Токен просрочен"
			}
			c.AbortWithStatusJSON(status, common_handler.ErrorResponse{Error: message})
			return
		}

		// Добавляем информацию о пользователе в контекст
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)

		log.WithFields(logrus.Fields{
			"userID":    claims.UserID,
			"userEmail": claims.Email,
		}).Debug("Пользователь успешно аутентифицирован")

		c.Next()
	}
}
