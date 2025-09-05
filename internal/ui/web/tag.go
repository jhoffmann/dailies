// Package web contains HTTP handlers for serving HTML templates and components
package web

import (
	"net/http"

	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
)

// GetTagsHTML returns HTML snippet for tag list (for HTMX).
func GetTagsHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	nameFilter := r.URL.Query().Get("name")
	tags, err := models.GetTags(database.GetDB(), nameFilter)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	html, err := componentRenderer.Render("tagsView", tags)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}
