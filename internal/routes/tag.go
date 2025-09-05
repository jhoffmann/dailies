// Package routes sets up HTTP routes for the web application
package routes

import (
	"net/http"
	"strings"

	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/ui/web"
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

	// Tag component routes for HTMX HTML snippets
	http.HandleFunc("/component/tags", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}
		web.GetTagsHTML(w, r)
	}))

	http.HandleFunc("/component/tag-selection", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}
		web.GetTagSelectionHTML(w, r)
	}))

	http.HandleFunc("/component/create/tag", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			web.GetTagCreateHTML(w, r)
		case http.MethodPost:
			web.CreateTagHTML(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))
}
