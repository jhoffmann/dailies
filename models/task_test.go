package models

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&Task{}, &Frequency{}, &Tag{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestTaskBeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	task := &Task{
		Name: "Test Task",
	}

	err := task.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if task.ID == "" {
		t.Error("Expected task ID to be generated, got empty string")
	}

	// Validate UUID format
	_, err = uuid.Parse(task.ID)
	if err != nil {
		t.Errorf("Expected valid UUID, got invalid format: %v", err)
	}
}

func TestTaskBeforeCreateWithExistingID(t *testing.T) {
	db := setupTestDB(t)

	existingID := uuid.New().String()
	task := &Task{
		ID:   existingID,
		Name: "Test Task",
	}

	err := task.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if task.ID != existingID {
		t.Errorf("Expected task ID to remain %s, got %s", existingID, task.ID)
	}
}

func TestTaskCreation(t *testing.T) {
	db := setupTestDB(t)

	task := &Task{
		Name:        "Test Task",
		Description: stringPtr("Test description"),
		Completed:   false,
		Priority:    intPtr(3),
	}

	result := db.Create(task)
	if result.Error != nil {
		t.Fatalf("Failed to create task: %v", result.Error)
	}

	if task.ID == "" {
		t.Error("Expected task ID to be generated")
	}

	var retrievedTask Task
	result = db.First(&retrievedTask, "id = ?", task.ID)
	if result.Error != nil {
		t.Fatalf("Failed to retrieve task: %v", result.Error)
	}

	if retrievedTask.Name != "Test Task" {
		t.Errorf("Expected name 'Test Task', got %s", retrievedTask.Name)
	}

	if retrievedTask.Description == nil || *retrievedTask.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %v", retrievedTask.Description)
	}

	if retrievedTask.Priority == nil || *retrievedTask.Priority != 3 {
		t.Errorf("Expected priority 3, got %v", retrievedTask.Priority)
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
