package common_handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	// Assuming common_handler package is at this path relative to your test file
	// "github.com/IlyushinDM/user-order-api/internal/common_handler" // Adjust import path if necessary
)

// Helper function to create a mock gin.Context
func createTestContext(method, path string, queryParams map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(method, path, nil)

	// Add query parameters
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	c.Request = req
	return c, w
}

// Mock Logger to prevent actual logging during tests and potentially check calls (though not done here)
type mockLogger struct {
	*logrus.Logger
}

func newMockLogger() *mockLogger {
	logger := logrus.New()
	logger.SetOutput(nil)              // Discard output during tests
	logger.SetLevel(logrus.PanicLevel) // Suppress warnings/errors unless needed for specific test
	return &mockLogger{logger}
}

func TestNewCommonHandler(t *testing.T) {
	// Test case 1: Providing a non-nil logger
	t.Run("With provided logger", func(t *testing.T) {
		mockLog := newMockLogger()
		handler := NewCommonHandler(mockLog.Logger)

		if handler == nil {
			t.Error("NewCommonHandler returned nil for a valid logger")
		}
		// Further checks could verify if the internal log field is the provided one,
		// but this would require exporting the field or using reflection, which is often avoided.
		// Checking for nil is usually sufficient here.
	})

	// Test case 2: Providing a nil logger
	t.Run("With nil logger", func(t *testing.T) {
		// This test relies on NewCommonHandler using a default logger internally.
		// We can't easily check *which* logger is used without reflection or exporting,
		// but we can check that a handler instance is still returned.
		handler := NewCommonHandler(nil)

		if handler == nil {
			t.Error("NewCommonHandler returned nil for a nil logger")
		}
		// Note: Testing the warning log for nil logger is complex without mocking logrus at a deeper level.
		// The current implementation logs a warning and uses a default; this test confirms the handler is created.
	})
}

func TestGetPaginationParams(t *testing.T) {
	mockLog := newMockLogger()
	handler := NewCommonHandler(mockLog.Logger)

	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedPage  int
		expectedLimit int
		expectedErr   error
	}{
		{
			name:          "Valid params",
			queryParams:   map[string]string{"page": "5", "limit": "20"},
			expectedPage:  5,
			expectedLimit: 20,
			expectedErr:   nil,
		},
		{
			name:          "Default params",
			queryParams:   map[string]string{},
			expectedPage:  1,  // c.DefaultQuery provides defaults
			expectedLimit: 10, // c.DefaultQuery provides defaults
			expectedErr:   nil,
		},
		{
			name:          "Page default",
			queryParams:   map[string]string{"limit": "50"},
			expectedPage:  1,
			expectedLimit: 50,
			expectedErr:   nil,
		},
		{
			name:          "Limit default",
			queryParams:   map[string]string{"page": "3"},
			expectedPage:  3,
			expectedLimit: 10,
			expectedErr:   nil,
		},
		{
			name:          "Invalid page format",
			queryParams:   map[string]string{"page": "abc", "limit": "10"},
			expectedPage:  0,
			expectedLimit: 0,
			expectedErr:   errors.New("invalid page parameter: must be a positive integer"),
		},
		{
			name:          "Invalid limit format",
			queryParams:   map[string]string{"page": "1", "limit": "xyz"},
			expectedPage:  0,
			expectedLimit: 0,
			expectedErr:   errors.New("invalid limit parameter: must be a positive integer"),
		},
		{
			name:          "Zero page",
			queryParams:   map[string]string{"page": "0", "limit": "10"},
			expectedPage:  0,
			expectedLimit: 0,
			expectedErr:   errors.New("invalid page parameter: must be a positive integer"),
		},
		{
			name:          "Zero limit",
			queryParams:   map[string]string{"page": "1", "limit": "0"},
			expectedPage:  0,
			expectedLimit: 0,
			expectedErr:   errors.New("invalid limit parameter: must be a positive integer"),
		},
		{
			name:          "Negative page",
			queryParams:   map[string]string{"page": "-5", "limit": "10"},
			expectedPage:  0,
			expectedLimit: 0,
			expectedErr:   errors.New("invalid page parameter: must be a positive integer"),
		},
		{
			name:          "Negative limit",
			queryParams:   map[string]string{"page": "1", "limit": "-10"},
			expectedPage:  0,
			expectedLimit: 0,
			expectedErr:   errors.New("invalid limit parameter: must be a positive integer"),
		},
		{
			name:          "Limit exceeds max (100)",
			queryParams:   map[string]string{"page": "1", "limit": "150"},
			expectedPage:  1,
			expectedLimit: 100, // Should be capped at 100
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := createTestContext(http.MethodGet, "/", tt.queryParams)
			page, limit, err := handler.GetPaginationParams(c)

			if page != tt.expectedPage {
				t.Errorf("Expected page %d, got %d", tt.expectedPage, page)
			}
			if limit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, limit)
			}

			if tt.expectedErr != nil && err == nil {
				t.Errorf("Expected error, but got nil")
			} else if tt.expectedErr == nil && err != nil {
				t.Errorf("Expected no error, but got %v", err)
			} else if tt.expectedErr != nil && err != nil && tt.expectedErr.Error() != err.Error() {
				t.Errorf("Expected error '%v', got '%v'", tt.expectedErr, err)
			}
		})
	}
}

