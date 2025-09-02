package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTaskStruct(t *testing.T) {
	task := Task{
		Name:      "Test Task",
		Completed: true,
	}

	if task.Name != "Test Task" {
		t.Errorf("Expected name 'Test Task', got %s", task.Name)
	}

	if !task.Completed {
		t.Errorf("Expected completed to be true")
	}
}

func TestBeforeCreate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&Task{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	t.Run("generates UUID when ID is nil", func(t *testing.T) {
		task := Task{Name: "Test Task"}

		if task.ID != uuid.Nil {
			t.Errorf("Expected ID to be nil initially")
		}

		err := task.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if task.ID == uuid.Nil {
			t.Errorf("Expected ID to be generated")
		}
	})

	t.Run("preserves existing UUID", func(t *testing.T) {
		existingID := uuid.New()
		task := Task{
			ID:   existingID,
			Name: "Test Task",
		}

		err := task.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if task.ID != existingID {
			t.Errorf("Expected ID to remain %s, got %s", existingID, task.ID)
		}
	})
}

func TestTaskJSONTags(t *testing.T) {
	task := Task{
		ID:           uuid.New(),
		Name:         "Test Task",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Completed:    true,
	}

	if task.ID == uuid.Nil {
		t.Error("Expected task to have an ID")
	}

	if task.Name == "" {
		t.Error("Expected task to have a name")
	}

	if task.DateCreated.IsZero() {
		t.Error("Expected task to have a creation date")
	}

	if task.DateModified.IsZero() {
		t.Error("Expected task to have a modification date")
	}
}
