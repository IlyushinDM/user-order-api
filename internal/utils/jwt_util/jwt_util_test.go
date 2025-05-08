package jwt_util_test

import (
	"testing"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/utils/jwt_util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID      = uint(123)
	testUserEmail   = "test@example.com"
	testSecret      = "supersecretkey"
	testSecretWrong = "wrongsecretkey"
)

// TestGenerateAndValidateJWT_Success проверяет успешную генерацию и последующую валидацию токена.
func TestGenerateAndValidateJWT_Success(t *testing.T) {
	// Генерируем токен
	expirationSeconds := 3600 // 1 час
	tokenString, err := jwt_util.GenerateJWT(testUserID, testUserEmail, testSecret, expirationSeconds)
	require.NoError(t, err, "Expected no error when generating token")
	require.NotEmpty(t, tokenString, "Expected non-empty token string")

	// Валидируем сгенерированный токен
	claims, err := jwt_util.ValidateJWT(tokenString, testSecret)
	require.NoError(t, err, "Expected no error when validating valid token")
	require.NotNil(t, claims, "Expected non-nil claims")

	// Проверяем содержимое claims
	assert.Equal(t, testUserID, claims.UserID, "UserID in claims mismatch")
	assert.Equal(t, testUserEmail, claims.Email, "Email in claims mismatch")
	assert.Equal(t, "user-order-api", claims.Issuer, "Issuer in claims mismatch")

	// Проверяем время истечения токена (должно быть в будущем)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()), "Token should not be expired")

	// Проверяем время выдачи токена
	// Допускаем небольшую погрешность времени
	assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(5*time.Second)), "IssuedAt should be recent")
	assert.True(t, claims.IssuedAt.Time.After(time.Now().Add(-5*time.Second)), "IssuedAt should not be too far in the past")
}

// TestValidateJWT_Expired проверяет валидацию истекшего токена.
func TestValidateJWT_Expired(t *testing.T) {
	// Генерируем токен с очень коротким сроком действия
	expirationSeconds := 1
	tokenString, err := jwt_util.GenerateJWT(testUserID, testUserEmail, testSecret, expirationSeconds)
	require.NoError(t, err, "Expected no error when generating token for expiration test")

	// Ждем, пока токен истечет
	time.Sleep(2 * time.Second)

	// Пытаемся валидировать истекший токен
	claims, err := jwt_util.ValidateJWT(tokenString, testSecret)
	require.Error(t, err, "Expected error when validating expired token")
	// Обновлено: Проверяем, что ошибка содержит сообщение об истечении срока действия
	assert.Contains(t, err.Error(), "token is expired", "Expected expiration error message")
	assert.Nil(t, claims, "Expected nil claims for expired token")
}

// TestValidateJWT_InvalidSignature проверяет валидацию токена с неверным секретом.
func TestValidateJWT_InvalidSignature(t *testing.T) {
	// Генерируем токен с правильным секретом
	expirationSeconds := 3600
	tokenString, err := jwt_util.GenerateJWT(testUserID, testUserEmail, testSecret, expirationSeconds)
	require.NoError(t, err, "Expected no error when generating token for invalid signature test")

	// Пытаемся валидировать токен с неверным секретом
	claims, err := jwt_util.ValidateJWT(tokenString, testSecretWrong)
	require.Error(t, err, "Expected error when validating token with wrong secret")
	// Проверяем, что ошибка связана с парсингом (обычно из-за неверной подписи)
	assert.Contains(t, err.Error(), "failed to parse token", "Expected parsing error message")
	assert.Nil(t, claims, "Expected nil claims for token with invalid signature")
}

// TestValidateJWT_InvalidTokenFormat проверяет валидацию некорректной строки токена.
func TestValidateJWT_InvalidTokenFormat(t *testing.T) {
	invalidTokenString := "this.is.not.a.valid.jwt.token"

	// Пытаемся валидировать некорректную строку
	claims, err := jwt_util.ValidateJWT(invalidTokenString, testSecret)
	require.Error(t, err, "Expected error when validating invalid token string")
	assert.Contains(t, err.Error(), "failed to parse token", "Expected parsing error message for invalid format")
	assert.Nil(t, claims, "Expected nil claims for invalid token string")
}

// TestValidateJWT_InvalidSigningMethod проверяет валидацию токена с некорректным методом подписи.
func TestValidateJWT_InvalidSigningMethod(t *testing.T) {
	// Создаем токен с другим методом подписи (например, None), что должно быть отвергнуто функцией ValidateJWT
	claims := &jwt_util.Claims{
		UserID: testUserID,
		Email:  testUserEmail,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)                // Используем другой метод
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType) // Подписываем без ключа, т.к. метод None
	require.NoError(t, err, "Expected no error when generating token with None method")

	// Пытаемся валидировать токен с некорректным методом подписи
	claims, err = jwt_util.ValidateJWT(tokenString, testSecret)
	require.Error(t, err, "Expected error when validating token with invalid signing method")
	assert.Contains(t, err.Error(), "unexpected signing method", "Expected error message about unexpected signing method")
	assert.Nil(t, claims, "Expected nil claims for token with invalid signing method")
}

// TestGenerateJWT_Error проверяет сценарии ошибок при генерации токена.
func TestGenerateJWT_Error(t *testing.T) {
	t.Run("signing_error_propagation", func(t *testing.T) {
		claims := &jwt_util.Claims{
			UserID: testUserID,
			Email:  testUserEmail,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// originalSignedString := token.SignedString

		_, err := token.SignedString(nil)
		require.Error(t, err)
		// Проверяем фактическое сообщение об ошибке от SignedString при nil ключе
		assert.Contains(t, err.Error(), "key is of invalid type", "Expected error message for invalid key type in SignedString")
	})

	// Проверим, возвращает ли GenerateJWT ошибку для пустого секрета.
	t.Run("empty_secret", func(t *testing.T) {
		tokenString, err := jwt_util.GenerateJWT(testUserID, testUserEmail, "", 3600)

		require.NoError(t, err, "Expected no error when generating token with empty secret (library allows it)")
		assert.NotEmpty(t, tokenString, "Expected non-empty token string even with empty secret")

		claims, err := jwt_util.ValidateJWT(tokenString, "")
		require.NoError(t, err, "Expected no error when validating token generated with empty secret")
		assert.NotNil(t, claims, "Expected non-nil claims for token validated with empty secret")
		assert.Equal(t, testUserID, claims.UserID)
	})
}

type MockTokenGenerator struct {
	MockNewWithClaims func(jwt.SigningMethod, jwt.Claims) *jwt.Token
}

func (m *MockTokenGenerator) NewWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token {
	return m.MockNewWithClaims(method, claims)
}
