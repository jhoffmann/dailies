// Package models defines the data structures and database models for the dailies application.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task represents a daily task with optional frequency and tags.
type Task struct {
	ID          string     `json:"id" gorm:"type:text;primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Description *string    `json:"description,omitempty"`
	Completed   bool       `json:"completed" gorm:"default:false"`
	Priority    *int       `json:"priority,omitempty" gorm:"check:priority >= 1 AND priority <= 5"`
	FrequencyID *string    `json:"frequency_id,omitempty" gorm:"type:text"`
	Frequency   *Frequency `json:"frequency,omitempty" gorm:"foreignKey:FrequencyID"`
	Tags        []Tag      `json:"tags,omitempty" gorm:"many2many:task_tags;"`
	Deleted     bool       `json:"deleted" gorm:"default:false"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BeforeCreate is a GORM hook that generates a UUID for the task before creation.
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
