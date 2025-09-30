package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestTagBeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	tag := &Tag{
		Name:  "Work",
		Color: "#FF0000",
	}

	err := tag.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if tag.ID == "" {
		t.Error("Expected tag ID to be generated, got empty string")
	}

	// Validate UUID format
	_, err = uuid.Parse(tag.ID)
	if err != nil {
		t.Errorf("Expected valid UUID, got invalid format: %v", err)
	}
}

func TestTagBeforeCreateWithExistingID(t *testing.T) {
	db := setupTestDB(t)

	existingID := uuid.New().String()
	tag := &Tag{
		ID:    existingID,
		Name:  "Personal",
		Color: "#00FF00",
	}

	err := tag.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if tag.ID != existingID {
		t.Errorf("Expected tag ID to remain %s, got %s", existingID, tag.ID)
	}
}

func TestTagCreation(t *testing.T) {
	db := setupTestDB(t)

	tag := &Tag{
		Name:  "Work",
		Color: "#FF0000",
	}

	result := db.Create(tag)
	if result.Error != nil {
		t.Fatalf("Failed to create tag: %v", result.Error)
	}

	if tag.ID == "" {
		t.Error("Expected tag ID to be generated")
	}

	var retrievedTag Tag
	result = db.First(&retrievedTag, "id = ?", tag.ID)
	if result.Error != nil {
		t.Fatalf("Failed to retrieve tag: %v", result.Error)
	}

	if retrievedTag.Name != "Work" {
		t.Errorf("Expected name 'Work', got %s", retrievedTag.Name)
	}

	if retrievedTag.Color != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got %s", retrievedTag.Color)
	}
}
