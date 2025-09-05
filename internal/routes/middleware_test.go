package routes

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jhoffmann/dailies/internal/logger"
)

func TestResponseWriter(t *testing.T) {
	t.Run("wraps ResponseWriter and captures status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		wrapped.WriteHeader(404)

		if wrapped.statusCode != 404 {
			t.Errorf("Expected status code 404, got %d", wrapped.statusCode)
		}

		if w.Code != 404 {
			t.Errorf("Expected underlying writer to have code 404, got %d", w.Code)
		}
	})

	t.Run("defaults to 200 when WriteHeader not called", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		if wrapped.statusCode != 200 {
			t.Errorf("Expected default status code 200, got %d", wrapped.statusCode)
		}
	})
}

func TestLogMiddleware(t *testing.T) {
	t.Run("logs successful request", func(t *testing.T) {
		// Capture log output
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)

		handler := LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		logStr := logOutput.String()
		if !strings.Contains(logStr, "GET /test - 200") {
			t.Errorf("Expected log to contain 'GET /test - 200', got: %s", logStr)
		}
	})

	t.Run("logs request with different status code", func(t *testing.T) {
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)

		handler := LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		req := httptest.NewRequest("POST", "/api/tasks", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		logStr := logOutput.String()
		if !strings.Contains(logStr, "POST /api/tasks - 404") {
			t.Errorf("Expected log to contain 'POST /api/tasks - 404', got: %s", logStr)
		}
	})

	t.Run("logs request with error context", func(t *testing.T) {
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)

		handler := LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
			// Add error to context
			errorInfo := logger.ErrorInfo{Message: "Test error"}
			ctx := context.WithValue(r.Context(), logger.ErrorKey, errorInfo)
			*r = *r.WithContext(ctx)
			w.WriteHeader(http.StatusInternalServerError)
		})

		req := httptest.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		logStr := logOutput.String()
		if !strings.Contains(logStr, "GET /error - 500") {
			t.Errorf("Expected log to contain 'GET /error - 500', got: %s", logStr)
		}
		if !strings.Contains(logStr, "Test error") {
			t.Errorf("Expected log to contain 'Test error', got: %s", logStr)
		}
	})

	t.Run("logs timing information", func(t *testing.T) {
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)

		handler := LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/timing", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		logStr := logOutput.String()
		if !strings.Contains(logStr, "GET /timing - 200") {
			t.Errorf("Expected log to contain 'GET /timing - 200', got: %s", logStr)
		}
		// Check that timing information is included (should contain time units)
		if !strings.Contains(logStr, "µs") && !strings.Contains(logStr, "ms") && !strings.Contains(logStr, "s") {
			t.Errorf("Expected log to contain timing information, got: %s", logStr)
		}
	})

	t.Run("defaults to 200 when WriteHeader not called", func(t *testing.T) {
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)

		handler := LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
			// Don't call WriteHeader, should default to 200
			w.Write([]byte("test"))
		})

		req := httptest.NewRequest("GET", "/default", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		logStr := logOutput.String()
		if !strings.Contains(logStr, "GET /default - 200") {
			t.Errorf("Expected log to contain 'GET /default - 200', got: %s", logStr)
		}
	})
}
