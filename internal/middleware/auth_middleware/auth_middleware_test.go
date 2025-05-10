package auth_middleware_test

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"os"
// 	"testing"

// 	auth_mw "github.com/IlyushinDM/user-order-api/internal/middleware/auth_middleware"
// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// )

// func TestAuthMiddleware(t *testing.T) {
// 	gin.SetMode(gin.TestMode)

// 	// Setup Logger
// 	log := logrus.New()
// 	log.SetLevel(logrus.ErrorLevel) // Suppress Debug and Info logs

// 	// Set JWT_SECRET env variable for tests
// 	os.Setenv("JWT_SECRET", "test_secret")
// 	defer os.Unsetenv("JWT_SECRET")

// 	// Helper function to create a test context
// 	createTestContext := func(method, path, authHeader string) (*gin.Context, *httptest.ResponseRecorder) {
// 		w := httptest.NewRecorder()
// 		c, _ := gin.CreateTestContext(w)
// 		c.Request = &http.Request{
// 			Method: method,
// 			URL:    &url.URL{Path: path},
// 			Header: make(http.Header),
// 		}
// 		if authHeader != "" {
// 			c.Request.Header.Set("Authorization", authHeader)
// 		}
// 		return c, w
// 	}

// 	t.Run("Bypass POST /api/users", func(t *testing.T) {
// 		c, w := createTestContext(http.MethodPost, "/api/users", "")

// 		middleware := auth_mw.AuthMiddleware(log)
// 		middleware(c)

// 		assert.Equal(t, http.StatusOK, w.Code) // Should not be aborted
// 		assert.False(t, c.IsAborted())
// 	})

// 	t.Run("Missing Authorization header", func(t *testing.T) {
// 		c, w := createTestContext(http.MethodGet, "/api/orders", "")

// 		middleware := auth_mw.AuthMiddleware(log)
// 		middleware(c)

// 		assert.Equal(t, http.StatusUnauthorized, w.Code)
// 		assert.True(t, c.IsAborted())
// 		expectedResponse := `{"error":"Authorization header required"}`
// 		assert.Equal(t, expectedResponse, w.Body.String())
// 	})

// 	t.Run("Invalid Authorization header format", func(t *testing.T) {
// 		c, w := createTestContext(http.MethodGet, "/api/orders", "InvalidHeader")

// 		middleware := auth_mw.AuthMiddleware(log)
// 		middleware(c)

// 		assert.Equal(t, http.StatusUnauthorized, w.Code)
// 		assert.True(t, c.IsAborted())
// 		expectedResponse := `{"error":"Authorization header format must be Bearer {token}"}`
// 		assert.Equal(t, expectedResponse, w.Body.String())
// 	})

// 	t.Run("Invalid JWT_SECRET env variable", func(t *testing.T) {
// 		os.Unsetenv("JWT_SECRET") // Unset for this test

// 		c, w := createTestContext(http.MethodGet, "/api/orders", "Bearer valid_token") // Doesn't matter, will fail before jwt check

// 		middleware := auth_mw.AuthMiddleware(log)
// 		middleware(c)

// 		assert.Equal(t, http.StatusInternalServerError, w.Code)
// 		assert.True(t, c.IsAborted())
// 		expectedResponse := `{"error":"Server configuration error"}`
// 		assert.Equal(t, expectedResponse, w.Body.String())

// 		os.Setenv("JWT_SECRET", "test_secret") // Restore for other tests
// 	})

// 	t.Run("Invalid JWT token", func(t *testing.T) {
// 		c, w := createTestContext(http.MethodGet, "/api/orders", "Bearer invalid_token")

// 		middleware := auth_mw.AuthMiddleware(log)
// 		middleware(c)

// 		assert.Equal(t, http.StatusUnauthorized, w.Code)
// 		assert.True(t, c.IsAborted())
// 		expectedResponse := `{"error":"Invalid or expired token"}`
// 		assert.Equal(t, expectedResponse, w.Body.String())
// 	})

// 	// t.Run("Expired JWT token", func(t *testing.T) {
// 	// 	// Generate an expired token
// 	// 	claims := &jwt_util.Claims{
// 	// 		UserID: 123,
// 	// 		Email:  "test@example.com",
// 	// 		RegisteredClaims: jwt.RegisteredClaims{
// 	// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
// 	// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 	// 		},
// 	// 	}

// 	// 	expirationSeconds := 3600
// 	// 	token, err := jwt_util.GenerateJWT(claims.UserID, claims.Email, "test_secret", expirationSeconds)
// 	// 	assert.NoError(t, err)

// 	// 	c, w := createTestContext(http.MethodGet, "/api/orders", "Bearer "+token)

// 	// 	middleware := auth_mw.AuthMiddleware(log)
// 	// 	middleware(c)

// 	// 	assert.Equal(t, http.StatusUnauthorized, w.Code)
// 	// 	assert.True(t, c.IsAborted())
// 	// 	expectedResponse := `{"error":"Token has expired"}`
// 	// 	assert.Equal(t, expectedResponse, w.Body.String())
// 	// })

// 	// t.Run("Valid JWT token", func(t *testing.T) {
// 	// 	claims := &jwt_util.Claims{
// 	// 		UserID: 456,
// 	// 		Email:  "valid@example.com",
// 	// 		RegisteredClaims: jwt.RegisteredClaims{
// 	// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // Expires in 1 hour
// 	// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 	// 		},
// 	// 	}

// 	// 	expirationSeconds := 3600
// 	// 	token, err := jwt_util.GenerateJWT(claims.UserID, claims.Email, "test_secret", expirationSeconds)
// 	// 	assert.NoError(t, err)

// 	// 	c, w := createTestContext(http.MethodGet, "/api/orders", "Bearer "+token)

// 	// 	middleware := auth_mw.AuthMiddleware(log)
// 	// 	middleware(c)

// 	// 	assert.Equal(t, http.StatusOK, w.Code) // Should not be aborted
// 	// 	assert.False(t, c.IsAborted())

// 	// 	userID, exists := c.Get("userID")
// 	// 	assert.True(t, exists)
// 	// 	assert.Equal(t, 456, userID)

// 	// 	userEmail, exists := c.Get("userEmail")
// 	// 	assert.True(t, exists)
// 	// 	assert.Equal(t, "valid@example.com", userEmail)
// 	// })
// }
