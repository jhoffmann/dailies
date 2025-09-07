// Package api implements HTTP handlers for tag management
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
)

// GetTags handles GET requests to retrieve all tags with optional name filtering.
//
//	@Summary		List tags
//	@Description	Get all tags with optional name filtering
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"Filter by tag name (partial matching)"
//	@Success		200		{array}		models.Tag
//	@Failure		500		{object}	map[string]string
//	@Router			/tags [get]
func GetTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	nameFilter := r.URL.Query().Get("name")

	tags, err := GetTagsWithFilter(nameFilter)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	json.NewEncoder(w).Encode(tags)
}

// GetTag handles GET requests to retrieve a single tag by ID.
//
//	@Summary		Get tag by ID
//	@Description	Get a single tag by its ID
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Tag ID"
//	@Success		200	{object}	models.Tag
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/tags/{id} [get]
func GetTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/tags/"):]
	tagID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid tag ID", http.StatusBadRequest, r)
		return
	}

	tag, err := GetTagByID(tagID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	json.NewEncoder(w).Encode(tag)
}

// CreateTag handles POST requests to create a new tag.
//
//	@Summary		Create a new tag
//	@Description	Create a new tag with name and optional color
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			tag	body		object{name=string,color=string}	true	"Tag data"
//	@Success		201	{object}	models.Tag
//	@Failure		400	{object}	map[string]string
//	@Router			/tags [post]
func CreateTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var tagData struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tagData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	tag, err := CreateTagWithValidation(tagData.Name, tagData.Color)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusBadRequest, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

// UpdateTag handles PUT requests to update an existing tag by ID.
//
//	@Summary		Update tag
//	@Description	Update an existing tag by ID
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string		true	"Tag ID"
//	@Param			tag	body		models.Tag	true	"Tag update data"
//	@Success		200	{object}	models.Tag
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/tags/{id} [put]
func UpdateTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/tags/"):]
	tagID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid tag ID", http.StatusBadRequest, r)
		return
	}

	var updateData models.Tag
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	tag, err := UpdateTagByID(tagID, &updateData)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		} else {
			logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		}
		return
	}

	json.NewEncoder(w).Encode(tag)
}

// DeleteTag handles DELETE requests to remove a tag by ID.
//
//	@Summary		Delete tag
//	@Description	Delete a tag by ID
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Tag ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/tags/{id} [delete]
func DeleteTag(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/tags/"):]
	tagID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid tag ID", http.StatusBadRequest, r)
		return
	}

	err = DeleteTagByID(tagID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateTagWithValidation creates a new tag with business logic validation.
// This function contains the business logic for tag creation.
func CreateTagWithValidation(name, color string) (*models.Tag, error) {
	if name == "" {
		return nil, errors.New("tag name is required")
	}

	tag := models.Tag{
		Name:  name,
		Color: color,
	}

	err := tag.Create(database.GetDB())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: tags.name") {
			return nil, errors.New("tag name must be unique")
		}
		return nil, err
	}

	return &tag, nil
}

// GetTagsWithFilter retrieves tags with optional name filtering.
// This function contains the business logic for tag retrieval.
func GetTagsWithFilter(nameFilter string) ([]models.Tag, error) {
	return models.GetTags(database.GetDB(), nameFilter)
}

// GetTagByID retrieves a single tag by ID.
// This function contains the business logic for single tag retrieval.
func GetTagByID(tagID uuid.UUID) (*models.Tag, error) {
	var tag models.Tag
	err := tag.LoadByID(database.GetDB(), tagID)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// UpdateTagByID updates an existing tag by ID.
// This function contains the business logic for tag updates.
func UpdateTagByID(tagID uuid.UUID, updateData *models.Tag) (*models.Tag, error) {
	var tag models.Tag
	err := tag.LoadByID(database.GetDB(), tagID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}

	err = tag.Update(database.GetDB(), updateData)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// DeleteTagByID deletes a tag by ID.
// This function contains the business logic for tag deletion.
func DeleteTagByID(tagID uuid.UUID) error {
	var tag models.Tag
	tag.ID = tagID
	return tag.Delete(database.GetDB())
}
