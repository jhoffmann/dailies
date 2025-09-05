// Package models defines the Task model and its database operations
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task represents a daily task with tracking information.
//
//	{
//	 "id": "a4837ac1-c807-4edd-ba49-d4e37f295be7",
//	 "name": "Review pull requests",
//	 "date_created": "2025-09-03T21:17:04.32338525-06:00",
//	 "date_modified": "2025-09-03T21:17:04.32338525-06:00",
//	 "completed": false,
//	 "priority": 3,
//	 "tags": [{"id": "...", "name": "work"}, {"id": "...", "name": "urgent"}]
//	}
type Task struct {
	ID           uuid.UUID `json:"id" gorm:"type:text;primary_key"`
	Name         string    `json:"name" gorm:"not null"`
	DateCreated  time.Time `json:"date_created" gorm:"autoCreateTime"`
	DateModified time.Time `json:"date_modified" gorm:"autoUpdateTime"`
	Completed    bool      `json:"completed" gorm:"default:false"`
	Priority     int       `json:"priority" gorm:"default:3"`
	Tags         []Tag     `json:"tags,omitempty" gorm:"many2many:task_tags;"`
}

// BeforeCreate is a GORM hook that generates a UUID for new tasks if not already set.
func (task *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	return err
}

// Save creates or updates the task in the database.
func (task *Task) Save(db *gorm.DB) error {
	result := db.Save(task)
	return result.Error
}

// Create inserts a new task into the database.
func (task *Task) Create(db *gorm.DB) error {
	if task.Name == "" {
		return errors.New("task name is required")
	}
	result := db.Create(task)
	return result.Error
}

// LoadByID loads a task by its ID from the database.
func (task *Task) LoadByID(db *gorm.DB, id uuid.UUID) error {
	result := db.Preload("Tags").First(task, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("task not found")
		}
		return result.Error
	}
	return nil
}

// Update updates the task in the database with new data.
func (task *Task) Update(db *gorm.DB, updateData *Task) error {
	if updateData.Name != "" {
		task.Name = updateData.Name
	}
	task.Completed = updateData.Completed
	if updateData.Priority >= 1 && updateData.Priority <= 5 {
		task.Priority = updateData.Priority
	}

	result := db.Save(task)
	return result.Error
}

// Delete removes the task from the database.
func (task *Task) Delete(db *gorm.DB) error {
	result := db.Delete(task, "id = ?", task.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found")
	}
	return nil
}

// GetTasks retrieves tasks with optional filtering and sorting.
// sortField can be "completed", "priority", or "name".
// tagIDFilter allows filtering tasks that have ALL specified tag IDs (AND operation).
func GetTasks(db *gorm.DB, completedFilter *bool, nameFilter string, tagIDFilter []uuid.UUID, sortField string) ([]Task, error) {
	var tasks []Task
	query := db.Preload("Tags")

	if completedFilter != nil {
		query = query.Where("completed = ?", *completedFilter)
	}

	if nameFilter != "" {
		query = query.Where("name LIKE ?", "%"+nameFilter+"%")
	}

	if len(tagIDFilter) > 0 {
		// Filter for tasks that have ALL specified tags
		query = query.Joins("JOIN task_tags tt ON tasks.id = tt.task_id").
			Where("tt.tag_id IN ?", tagIDFilter).
			Group("tasks.id").
			Having("COUNT(DISTINCT tt.tag_id) = ?", len(tagIDFilter))
	}

	switch sortField {
	case "completed":
		query = query.Order("completed ASC")
	case "priority":
		query = query.Order("completed ASC, priority ASC")
	case "name":
		query = query.Order("completed ASC, name")
	default:
		query = query.Order("completed ASC, date_created DESC")
	}

	result := query.Find(&tasks)
	return tasks, result.Error
}
