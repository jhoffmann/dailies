// Package web contains HTTP handlers for serving HTML templates and components
package web

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/jhoffmann/dailies/internal/components"
	"github.com/jhoffmann/dailies/internal/logger"
)

var componentRenderer = components.NewComponentRenderer()

// ServeIndex handles GET requests to serve the main HTML template.
// Renders the index.html template for the single-page application.
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		logger.LoggedError(w, "Error loading template", http.StatusInternalServerError, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, nil); err != nil {
		logger.LoggedError(w, "Error rendering template", http.StatusInternalServerError, r)
		return
	}
}
