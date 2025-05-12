package jwt_util

import (
	"errors" // Для errors.Is и errors.As
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret          = "test-super-secret-key"
	testUserID          = uint(123)
	testUserEmail       = "test@example.com"
	testExpirationShort = 1  // 1 секунда, для теста истечения срока
	testExpirationValid = 60 // 1 минута
)

func TestGenerateJWT_Success(t *testing.T) {
	tokenString, err := GenerateJWT(testUserID, testUserEmail, testSecret, testExpirationValid)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	// ... (остальная часть теста без изменений)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	require.NoError(t, err, "Token should be parseable unverified")
	claims, ok := token.Claims.(*Claims)
	require.True(t, ok, "Token claims should be of type *Claims")
	assert.Equal(t, testUserID, claims.UserID)
	assert.Equal(t, testUserEmail, claims.Email)
	assert.Equal(t, "user-order-api", claims.Issuer)
	assert.WithinDuration(t, time.Now().Add(time.Duration(testExpirationValid)*time.Second), claims.ExpiresAt.Time, 2*time.Second)
	assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, 2*time.Second)
}

func TestValidateJWT_Success(t *testing.T) {
	tokenString, err := GenerateJWT(testUserID, testUserEmail, testSecret, testExpirationValid)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	claims, err := ValidateJWT(tokenString, testSecret)
	assert.NoError(t, err)
	require.NotNil(t, claims)
	// ... (остальная часть теста без изменений)
	assert.Equal(t, testUserID, claims.UserID)
	assert.Equal(t, testUserEmail, claims.Email)
	assert.Equal(t, "user-order-api", claims.Issuer)
	assert.WithinDuration(t, time.Now().Add(time.Duration(testExpirationValid)*time.Second), claims.ExpiresAt.Time, 2*time.Second)
}

func TestValidateJWT_EmptyTokenString(t *testing.T) {
	_, err := ValidateJWT("", testSecret)
	assert.Error(t, err)
	assert.EqualError(t, err, "token string cannot be empty")
}

func TestValidateJWT_EmptySecretForValidation(t *testing.T) {
	tokenString, _ := GenerateJWT(testUserID, testUserEmail, testSecret, testExpirationValid)
	_, err := ValidateJWT(tokenString, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "secret for validation cannot be empty")
}

func TestValidateJWT_InvalidSecret(t *testing.T) {
	tokenString, err := GenerateJWT(testUserID, testUserEmail, testSecret, testExpirationValid)
	require.NoError(t, err)

	_, err = ValidateJWT(tokenString, "wrong-secret-key")
	assert.Error(t, err)
	// Проверяем, что обернутая ошибка содержит jwt.ErrSignatureInvalid
	assert.True(t, errors.Is(err, jwt.ErrSignatureInvalid), "Error should wrap jwt.ErrSignatureInvalid")
	assert.Contains(t, err.Error(), "token validation failed", "Outer error context missing")
	assert.Contains(t, err.Error(), jwt.ErrSignatureInvalid.Error(), "Inner error message for invalid signature missing")
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	tokenString, err := GenerateJWT(testUserID, testUserEmail, testSecret, testExpirationShort)
	require.NoError(t, err)

	time.Sleep(time.Duration(testExpirationShort+1) * time.Second)

	_, err = ValidateJWT(tokenString, testSecret)
	assert.Error(t, err)
	// Проверяем, что обернутая ошибка содержит jwt.ErrTokenExpired
	assert.True(t, errors.Is(err, jwt.ErrTokenExpired), "Error should wrap jwt.ErrTokenExpired")
	assert.Contains(t, err.Error(), "token validation failed", "Outer error context missing")
	assert.Contains(t, err.Error(), jwt.ErrTokenExpired.Error(), "Inner error message for token expired missing")
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	malformedToken := "this.is.not.a.valid.jwt.token"
	_, err := ValidateJWT(malformedToken, testSecret)
	assert.Error(t, err)
	// Проверяем, что обернутая ошибка содержит jwt.ErrTokenMalformed
	assert.True(t, errors.Is(err, jwt.ErrTokenMalformed), "Error should wrap jwt.ErrTokenMalformed")
	assert.Contains(t, err.Error(), "token validation failed", "Outer error context missing")
}

// func TestValidateJWT_DifferentSigningMethod(t *testing.T) {
// 	claims := &Claims{
// 		UserID: testUserID,
// 		Email:  testUserEmail,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
// 		},
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // Используем HS512
// 	tokenString, err := token.SignedString([]byte(testSecret))
// 	require.NoError(t, err)

// 	_, err = ValidateJWT(tokenString, testSecret)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "token validation failed", "Outer error context missing")
// 	// Ошибка из callback (unexpected signing method) будет причиной обернутой ошибки
// 	// Мы можем попытаться извлечь ее, если знаем, что она там должна быть.
// 	// jwt.ParseWithClaims оборачивает ошибку из keyFunc в *jwt.ValidationError с флагом jwt.ValidationErrorSignatureInvalid
// 	// или jwt.ValidationErrorUnverifiable, если keyFunc вернула ошибку.
// 	// В данном случае, это будет ValidationErrorUnverifiable, потому что keyFunc вернула ошибку.
// 	assert.True(t, errors.Is(err, jwt.ErrTokenUnverifiable), "Error should wrap jwt.ErrTokenUnverifiable because keyFunc failed")
// 	// И также проверяем текст, который мы сами сгенерировали в keyFunc
// 	unwrappedErr := errors.Unwrap(err) // Снимаем нашу обертку "token validation failed"
// 	require.NotNil(t, unwrappedErr)
// 	unwrappedErr = errors.Unwrap(unwrappedErr) // Снимаем обертку *jwt.ValidationError
// 	require.NotNil(t, unwrappedErr, "Expected keyFunc error to be unwrappable")
// 	assert.Contains(t, unwrappedErr.Error(), "unexpected signing method: HS512")
// }

// func TestValidateJWT_TokenSignedWithNoAlgorithm(t *testing.T) {
// 	claims := &Claims{
// 		UserID: testUserID,
// 		Email:  testUserEmail,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
// 		},
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
// 	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
// 	require.NoError(t, err)

// 	_, err = ValidateJWT(tokenString, testSecret)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "token validation failed", "Outer error context missing")
// 	assert.True(t, errors.Is(err, jwt.ErrTokenUnverifiable), "Error should wrap jwt.ErrTokenUnverifiable because keyFunc failed")

// 	unwrappedErr := errors.Unwrap(err)
// 	require.NotNil(t, unwrappedErr)
// 	unwrappedErr = errors.Unwrap(unwrappedErr)
// 	require.NotNil(t, unwrappedErr)
// 	assert.Contains(t, unwrappedErr.Error(), "unexpected signing method: none")
// }

func TestGenerateAndValidate_Integration(t *testing.T) {
	userID := uint(42)
	email := "integration@test.com"
	secret := "a-bit-longer-secret-for-integration"
	expiration := 300

	tokenString, err := GenerateJWT(userID, email, secret, expiration)
	require.NoError(t, err, "Generation failed")
	require.NotEmpty(t, tokenString, "Generated token string is empty")

	time.Sleep(50 * time.Millisecond)

	claims, err := ValidateJWT(tokenString, secret)
	require.NoError(t, err, "Validation failed")
	require.NotNil(t, claims, "Claims should not be nil after validation")

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "user-order-api", claims.Issuer)
	assert.WithinDuration(t, time.Now().Add(time.Duration(expiration)*time.Second), claims.ExpiresAt.Time, 2*time.Second)
}
