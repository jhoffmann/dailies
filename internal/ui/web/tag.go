// Package web contains HTTP handlers for serving HTML templates and components
package web

import (
	"encoding/json"
	"net/http"
	"strings"

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

// GetTagCreateHTML returns HTML snippet for tag creation form (for HTMX).
func GetTagCreateHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html, err := componentRenderer.Render("tagCreate", nil)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}

// CreateTagHTML creates a new tag and closes modal (for HTMX).
func CreateTagHTML(w http.ResponseWriter, r *http.Request) {
	var tagData struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tagData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	if tagData.Name == "" {
		logger.LoggedError(w, "Tag name is required", http.StatusBadRequest, r)
		return
	}

	tag := models.Tag{
		Name:  tagData.Name,
		Color: tagData.Color,
	}

	if err := tag.Create(database.GetDB()); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: tags.name") {
			logger.LoggedError(w, "Tag name must be unique", http.StatusBadRequest, r)
			return
		}
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetTagSelectionHTML returns HTML snippet for tag selection checkboxes (for HTMX).
func GetTagSelectionHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	nameFilter := r.URL.Query().Get("name")
	tags, err := models.GetTags(database.GetDB(), nameFilter)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	html, err := componentRenderer.Render("tagSelection", tags)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}
