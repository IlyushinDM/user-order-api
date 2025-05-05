package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/IlyushinDM/user-order-api/internal/handlers" // For ErrorResponse
	"github.com/IlyushinDM/user-order-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
func AuthMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("AuthMiddleware: Authorization header missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, handlers.ErrorResponse{Error: "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Warn("AuthMiddleware: Invalid Authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, handlers.ErrorResponse{Error: "Authorization header format must be Bearer {token}"})
			return
		}

		tokenString := parts[1]
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Error("AuthMiddleware: JWT_SECRET environment variable not set")
			c.AbortWithStatusJSON(http.StatusInternalServerError, handlers.ErrorResponse{Error: "Server configuration error"})
			return
		}

		claims, err := utils.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			log.WithError(err).Warn("AuthMiddleware: Invalid JWT token")
			status := http.StatusUnauthorized
			message := "Invalid or expired token"
			if errors.Is(err, jwt.ErrTokenExpired) {
				message = "Token has expired"
			}
			c.AbortWithStatusJSON(status, handlers.ErrorResponse{Error: message})
			return
		}

		// Set user information in the context for handlers to use
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email) // Add email if needed

		log.WithFields(logrus.Fields{
			"userID":    claims.UserID,
			"userEmail": claims.Email,
		}).Debug("AuthMiddleware: User authenticated")

		c.Next()
	}
}