func TestGetFilteringParams(t *testing.T) {
	mockLog := newMockLogger()
	handler := NewCommonHandler(mockLog.Logger)

	tests := []struct {
		name            string
		queryParams     map[string]string
		expectedFilters map[string]interface{}
		expectedErr     error
	}{
		{
			name:            "Valid min_age, max_age, name",
			queryParams:     map[string]string{"min_age": "20", "max_age": "30", "name": "John"},
			expectedFilters: map[string]interface{}{"min_age": 20, "max_age": 30, "name": "John"},
			expectedErr:     nil,
		},
		{
			name:            "Only name filter",
			queryParams:     map[string]string{"name": "Alice"},
			expectedFilters: map[string]interface{}{"name": "Alice"},
			expectedErr:     nil,
		},
		{
			name:            "Only min_age filter",
			queryParams:     map[string]string{"min_age": "18"},
			expectedFilters: map[string]interface{}{"min_age": 18},
			expectedErr:     nil,
		},
		{
			name:            "Only max_age filter",
			queryParams:     map[string]string{"max_age": "65"},
			expectedFilters: map[string]interface{}{"max_age": 65},
			expectedErr:     nil,
		},
		{
			name:            "No filters",
			queryParams:     map[string]string{},
			expectedFilters: map[string]interface{}{}, // Should return empty map
			expectedErr:     nil,
		},
		{
			name:            "Invalid min_age format",
			queryParams:     map[string]string{"min_age": "abc", "max_age": "30"},
			expectedFilters: nil, // Should return nil map on error
			expectedErr:     errors.New("invalid min_age parameter: must be a positive integer"),
		},
		{
			name:            "Invalid max_age format",
			queryParams:     map[string]string{"min_age": "20", "max_age": "xyz"},
			expectedFilters: nil, // Should return nil map on error
			expectedErr:     errors.New("invalid max_age parameter: must be a positive integer"),
		},
		{
			name:            "Zero min_age",
			queryParams:     map[string]string{"min_age": "0"},
			expectedFilters: nil, // Should return nil map on error
			expectedErr:     errors.New("invalid min_age parameter: must be a positive integer"),
		},
		{
			name:            "Zero max_age",
			queryParams:     map[string]string{"max_age": "0"},
			expectedFilters: nil, // Should return nil map on error
			expectedErr:     errors.New("invalid max_age parameter: must be a positive integer"),
		},
		{
			name:            "Negative min_age",
			queryParams:     map[string]string{"min_age": "-10"},
			expectedFilters: nil, // Should return nil map on error
			expectedErr:     errors.New("invalid min_age parameter: must be a positive integer"),
		},
		{
			name:            "Negative max_age",
			queryParams:     map[string]string{"max_age": "-5"},
			expectedFilters: nil, // Should return nil map on error
			expectedErr:     errors.New("invalid max_age parameter: must be a positive integer"),
		},
		{
			name:            "Invalid min_age and valid max_age",
			queryParams:     map[string]string{"min_age": "abc", "max_age": "30"},
			expectedFilters: nil,
			expectedErr:     errors.New("invalid min_age parameter: must be a positive integer"),
		},
		{
			name:            "Valid min_age and invalid max_age",
			queryParams:     map[string]string{"min_age": "20", "max_age": "xyz"},
			expectedFilters: nil,
			expectedErr:     errors.New("invalid max_age parameter: must be a positive integer"),
		},
		{
			name:            "Both invalid age formats",
			queryParams:     map[string]string{"min_age": "abc", "max_age": "xyz"},
			expectedFilters: nil, // Should return nil map on the first error encountered
			expectedErr:     errors.New("invalid min_age parameter: must be a positive integer"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := createTestContext(http.MethodGet, "/", tt.queryParams)
			filters, err := handler.GetFilteringParams(c)

			// Compare errors
			if tt.expectedErr != nil && err == nil {
				t.Errorf("Expected error, but got nil")
			} else if tt.expectedErr == nil && err != nil {
				t.Errorf("Expected no error, but got %v", err)
			} else if tt.expectedErr != nil && err != nil && tt.expectedErr.Error() != err.Error() {
				t.Errorf("Expected error '%v', got '%v'", tt.expectedErr, err)
			}

			// Compare filters only if no error was expected/returned
			if tt.expectedErr == nil && err == nil {
				// Compare maps - need to iterate and check key-value pairs
				if len(filters) != len(tt.expectedFilters) {
					t.Errorf("Expected %d filters, got %d", len(tt.expectedFilters), len(filters))
				} else {
					for key, expectedValue := range tt.expectedFilters {
						actualValue, ok := filters[key]
						if !ok {
							t.Errorf("Expected filter key '%s' not found in result", key)
						} else if actualValue != expectedValue {
							t.Errorf("Expected filter '%s' value '%v', got '%v'", key, expectedValue, actualValue)
						}
					}
				}
			} else if tt.expectedErr != nil && filters != nil {
				// If an error was expected but filters map is not nil
				t.Errorf("Expected nil filters map on error, but got %v", filters)
			}
		})
	}
}
