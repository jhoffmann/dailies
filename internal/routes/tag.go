// Package routes sets up HTTP routes for the web application
package routes

import (
	"net/http"
	"strings"

	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
)

// SetupTagRoutes configures HTTP routes for tag management.
// Registers API endpoints for tags.
func SetupTagRoutes() {
	// Tag API endpoints
	http.HandleFunc("/tags", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetTags(w, r)
		case http.MethodPost:
			api.CreateTag(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	http.HandleFunc("/tags/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/tags/") == "" {
			logger.LoggedError(w, "Tag ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.GetTag(w, r)
		case http.MethodPut:
			api.UpdateTag(w, r)
		case http.MethodDelete:
			api.DeleteTag(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))
}
