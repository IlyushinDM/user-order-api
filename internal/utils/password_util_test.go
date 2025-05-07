package utils

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestHashPassword(t *testing.T) {
// 	password := "password123"
// 	hashedPassword, err := HashPassword(password)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, hashedPassword)
// }

// func TestCheckPasswordHash(t *testing.T) {
// 	password := "password123"
// 	hashedPassword, err := HashPassword(password)
// 	assert.NoError(t, err)

// 	assert.True(t, CheckPasswordHash(password, hashedPassword))

// 	// Test with wrong password
// 	assert.False(t, CheckPasswordHash("wrongpassword", hashedPassword))
// }
