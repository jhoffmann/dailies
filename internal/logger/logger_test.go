package logger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextKey(t *testing.T) {
	if ErrorKey != ContextKey("error") {
		t.Errorf("Expected ErrorKey to be 'error', got %s", string(ErrorKey))
	}

	if InfoKey != ContextKey("info") {
		t.Errorf("Expected InfoKey to be 'info', got %s", string(InfoKey))
	}

	if WarningKey != ContextKey("warning") {
		t.Errorf("Expected WarningKey to be 'warning', got %s", string(WarningKey))
	}

	if DebugKey != ContextKey("debug") {
		t.Errorf("Expected DebugKey to be 'debug', got %s", string(DebugKey))
	}

	if FatalKey != ContextKey("fatal") {
		t.Errorf("Expected FatalKey to be 'fatal', got %s", string(FatalKey))
	}
}

func TestErrorInfo(t *testing.T) {
	errorInfo := ErrorInfo{
		Message: "Test error",
		Code:    500,
	}

	if errorInfo.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got %s", errorInfo.Message)
	}

	if errorInfo.Code != 500 {
		t.Errorf("Expected code 500, got %d", errorInfo.Code)
	}
}

func TestLoggedError(t *testing.T) {
	t.Run("sets error context and writes HTTP error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		LoggedError(w, "Test error message", http.StatusInternalServerError, req)

		// Check HTTP response
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		body := strings.TrimSpace(w.Body.String())
		if !strings.Contains(body, "Test error message") {
			t.Errorf("Expected body to contain 'Test error message', got: %s", body)
		}

		// Check context
		errorInfo, ok := req.Context().Value(ErrorKey).(ErrorInfo)
		if !ok {
			t.Error("Expected error info to be stored in context")
		}

		if errorInfo.Message != "Test error message" {
			t.Errorf("Expected context message 'Test error message', got %s", errorInfo.Message)
		}

		if errorInfo.Code != http.StatusInternalServerError {
			t.Errorf("Expected context code 500, got %d", errorInfo.Code)
		}
	})

	t.Run("works with different error codes", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/test", nil)
		w := httptest.NewRecorder()

		LoggedError(w, "Bad request", http.StatusBadRequest, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		errorInfo, ok := req.Context().Value(ErrorKey).(ErrorInfo)
		if !ok {
			t.Error("Expected error info to be stored in context")
		}

		if errorInfo.Code != http.StatusBadRequest {
			t.Errorf("Expected context code 400, got %d", errorInfo.Code)
		}
	})

	t.Run("works with empty error message", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/empty", nil)
		w := httptest.NewRecorder()

		LoggedError(w, "", http.StatusNotFound, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		errorInfo, ok := req.Context().Value(ErrorKey).(ErrorInfo)
		if !ok {
			t.Error("Expected error info to be stored in context")
		}

		if errorInfo.Message != "" {
			t.Errorf("Expected empty message, got %s", errorInfo.Message)
		}
	})

	t.Run("preserves other context values", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/context", nil)

		// Add some existing context value
		ctx := req.Context()
		ctx = context.WithValue(ctx, InfoKey, "existing info")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		LoggedError(w, "Error with existing context", http.StatusConflict, req)

		// Check that error was added
		errorInfo, ok := req.Context().Value(ErrorKey).(ErrorInfo)
		if !ok {
			t.Error("Expected error info to be stored in context")
		}

		if errorInfo.Message != "Error with existing context" {
			t.Errorf("Expected error message, got %s", errorInfo.Message)
		}

		// Check that existing context value is preserved
		existingInfo, ok := req.Context().Value(InfoKey).(string)
		if !ok {
			t.Error("Expected existing info to be preserved in context")
		}

		if existingInfo != "existing info" {
			t.Errorf("Expected 'existing info', got %s", existingInfo)
		}
	})

	t.Run("sets correct content type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/content-type", nil)
		w := httptest.NewRecorder()

		LoggedError(w, "Content type test", http.StatusTeapot, req)

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/plain") {
			t.Errorf("Expected Content-Type to contain 'text/plain', got: %s", contentType)
		}
	})
}
