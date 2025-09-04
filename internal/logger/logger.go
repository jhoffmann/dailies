// Package logger provides utilities for logging errors in HTTP handlers
package logger

import (
	"context"
	"net/http"
)

// ContextKey represents a context key type to avoid string collision
type ContextKey string

// Common message type constants for context keys
const (
	ErrorKey   ContextKey = "error"
	InfoKey    ContextKey = "info"
	WarningKey ContextKey = "warning"
	DebugKey   ContextKey = "debug"
	FatalKey   ContextKey = "fatal"
)

// ErrorInfo stores error details in request context
type ErrorInfo struct {
	Message string
	Code    int
}

// LoggedError stores error info in request context for deferred logging
func LoggedError(w http.ResponseWriter, error string, code int, r *http.Request) {
	// Store error info in request context
	ctx := context.WithValue(r.Context(), ErrorKey, ErrorInfo{
		Message: error,
		Code:    code,
	})
	*r = *r.WithContext(ctx)

	http.Error(w, error, code)
}
