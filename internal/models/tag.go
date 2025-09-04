// Package models defines the Tag model and its database operations
package models

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// tagColors contains a selection of light colors that work well with black text
var tagColors = []string{
	"#fecaca", // red-200
	"#fed7aa", // orange-200
	"#fde68a", // yellow-200
	"#bbf7d0", // green-200
	"#a7f3d0", // emerald-200
	"#a5f3fc", // cyan-200
	"#bfdbfe", // blue-200
	"#c7d2fe", // indigo-200
	"#ddd6fe", // violet-200
	"#e9d5ff", // purple-200
	"#f5d0fe", // fuchsia-200
	"#fce7f3", // pink-200
	"#fecdd3", // rose-200
	"#e2e8f0", // slate-200
	"#e5e7eb", // gray-200
	"#e7e5e4", // stone-200
	"#fca5a5", // red-300
	"#fdba74", // orange-300
	"#fcd34d", // yellow-300
	"#86efac", // green-300
	"#6ee7b7", // emerald-300
	"#67e8f9", // cyan-300
	"#93c5fd", // blue-300
	"#a5b4fc", // indigo-300
	"#c4b5fd", // violet-300
	"#d8b4fe", // purple-300
	"#f0abfc", // fuchsia-300
	"#f9a8d4", // pink-300
	"#fda4af", // rose-300
	"#cbd5e1", // slate-300
	"#d1d5db", // gray-300
	"#d6d3d1", // stone-300
}

// Tag represents a tag that can be associated with tasks.
//
//	{
//	 "id": "a4837ac1-c807-4edd-ba49-d4e37f295be7",
//	 "name": "personal",
//	 "color": "#3b82f6"
//	}
type Tag struct {
	ID    uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Name  string    `json:"name" gorm:"not null;unique"`
	Color string    `json:"color" gorm:"not null"`
	Tasks []Task    `json:"tasks,omitempty" gorm:"many2many:task_tags;"`
}

// BeforeCreate is a GORM hook that generates a UUID for new tags if not already set.
func (tag *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if tag.ID == uuid.Nil {
		tag.ID = uuid.New()
	}
	return err
}

// Save creates or updates the tag in the database.
func (tag *Tag) Save(db *gorm.DB) error {
	result := db.Save(tag)
	return result.Error
}

// Create inserts a new tag into the database.
func (tag *Tag) Create(db *gorm.DB) error {
	if tag.Name == "" {
		return errors.New("tag name is required")
	}

	// Assign random color if not specified
	if tag.Color == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		tag.Color = tagColors[r.Intn(len(tagColors))]
	}

	result := db.Create(tag)
	return result.Error
}

// LoadByID loads a tag by its ID from the database.
func (tag *Tag) LoadByID(db *gorm.DB, id uuid.UUID) error {
	result := db.First(tag, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("tag not found")
		}
		return result.Error
	}
	return nil
}

// Update updates the tag in the database with new data.
func (tag *Tag) Update(db *gorm.DB, updateData *Tag) error {
	if updateData.Name != "" {
		tag.Name = updateData.Name
	}
	if updateData.Color != "" {
		tag.Color = updateData.Color
	}

	result := db.Save(tag)
	return result.Error
}

// Delete removes the tag from the database.
func (tag *Tag) Delete(db *gorm.DB) error {
	result := db.Delete(tag, "id = ?", tag.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("tag not found")
	}
	return nil
}

// GetTags retrieves all tags from the database with optional name filtering.
func GetTags(db *gorm.DB, nameFilter string) ([]Tag, error) {
	var tags []Tag
	query := db

	if nameFilter != "" {
		query = query.Where("name LIKE ?", "%"+nameFilter+"%")
	}

	query = query.Order("name")

	result := query.Find(&tags)
	return tags, result.Error
}
