package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task represents a daily task with tracking information.
type Task struct {
	ID           uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Name         string    `json:"name" gorm:"not null"`
	DateCreated  time.Time `json:"date_created" gorm:"autoCreateTime"`
	DateModified time.Time `json:"date_modified" gorm:"autoUpdateTime"`
	Completed    bool      `json:"completed" gorm:"default:false"`
}

// BeforeCreate is a GORM hook that generates a UUID for new tasks if not already set.
func (task *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	return
}
