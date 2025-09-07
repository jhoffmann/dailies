// Package api implements HTTP handlers for frequency management
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

// GetFrequencies handles GET requests to retrieve frequencies with optional name filtering.
//
//	@Summary		List frequencies
//	@Description	Get all frequencies with optional name filtering
//	@Tags			frequencies
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"Filter by frequency name (partial matching)"
//	@Success		200		{array}		models.Frequency
//	@Failure		500		{object}	map[string]string
//	@Router			/frequencies [get]
func GetFrequencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	nameFilter := r.URL.Query().Get("name")

	frequencies, err := GetFrequenciesWithFilter(nameFilter)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	json.NewEncoder(w).Encode(frequencies)
}

// GetFrequency handles GET requests to retrieve a single frequency by ID.
//
//	@Summary		Get frequency by ID
//	@Description	Get a single frequency by its ID
//	@Tags			frequencies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Frequency ID"
//	@Success		200	{object}	models.Frequency
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/frequencies/{id} [get]
func GetFrequency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/frequencies/"):]
	frequencyID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid frequency ID", http.StatusBadRequest, r)
		return
	}

	frequency, err := GetFrequencyByID(frequencyID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	json.NewEncoder(w).Encode(frequency)
}

// CreateFrequency handles POST requests to create a new frequency.
//
//	@Summary		Create a new frequency
//	@Description	Create a new frequency with name and cron schedule
//	@Tags			frequencies
//	@Accept			json
//	@Produce		json
//	@Param			frequency	body		object{name=string,reset=string}	true	"Frequency data"
//	@Success		201			{object}	models.Frequency
//	@Failure		400			{object}	map[string]string
//	@Router			/frequencies [post]
func CreateFrequency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var frequencyData struct {
		Name  string `json:"name"`
		Reset string `json:"reset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&frequencyData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	frequency, err := CreateFrequencyWithData(frequencyData.Name, frequencyData.Reset)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusBadRequest, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(frequency)
}

// CreateFrequencyWithData creates a new frequency with the provided data.
// This function contains the business logic for frequency creation.
func CreateFrequencyWithData(name, reset string) (*models.Frequency, error) {
	if name == "" {
		return nil, errors.New("frequency name is required")
	}

	if reset == "" {
		return nil, errors.New("frequency reset is required")
	}

	frequency := models.Frequency{
		Name:  name,
		Reset: reset,
	}

	err := frequency.Create(database.GetDB())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, errors.New("frequency name must be unique")
		}
		return nil, err
	}

	return &frequency, nil
}

// GetFrequenciesWithFilter retrieves frequencies with optional name filtering.
// This function contains the business logic for frequency retrieval.
func GetFrequenciesWithFilter(nameFilter string) ([]models.Frequency, error) {
	return models.GetFrequencies(database.GetDB(), nameFilter)
}

// GetFrequencyByID retrieves a single frequency by ID.
// This function contains the business logic for single frequency retrieval.
func GetFrequencyByID(frequencyID uuid.UUID) (*models.Frequency, error) {
	var frequency models.Frequency
	err := frequency.LoadByID(database.GetDB(), frequencyID)
	if err != nil {
		return nil, err
	}
	return &frequency, nil
}

// UpdateFrequencyByID updates an existing frequency by ID.
// This function contains the business logic for frequency updates.
func UpdateFrequencyByID(frequencyID uuid.UUID, updateData *models.Frequency) (*models.Frequency, error) {
	var frequency models.Frequency
	err := frequency.LoadByID(database.GetDB(), frequencyID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.New("frequency not found")
		}
		return nil, err
	}

	err = frequency.Update(database.GetDB(), updateData)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, errors.New("frequency name must be unique")
		}
		return nil, err
	}

	return &frequency, nil
}

// DeleteFrequencyByID deletes a frequency by ID.
// This function contains the business logic for frequency deletion.
func DeleteFrequencyByID(frequencyID uuid.UUID) error {
	var frequency models.Frequency
	frequency.ID = frequencyID
	return frequency.Delete(database.GetDB())
}

// UpdateFrequency handles PUT requests to update an existing frequency by ID.
//
//	@Summary		Update frequency
//	@Description	Update an existing frequency by ID
//	@Tags			frequencies
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string				true	"Frequency ID"
//	@Param			frequency	body		models.Frequency	true	"Frequency update data"
//	@Success		200			{object}	models.Frequency
//	@Failure		400			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/frequencies/{id} [put]
func UpdateFrequency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/frequencies/"):]
	frequencyID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid frequency ID", http.StatusBadRequest, r)
		return
	}

	var updateData models.Frequency
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	frequency, err := UpdateFrequencyByID(frequencyID, &updateData)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		} else if strings.Contains(err.Error(), "unique") {
			logger.LoggedError(w, err.Error(), http.StatusBadRequest, r)
		} else {
			logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		}
		return
	}

	json.NewEncoder(w).Encode(frequency)
}

// DeleteFrequency handles DELETE requests to remove a frequency by ID.
//
//	@Summary		Delete frequency
//	@Description	Delete a frequency by ID
//	@Tags			frequencies
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Frequency ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/frequencies/{id} [delete]
func DeleteFrequency(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/frequencies/"):]
	frequencyID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid frequency ID", http.StatusBadRequest, r)
		return
	}

	err = DeleteFrequencyByID(frequencyID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
