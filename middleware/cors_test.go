package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test regular GET request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check CORS headers
	headers := map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "POST, OPTIONS, GET, PUT, DELETE",
	}

	for header, expectedValue := range headers {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s to be %s, got %s", header, expectedValue, actualValue)
		}
	}

	// Check Access-Control-Allow-Headers contains expected values
	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	expectedHeaders := []string{"Content-Type", "Authorization", "X-Requested-With"}
	for _, expected := range expectedHeaders {
		if !contains(allowHeaders, expected) {
			t.Errorf("Expected Access-Control-Allow-Headers to contain %s, got %s", expected, allowHeaders)
		}
	}
}

func TestCORSOptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test OPTIONS request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for OPTIONS request, got %d", http.StatusNoContent, w.Code)
	}

	// Check CORS headers are still present
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin to be *, got %s", origin)
	}
}

func TestCORSMiddlewareOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var middlewareOrder []string

	r := gin.New()
	r.Use(func(c *gin.Context) {
		middlewareOrder = append(middlewareOrder, "first")
		c.Next()
	})
	r.Use(CORS())
	r.Use(func(c *gin.Context) {
		middlewareOrder = append(middlewareOrder, "last")
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	expectedOrder := []string{"first", "last"}
	if len(middlewareOrder) != len(expectedOrder) {
		t.Errorf("Expected middleware order length %d, got %d", len(expectedOrder), len(middlewareOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(middlewareOrder) || middlewareOrder[i] != expected {
			t.Errorf("Expected middleware order[%d] to be %s, got %v", i, expected, middlewareOrder)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			hasSubstring(s, substr))))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
