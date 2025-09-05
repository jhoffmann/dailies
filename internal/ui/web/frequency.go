// Package web contains HTTP handlers for serving HTML templates and components
package web

import (
	"net/http"

	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
)

// GetTimerResetsHTML returns HTML snippet for timer resets list (for HTMX).
// Shows all frequencies with their reset times.
func GetTimerResetsHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Use the API layer to get all frequencies
	frequencies, err := api.GetFrequenciesWithFilter("")
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	html, err := componentRenderer.Render("timerResets", frequencies)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}

// GetFrequencySelectionHTML returns HTML snippet for frequency selection radio buttons (for HTMX).
func GetFrequencySelectionHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	nameFilter := r.URL.Query().Get("name")

	// Use the API layer to get all frequencies
	frequencies, err := api.GetFrequenciesWithFilter(nameFilter)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	html, err := componentRenderer.Render("frequencySelection", frequencies)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}
