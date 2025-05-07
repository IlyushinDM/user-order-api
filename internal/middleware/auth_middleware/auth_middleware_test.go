package auth_middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/handlers/common_handler"
	"github.com/IlyushinDM/user-order-api/internal/utils/jwt_util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup
	log := logrus.New()
	jwtSecret := "test_secret" // Use a test secret
	os.Setenv("JWT_SECRET", jwtSecret)
	defer os.Unsetenv("JWT_SECRET")

	// Helper function to create a test context
	createTestContext := func(authHeader string) (*gin.Context, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}
		c.Request.Header.Set("Authorization", authHeader)
		return c, w
	}

	t.Run("Valid Token", func(t *testing.T) {
		// Generate a valid token
		userID := uint(1)
		email := "test@example.com"
		tokenString, err := jwt_util.GenerateJWT(userID, email, jwtSecret, 3600)
		assert.NoError(t, err)

		c, w := createTestContext("Bearer " + tokenString)
		AuthMiddleware(log)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, userID, c.MustGet("userID"))
		assert.Equal(t, email, c.MustGet("userEmail"))
		assert.False(t, c.IsAborted())
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		c, w := createTestContext("")
		AuthMiddleware(log)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
		var errorResponse common_handler.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Authorization header required", errorResponse.Error)
	})

	t.Run("Invalid Authorization Header Format", func(t *testing.T) {
		c, w := createTestContext("InvalidFormat")
		AuthMiddleware(log)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
		var errorResponse common_handler.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Authorization header format must be Bearer {token}", errorResponse.Error)
	})

	t.Run("Invalid JWT Token", func(t *testing.T) {
		c, w := createTestContext("Bearer invalid_token")
		AuthMiddleware(log)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
		var errorResponse common_handler.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid or expired token", errorResponse.Error)
	})

	t.Run("Expired JWT Token", func(t *testing.T) {
		// Generate an expired token
		userID := uint(1)
		email := "test@example.com"
		expirationTime := time.Now().Add(-1 * time.Hour).Unix() // Token expired 1 hour ago
		claims := &jwt_util.Claims{
			UserID: userID,
			Email:  email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Unix(expirationTime, 0)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "user-order-api",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		assert.NoError(t, err)

		c, w := createTestContext("Bearer " + tokenString)
		AuthMiddleware(log)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
		var errorResponse common_handler.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Token has expired", errorResponse.Error)
	})

	t.Run("JWT_SECRET not set", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET") //Unset to test when the env is empty
		c, w := createTestContext("Bearer valid_token")
		AuthMiddleware(log)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.True(t, c.IsAborted())
		var errorResponse common_handler.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Server configuration error", errorResponse.Error)
		os.Setenv("JWT_SECRET", jwtSecret) //reset the env var
	})

	t.Run("Bypass POST /api/users", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Method: http.MethodPost,
			URL:    &url.URL{Path: "/api/users"},
			Header: make(http.Header),
		}

		AuthMiddleware(log)(c)

		assert.Equal(t, http.StatusOK, w.Code) // Should not be aborted
		assert.False(t, c.IsAborted())
	})
}
