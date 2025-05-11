package auth_middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IlyushinDM/user-order-api/internal/middleware/auth_middleware"
	"github.com/IlyushinDM/user-order-api/internal/utils/jwt_util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// mockClaims implements the claims returned by jwt_util.ValidateJWT
type mockClaims struct {
	UserID string
	Email  string
}

func TestAuthMiddleware_RegistrationBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	router.Use(auth_middleware.AuthMiddleware(log, "secret"))
	router.POST("/api/users", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	router.Use(auth_middleware.AuthMiddleware(log, "secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Требуется заголовок Authorization")
}

func TestAuthMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	router.Use(auth_middleware.AuthMiddleware(log, "secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Формат заголовка должен быть Bearer {token}")
}

func TestAuthMiddleware_MissingJWTSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	router.Use(auth_middleware.AuthMiddleware(log, ""))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка конфигурации сервера")
}

// mockValidateJWT replaces jwt_util.ValidateJWT for testing
func mockValidateJWT(tokenString, secret string) (*jwt_util.Claims, error) {
	if tokenString == "validtoken" {
		return &jwt_util.Claims{UserID: 123, Email: "test@example.com"}, nil
	}
	if tokenString == "expiredtoken" {
		return nil, jwt.ErrTokenExpired
	}
	return nil, errors.New("invalid token")
}

func TestAuthMiddleware_InvalidJWTToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	router.Use(auth_middleware.AuthMiddleware(log, "secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный или просроченный токен")
}

func TestAuthMiddleware_ExpiredJWTToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	router.Use(auth_middleware.AuthMiddleware(log, "secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer expiredtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный или просроченный токен")
}

func TestAuthMiddleware_InvalidJWTToken_CustomValidator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	router.Use(auth_middleware.AuthMiddlewareWithValidator(log, "secret", mockValidateJWT))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer sometotallyinvalidtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный или просроченный токен")
}

func TestAuthMiddleware_NextHandlerNotCalledOnFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	log := logrus.New()
	called := false
	router.Use(auth_middleware.AuthMiddlewareWithValidator(log, "secret", mockValidateJWT))
	router.GET("/protected", func(c *gin.Context) {
		called = true
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer sometotallyinvalidtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.False(t, called, "Handler should not be called on auth failure")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
