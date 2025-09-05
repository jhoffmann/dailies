// Package routes sets up HTTP routes for the web application
package routes

import (
	"net/http"

	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/ui/web"
)

// Setup configures HTTP routes for the application.
// Registers handlers for static files, root path, and calls setup functions for task and tag routes.
func Setup() {
	// Static file server for bundled assets
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/dist/"))))

	// Root path serves the main HTML template
	http.HandleFunc("/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			logger.LoggedError(w, "Not Found", http.StatusNotFound, r)
			return
		}
		web.ServeIndex(w, r)
	}))

	// Health check endpoint
	http.HandleFunc("/healthz", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}
		api.HealthCheck(w, r)
	}))

	// Setup task and tag routes
	SetupTaskRoutes()
	SetupTagRoutes()
}
