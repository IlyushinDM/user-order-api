package jwt_util

import (
	// Для errors.Is
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the structure for JWT claims.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for a user.
func GenerateJWT(userID uint, email string, secret string, expirationSeconds int) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
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

// ValidateJWT validates the given JWT token string.
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
			// Эта ошибка будет обернута ParseWithClaims в jwt.ValidationError
			// с флагом jwt.ValidationErrorSignatureInvalid или подобным
			// (на самом деле, это будет jwt.ValidationErrorMalformed, т.к. метод подписи не тот, что ожидался)
			// Более точно, это будет ошибка, которую мы возвращаем, и она будет частью Inner
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		// jwt.ParseWithClaims возвращает ошибку типа *jwt.ValidationError
		// Эта ошибка содержит битовую маску причин (например, токен истек, подпись неверна, и т.д.)
		// Мы можем проверять конкретные причины с помощью errors.Is или по маске.

		// Если мы хотим, чтобы вызывающий код мог использовать errors.Is(err, jwt.ErrTokenExpired),
		// то лучше не оборачивать эту ошибку дополнительно, либо оборачивать с %w.
		// В данном случае, возвращая err "как есть", мы сохраняем эту возможность.
		// Например, если токен истек, err будет *jwt.ValidationError, и errors.Is(err, jwt.ErrTokenExpired) вернет true.
		return nil, fmt.Errorf("token validation failed: %w", err) // Оборачиваем, чтобы добавить контекст, но сохраняем исходную ошибку
	}

	// Проверка token.Valid здесь почти избыточна, так как если бы он был невалиден,
	// jwt.ParseWithClaims уже вернул бы ошибку. Но на всякий случай оставим.
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
