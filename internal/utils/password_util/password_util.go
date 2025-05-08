package password_util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// generateFromPasswordWrapper позволяет мокать bcrypt.GenerateFromPassword в тестах
var generateFromPasswordWrapper = bcrypt.GenerateFromPassword

func HashPassword(password string) (string, error) {
	bytes, err := generateFromPasswordWrapper([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash: %w", err)
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
