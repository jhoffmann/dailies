package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag represents a categorization label that can be assigned to tasks.
type Tag struct {
	ID        string    `json:"id" gorm:"type:text;primaryKey"`
	Name      string    `json:"name" gorm:"not null;unique"`
	Color     string    `json:"color" gorm:"not null"`
	Tasks     []Task    `json:"tasks,omitempty" gorm:"many2many:task_tags;"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate is a GORM hook that generates a UUID for the tag before creation.
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
