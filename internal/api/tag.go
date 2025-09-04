// Package api implements HTTP handlers for tag management
package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
)

// GetTags handles GET requests to retrieve all tags with optional name filtering.
// Supports query parameter: name (string for partial matching).
func GetTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	nameFilter := r.URL.Query().Get("name")

	tags, err := models.GetTags(database.GetDB(), nameFilter)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	json.NewEncoder(w).Encode(tags)
}

// GetTag handles GET requests to retrieve a single tag by ID.
func GetTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/tags/"):]
	tagID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid tag ID", http.StatusBadRequest, r)
		return
	}

	var tag models.Tag
	err = tag.LoadByID(database.GetDB(), tagID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	json.NewEncoder(w).Encode(tag)
}

// CreateTag handles POST requests to create a new tag.
// Requires a JSON body with a tag name.
func CreateTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var tag models.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	err := tag.Create(database.GetDB())
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusBadRequest, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

// UpdateTag handles PUT requests to update an existing tag by ID.
// Accepts a JSON body with fields to update (name).
func UpdateTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/tags/"):]
	tagID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid tag ID", http.StatusBadRequest, r)
		return
	}

	var tag models.Tag
	err = tag.LoadByID(database.GetDB(), tagID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	var updateData models.Tag
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	err = tag.Update(database.GetDB(), &updateData)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	json.NewEncoder(w).Encode(tag)
}

// DeleteTag handles DELETE requests to remove a tag by ID.
func DeleteTag(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/tags/"):]
	tagID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid tag ID", http.StatusBadRequest, r)
		return
	}

	var tag models.Tag
	tag.ID = tagID
	err = tag.Delete(database.GetDB())
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
