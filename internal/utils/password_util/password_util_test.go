package password_util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword проверяет, что хеширование пароля работает корректно.
func TestHashPassword(t *testing.T) {
	password := "plainpassword123"
	hashedPassword, err := HashPassword(password)

	require.NoError(t, err, "HashPassword не должна возвращать ошибку")
	require.NotEmpty(t, hashedPassword, "Хешированный пароль не должен быть пустым")

	// Проверяем, что это валидный bcrypt хеш, сравнивая его с исходным паролем
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err, "Сгенерированный хеш должен соответствовать исходному паролю")
}

// TestHashPassword_EmptyInput проверяет поведение HashPassword с пустым вводом.
// bcrypt может обрабатывать пустые строки, но это может быть нежелательно.
// Данный тест просто проверяет, что функция не паникует.
func TestHashPassword_EmptyInput(t *testing.T) {
	_, err := HashPassword("")
	assert.NoError(t, err, "HashPassword не должна возвращать ошибку для пустой строки")
}

// TestHashPassword_ErrorCase проверяет обработку ошибки при генерации хеша
func TestHashPassword_ErrorCase(t *testing.T) {
	// Сохраняем оригинальную функцию
	origWrapper := generateFromPasswordWrapper
	defer func() { generateFromPasswordWrapper = origWrapper }()

	// Подменяем функцию для возврата ошибки
	generateFromPasswordWrapper = func(password []byte, cost int) ([]byte, error) {
		return nil, assert.AnError
	}

	_, err := HashPassword("password")
	assert.Error(t, err, "Должна вернуться ошибка")
	assert.Contains(t, err.Error(), "failed to generate hash")
}

func TestHashPassword_Success(t *testing.T) {
	hash, err := HashPassword("password")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

// TestCheckPasswordHash проверяет корректность сравнения пароля и хеша.
func TestCheckPasswordHash(t *testing.T) {
	password := "securePa$$w0rd"
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err, "Предварительное хеширование пароля не удалось")

	t.Run("correct_password", func(t *testing.T) {
		match := CheckPasswordHash(password, hashedPassword)
		assert.True(t, match, "CheckPasswordHash должна возвращать true для корректного пароля")
	})

	t.Run("incorrect_password", func(t *testing.T) {
		match := CheckPasswordHash("wrongpassword", hashedPassword)
		assert.False(t, match, "CheckPasswordHash должна возвращать false для некорректного пароля")
	})

	t.Run("empty_password_vs_valid_hash", func(t *testing.T) {
		match := CheckPasswordHash("", hashedPassword)
		assert.False(t, match, "CheckPasswordHash должна возвращать false для пустого пароля против валидного хеша")
	})

	t.Run("valid_password_vs_empty_hash", func(t *testing.T) {
		match := CheckPasswordHash(password, "")
		assert.False(t, match, "CheckPasswordHash должна возвращать false для валидного пароля против пустого хеша")
	})

	t.Run("valid_password_vs_invalid_hash_format", func(t *testing.T) {
		// bcrypt.CompareHashAndPassword вернет ошибку, которую CheckPasswordHash преобразует в false
		match := CheckPasswordHash(password, "not_a_valid_bcrypt_hash")
		assert.False(t, match, "CheckPasswordHash должна возвращать false для невалидного формата хеша")
	})
}
