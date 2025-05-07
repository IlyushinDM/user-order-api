package utils

// import (
// 	"testing"

// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/stretchr/testify/assert"
// )

// func TestGenerateJWT(t *testing.T) {
// 	userID := uint(123)
// 	email := "test@example.com"
// 	secret := "test-secret"
// 	expirationSeconds := 3600

// 	tokenString, err := GenerateJWT(userID, email, secret, expirationSeconds)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, tokenString)

// 	claims := &Claims{}
// 	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(secret), nil
// 	})

// 	assert.NoError(t, err)
// 	assert.True(t, token.Valid)
// 	assert.Equal(t, userID, claims.UserID)
// 	assert.Equal(t, email, claims.Email)
// 	assert.Equal(t, "user-order-api", claims.Issuer)
// }

// func TestValidateJWT(t *testing.T) {
// 	userID := uint(123)
// 	email := "test@example.com"
// 	secret := "test-secret"
// 	expirationSeconds := 3600

// 	tokenString, err := GenerateJWT(userID, email, secret, expirationSeconds)
// 	assert.NoError(t, err)

// 	claims, err := ValidateJWT(tokenString, secret)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, claims)
// 	assert.Equal(t, userID, claims.UserID)
// 	assert.Equal(t, email, claims.Email)

// 	// Test with invalid token
// 	invalidToken := "invalid-token"
// 	claims, err = ValidateJWT(invalidToken, secret)
// 	assert.Error(t, err)
// 	assert.Nil(t, claims)

// 	// Test with invalid secret
// 	claims, err = ValidateJWT(tokenString, "wrong-secret")
// 	assert.Error(t, err)
// 	assert.Nil(t, claims)

// 	// Test with expired token
// 	expiredTokenString, _ := GenerateJWT(userID, email, secret, -1)
// 	claims, err = ValidateJWT(expiredTokenString, secret)
// 	assert.Error(t, err)
// 	assert.Nil(t, claims)
// }
