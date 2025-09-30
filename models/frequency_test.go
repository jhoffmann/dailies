package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestFrequencyBeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	frequency := &Frequency{
		Name:   "Daily",
		Period: "daily",
	}

	err := frequency.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if frequency.ID == "" {
		t.Error("Expected frequency ID to be generated, got empty string")
	}

	// Validate UUID format
	_, err = uuid.Parse(frequency.ID)
	if err != nil {
		t.Errorf("Expected valid UUID, got invalid format: %v", err)
	}
}

func TestFrequencyBeforeCreateWithExistingID(t *testing.T) {
	db := setupTestDB(t)

	existingID := uuid.New().String()
	frequency := &Frequency{
		ID:     existingID,
		Name:   "Weekly",
		Period: "weekly",
	}

	err := frequency.BeforeCreate(db)
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if frequency.ID != existingID {
		t.Errorf("Expected frequency ID to remain %s, got %s", existingID, frequency.ID)
	}
}

func TestFrequencyCreation(t *testing.T) {
	db := setupTestDB(t)

	frequency := &Frequency{
		Name:   "Daily",
		Period: "daily",
	}

	result := db.Create(frequency)
	if result.Error != nil {
		t.Fatalf("Failed to create frequency: %v", result.Error)
	}

	if frequency.ID == "" {
		t.Error("Expected frequency ID to be generated")
	}

	var retrievedFrequency Frequency
	result = db.First(&retrievedFrequency, "id = ?", frequency.ID)
	if result.Error != nil {
		t.Fatalf("Failed to retrieve frequency: %v", result.Error)
	}

	if retrievedFrequency.Name != "Daily" {
		t.Errorf("Expected name 'Daily', got %s", retrievedFrequency.Name)
	}

	if retrievedFrequency.Period != "daily" {
		t.Errorf("Expected period 'daily', got %s", retrievedFrequency.Period)
	}
}
