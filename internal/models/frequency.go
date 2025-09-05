// Package models defines the Frequency model and its database operations
package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Frequency represents a frequency schedule for task completion resets.
//
//	{
//	 "id": "a4837ac1-c807-4edd-ba49-d4e37f295be7",
//	 "name": "Daily Evening",
//	 "reset": "0 18 * * *"
//	}
type Frequency struct {
	ID    uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Name  string    `json:"name" gorm:"not null;unique"`
	Reset string    `json:"reset" gorm:"not null"` // Cron expression
	Tasks []Task    `json:"tasks,omitempty" gorm:"foreignKey:FrequencyID"`
}

// BeforeCreate is a GORM hook that generates a UUID for new frequencies if not already set.
func (frequency *Frequency) BeforeCreate(tx *gorm.DB) (err error) {
	if frequency.ID == uuid.Nil {
		frequency.ID = uuid.New()
	}
	return err
}

// Save creates or updates the frequency in the database.
func (frequency *Frequency) Save(db *gorm.DB) error {
	result := db.Save(frequency)
	return result.Error
}

// Create inserts a new frequency into the database.
func (frequency *Frequency) Create(db *gorm.DB) error {
	if frequency.Name == "" {
		return errors.New("frequency name is required")
	}
	if frequency.Reset == "" {
		return errors.New("frequency reset is required")
	}
	result := db.Create(frequency)
	return result.Error
}

// LoadByID loads a frequency by its ID from the database.
func (frequency *Frequency) LoadByID(db *gorm.DB, id uuid.UUID) error {
	result := db.First(frequency, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("frequency not found")
		}
		return result.Error
	}
	return nil
}

// Update updates the frequency in the database with new data.
func (frequency *Frequency) Update(db *gorm.DB, updateData *Frequency) error {
	if updateData.Name != "" {
		frequency.Name = updateData.Name
	}
	if updateData.Reset != "" {
		frequency.Reset = updateData.Reset
	}

	result := db.Save(frequency)
	return result.Error
}

// Delete removes the frequency from the database.
func (frequency *Frequency) Delete(db *gorm.DB) error {
	result := db.Delete(frequency, "id = ?", frequency.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("frequency not found")
	}
	return nil
}

// GetFrequencies retrieves all frequencies from the database with optional name filtering.
func GetFrequencies(db *gorm.DB, nameFilter string) ([]Frequency, error) {
	var frequencies []Frequency
	query := db

	if nameFilter != "" {
		query = query.Where("name LIKE ?", "%"+nameFilter+"%")
	}

	query = query.Order("name")

	result := query.Find(&frequencies)
	return frequencies, result.Error
}
