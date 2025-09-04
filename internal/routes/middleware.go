package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/jhoffmann/dailies/internal/logger"
)

// responseWriter wraps http.ResponseWriter to capture status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LogMiddleware wraps handlers to log all incoming requests with status codes and timing
func LogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // default if WriteHeader isn't called
		}

		start := time.Now()
		next(wrapped, r)
		duration := time.Since(start)

		// Check if an error was stored in context
		if errorInfo, ok := r.Context().Value(logger.ErrorKey).(logger.ErrorInfo); ok {
			log.Printf("Request: %s %s - %d - %s - %v",
				r.Method, r.URL.Path, wrapped.statusCode, errorInfo.Message, duration)
		} else {
			log.Printf("Request: %s %s - %d - %v",
				r.Method, r.URL.Path, wrapped.statusCode, duration)
		}
	}
}
