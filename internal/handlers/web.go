package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// ServeIndex handles GET requests to serve the main HTML template.
// Renders the index.html template for the single-page application.
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
