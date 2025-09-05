// Package routes sets up HTTP routes for frequency management
package routes

import (
	"net/http"
	"strings"

	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
)

// SetupFrequencyRoutes configures HTTP routes for frequency management.
// Registers API endpoints for frequency CRUD operations.
func SetupFrequencyRoutes() {
	// Frequency API endpoints
	http.HandleFunc("/frequencies", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetFrequencies(w, r)
		case http.MethodPost:
			api.CreateFrequency(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	http.HandleFunc("/frequencies/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/frequencies/") == "" {
			logger.LoggedError(w, "Frequency ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.GetFrequency(w, r)
		case http.MethodPut:
			api.UpdateFrequency(w, r)
		case http.MethodDelete:
			api.DeleteFrequency(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))
}
