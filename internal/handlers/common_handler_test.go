package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetPaginationParams(t *testing.T) {
	// Setup
	router := gin.Default()
	router.GET("/test", func(c *gin.Context) {
		page, limit := getPaginationParams(c)
		c.JSON(http.StatusOK, gin.H{"page": page, "limit": limit})
	})

	// Test case 1: No parameters provided
	req1, _ := http.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.JSONEq(t, `{"page":1,"limit":10}`, w1.Body.String())

	// Test case 2: Valid parameters provided
	req2, _ := http.NewRequest("GET", "/test?page=2&limit=20", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.JSONEq(t, `{"page":2,"limit":20}`, w2.Body.String())

	// Test case 3: Invalid parameters provided
	req3, _ := http.NewRequest("GET", "/test?page=invalid&limit=-1", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.JSONEq(t, `{"page":1,"limit":10}`, w3.Body.String())

	// Test case 4: Limit exceeds maximum
	req4, _ := http.NewRequest("GET", "/test?limit=150", nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusOK, w4.Code)
	assert.JSONEq(t, `{"page":1,"limit":100}`, w4.Body.String())
}

func TestGetFilteringParams(t *testing.T) {
	// Setup
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("logger", &mockLogger{})
	})
	router.GET("/test", func(c *gin.Context) {
		filters := getFilteringParams(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	// Test case 1: No parameters provided
	req1, _ := http.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.JSONEq(t, `{"filters":{}}`, w1.Body.String())

	// Test case 2: Valid parameters provided
	req2, _ := http.NewRequest("GET", "/test?min_age=18&max_age=30&name=John", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), `"min_age":18`)
	assert.Contains(t, w2.Body.String(), `"max_age":30`)
	assert.Contains(t, w2.Body.String(), `"name":"John"`)

	// Test case 3: Invalid parameters provided
	req3, _ := http.NewRequest("GET", "/test?min_age=invalid&max_age=-1", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.JSONEq(t, `{"filters":{}}`, w3.Body.String())
}

// Mock logger for testing
type mockLogger struct{}

func (l *mockLogger) Warnf(format string, args ...interface{}) {}
