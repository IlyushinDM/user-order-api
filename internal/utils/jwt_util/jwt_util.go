package jwt_util

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT создает новый токен JWT для пользователя
func GenerateJWT(userID uint, email string, secret string, expirationSeconds int) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("секрет не может быть пустой")
	}

	expirationTime := time.Now().Add(time.Duration(expirationSeconds) * time.Second)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "user-order-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT проверяет JWT токен.
func ValidateJWT(tokenString string, secret string) (*Claims, error) {
	if secret == "" {
		return nil, fmt.Errorf("secret for validation cannot be empty")
	}
	if tokenString == "" {
		return nil, fmt.Errorf("token string cannot be empty")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
