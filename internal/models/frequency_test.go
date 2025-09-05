package models

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFrequencyStruct(t *testing.T) {
	frequency := Frequency{
		Name:  "Daily Evening",
		Reset: "0 18 * * *",
	}

	if frequency.Name != "Daily Evening" {
		t.Errorf("Expected name 'Daily Evening', got %s", frequency.Name)
	}

	if frequency.Reset != "0 18 * * *" {
		t.Errorf("Expected reset '0 18 * * *', got %s", frequency.Reset)
	}
}

func setupFrequencyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&Frequency{}, &Task{}, &Tag{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestFrequencyBeforeCreate(t *testing.T) {
	db := setupFrequencyTestDB(t)

	t.Run("generates UUID when ID is nil", func(t *testing.T) {
		frequency := Frequency{Name: "Test Frequency", Reset: "0 18 * * *"}

		if frequency.ID != uuid.Nil {
			t.Errorf("Expected ID to be nil initially")
		}

		err := frequency.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if frequency.ID == uuid.Nil {
			t.Errorf("Expected ID to be generated")
		}
	})

	t.Run("preserves existing UUID", func(t *testing.T) {
		existingID := uuid.New()
		frequency := Frequency{
			ID:    existingID,
			Name:  "Test Frequency",
			Reset: "0 18 * * *",
		}

		err := frequency.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if frequency.ID != existingID {
			t.Errorf("Expected ID to remain %s, got %s", existingID, frequency.ID)
		}
	})
}

func TestFrequencyCreate(t *testing.T) {
	db := setupFrequencyTestDB(t)

	t.Run("creates frequency with name and reset", func(t *testing.T) {
		frequency := Frequency{Name: "Daily Morning", Reset: "0 6 * * *"}

		err := frequency.Create(db)
		if err != nil {
			t.Errorf("Create returned error: %v", err)
		}

		if frequency.ID == uuid.Nil {
			t.Error("Expected ID to be generated")
		}

		var dbFrequency Frequency
		result := db.First(&dbFrequency, "id = ?", frequency.ID)
		if result.Error != nil {
			t.Errorf("Frequency not found in database: %v", result.Error)
		}

		if dbFrequency.Name != "Daily Morning" {
			t.Errorf("Expected name 'Daily Morning', got %s", dbFrequency.Name)
		}

		if dbFrequency.Reset != "0 6 * * *" {
			t.Errorf("Expected reset '0 6 * * *', got %s", dbFrequency.Reset)
		}
	})

	t.Run("fails to create frequency without name", func(t *testing.T) {
		frequency := Frequency{Name: "", Reset: "0 18 * * *"}

		err := frequency.Create(db)
		if err == nil {
			t.Error("Expected error when creating frequency without name")
		}

		if err.Error() != "frequency name is required" {
			t.Errorf("Expected 'frequency name is required', got %s", err.Error())
		}
	})

	t.Run("fails to create frequency without reset", func(t *testing.T) {
		frequency := Frequency{Name: "Daily", Reset: ""}

		err := frequency.Create(db)
		if err == nil {
			t.Error("Expected error when creating frequency without reset")
		}

		if err.Error() != "frequency reset is required" {
			t.Errorf("Expected 'frequency reset is required', got %s", err.Error())
		}
	})

	t.Run("enforces unique name constraint", func(t *testing.T) {
		frequency1 := Frequency{Name: "Daily", Reset: "0 18 * * *"}
		err := frequency1.Create(db)
		if err != nil {
			t.Fatalf("Failed to create first frequency: %v", err)
		}

		frequency2 := Frequency{Name: "Daily", Reset: "0 6 * * *"}
		err = frequency2.Create(db)
		if err == nil {
			t.Error("Expected error when creating frequency with duplicate name")
		}
	})
}

func TestFrequencySave(t *testing.T) {
	db := setupFrequencyTestDB(t)

	frequency := Frequency{Name: "Weekly", Reset: "0 23 * * 1"}
	err := frequency.Create(db)
	if err != nil {
		t.Fatalf("Failed to create initial frequency: %v", err)
	}

	frequency.Reset = "0 23 * * 0" // Change to Sunday
	err = frequency.Save(db)
	if err != nil {
		t.Errorf("Save returned error: %v", err)
	}

	var dbFrequency Frequency
	result := db.First(&dbFrequency, "id = ?", frequency.ID)
	if result.Error != nil {
		t.Errorf("Frequency not found: %v", result.Error)
	}

	if dbFrequency.Reset != "0 23 * * 0" {
		t.Errorf("Expected reset '0 23 * * 0', got %s", dbFrequency.Reset)
	}
}

func TestFrequencyLoadByID(t *testing.T) {
	db := setupFrequencyTestDB(t)

	originalFrequency := Frequency{Name: "Monthly", Reset: "0 0 15 * *"}
	err := originalFrequency.Create(db)
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	t.Run("loads existing frequency", func(t *testing.T) {
		var loadedFrequency Frequency
		err := loadedFrequency.LoadByID(db, originalFrequency.ID)
		if err != nil {
			t.Errorf("LoadByID returned error: %v", err)
		}

		if loadedFrequency.Name != "Monthly" {
			t.Errorf("Expected name 'Monthly', got %s", loadedFrequency.Name)
		}

		if loadedFrequency.Reset != "0 0 15 * *" {
			t.Errorf("Expected reset '0 0 15 * *', got %s", loadedFrequency.Reset)
		}
	})

	t.Run("fails to load non-existent frequency", func(t *testing.T) {
		var loadedFrequency Frequency
		nonExistentID := uuid.New()
		err := loadedFrequency.LoadByID(db, nonExistentID)
		if err == nil {
			t.Error("Expected error when loading non-existent frequency")
		}

		if err.Error() != "frequency not found" {
			t.Errorf("Expected 'frequency not found', got %s", err.Error())
		}
	})
}

func TestFrequencyUpdate(t *testing.T) {
	db := setupFrequencyTestDB(t)

	frequency := Frequency{Name: "Original", Reset: "0 18 * * *"}
	err := frequency.Create(db)
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	t.Run("updates name", func(t *testing.T) {
		updateData := Frequency{Name: "Updated Name"}
		err := frequency.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if frequency.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %s", frequency.Name)
		}
	})

	t.Run("updates reset", func(t *testing.T) {
		updateData := Frequency{Reset: "0 6 * * *"}
		err := frequency.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if frequency.Reset != "0 6 * * *" {
			t.Errorf("Expected reset '0 6 * * *', got %s", frequency.Reset)
		}
	})

	t.Run("updates both name and reset", func(t *testing.T) {
		updateData := Frequency{Name: "Final Name", Reset: "0 23 * * 1"}
		err := frequency.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if frequency.Name != "Final Name" {
			t.Errorf("Expected name 'Final Name', got %s", frequency.Name)
		}

		if frequency.Reset != "0 23 * * 1" {
			t.Errorf("Expected reset '0 23 * * 1', got %s", frequency.Reset)
		}
	})

	t.Run("ignores empty values", func(t *testing.T) {
		originalName := frequency.Name
		originalReset := frequency.Reset
		updateData := Frequency{Name: "", Reset: ""}
		err := frequency.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if frequency.Name != originalName {
			t.Errorf("Expected name to remain %s, got %s", originalName, frequency.Name)
		}

		if frequency.Reset != originalReset {
			t.Errorf("Expected reset to remain %s, got %s", originalReset, frequency.Reset)
		}
	})
}

func TestFrequencyDelete(t *testing.T) {
	db := setupFrequencyTestDB(t)

	t.Run("deletes existing frequency", func(t *testing.T) {
		frequency := Frequency{Name: "Delete Me", Reset: "0 18 * * *"}
		err := frequency.Create(db)
		if err != nil {
			t.Fatalf("Failed to create frequency: %v", err)
		}

		err = frequency.Delete(db)
		if err != nil {
			t.Errorf("Delete returned error: %v", err)
		}

		var dbFrequency Frequency
		result := db.First(&dbFrequency, "id = ?", frequency.ID)
		if result.Error == nil {
			t.Error("Expected frequency to be deleted")
		}
	})

	t.Run("fails to delete non-existent frequency", func(t *testing.T) {
		frequency := Frequency{ID: uuid.New()}
		err := frequency.Delete(db)
		if err == nil {
			t.Error("Expected error when deleting non-existent frequency")
		}

		if err.Error() != "frequency not found" {
			t.Errorf("Expected 'frequency not found', got %s", err.Error())
		}
	})
}

func TestGetFrequencies(t *testing.T) {
	db := setupFrequencyTestDB(t)

	freq1 := Frequency{Name: "Daily", Reset: "0 18 * * *"}
	freq2 := Frequency{Name: "Weekly", Reset: "0 23 * * 1"}
	freq3 := Frequency{Name: "Monthly", Reset: "0 0 15 * *"}

	db.Create(&freq1)
	db.Create(&freq2)
	db.Create(&freq3)

	t.Run("gets all frequencies", func(t *testing.T) {
		frequencies, err := GetFrequencies(db, "")
		if err != nil {
			t.Errorf("GetFrequencies returned error: %v", err)
		}

		if len(frequencies) != 3 {
			t.Errorf("Expected 3 frequencies, got %d", len(frequencies))
		}
	})

	t.Run("filters by name", func(t *testing.T) {
		frequencies, err := GetFrequencies(db, "Week")
		if err != nil {
			t.Errorf("GetFrequencies returned error: %v", err)
		}

		if len(frequencies) != 1 {
			t.Errorf("Expected 1 frequency with 'Week' in name, got %d", len(frequencies))
		}

		if frequencies[0].Name != "Weekly" {
			t.Errorf("Expected name 'Weekly', got %s", frequencies[0].Name)
		}
	})

	t.Run("sorts by name", func(t *testing.T) {
		frequencies, err := GetFrequencies(db, "")
		if err != nil {
			t.Errorf("GetFrequencies returned error: %v", err)
		}

		if len(frequencies) < 2 {
			t.Error("Expected at least 2 frequencies for sorting test")
		}

		// Check if sorted alphabetically
		for i := 1; i < len(frequencies); i++ {
			if frequencies[i-1].Name > frequencies[i].Name {
				t.Errorf("Frequencies not sorted by name: %s > %s", frequencies[i-1].Name, frequencies[i].Name)
			}
		}
	})

	t.Run("returns empty for non-matching filter", func(t *testing.T) {
		frequencies, err := GetFrequencies(db, "NonExistent")
		if err != nil {
			t.Errorf("GetFrequencies returned error: %v", err)
		}

		if len(frequencies) != 0 {
			t.Errorf("Expected 0 frequencies, got %d", len(frequencies))
		}
	})
}

func TestFrequencyTaskRelationship(t *testing.T) {
	db := setupFrequencyTestDB(t)

	frequency := Frequency{Name: "Daily", Reset: "0 18 * * *"}
	err := frequency.Create(db)
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	t.Run("task can be associated with frequency", func(t *testing.T) {
		task := Task{Name: "Test Task", FrequencyID: &frequency.ID}
		err := task.Create(db)
		if err != nil {
			t.Errorf("Failed to create task with frequency: %v", err)
		}

		var loadedTask Task
		err = loadedTask.LoadByID(db, task.ID)
		if err != nil {
			t.Errorf("Failed to load task: %v", err)
		}

		if loadedTask.FrequencyID == nil {
			t.Error("Expected task to have frequency ID")
		}

		if *loadedTask.FrequencyID != frequency.ID {
			t.Errorf("Expected frequency ID %s, got %s", frequency.ID, *loadedTask.FrequencyID)
		}

		if loadedTask.Frequency == nil {
			t.Error("Expected frequency to be preloaded")
		}

		if loadedTask.Frequency.Name != "Daily" {
			t.Errorf("Expected frequency name 'Daily', got %s", loadedTask.Frequency.Name)
		}
	})

	t.Run("task can exist without frequency", func(t *testing.T) {
		task := Task{Name: "Task without frequency"}
		err := task.Create(db)
		if err != nil {
			t.Errorf("Failed to create task without frequency: %v", err)
		}

		var loadedTask Task
		err = loadedTask.LoadByID(db, task.ID)
		if err != nil {
			t.Errorf("Failed to load task: %v", err)
		}

		if loadedTask.FrequencyID != nil {
			t.Error("Expected task to have no frequency ID")
		}

		if loadedTask.Frequency != nil {
			t.Error("Expected frequency to be nil")
		}
	})

	t.Run("frequency can have multiple tasks", func(t *testing.T) {
		task1 := Task{Name: "Task 1", FrequencyID: &frequency.ID}
		task2 := Task{Name: "Task 2", FrequencyID: &frequency.ID}

		err := task1.Create(db)
		if err != nil {
			t.Errorf("Failed to create task1: %v", err)
		}

		err = task2.Create(db)
		if err != nil {
			t.Errorf("Failed to create task2: %v", err)
		}

		var loadedFrequency Frequency
		err = db.Preload("Tasks").First(&loadedFrequency, "id = ?", frequency.ID).Error
		if err != nil {
			t.Errorf("Failed to load frequency with tasks: %v", err)
		}

		if len(loadedFrequency.Tasks) < 2 {
			t.Errorf("Expected at least 2 tasks for frequency, got %d", len(loadedFrequency.Tasks))
		}
	})
}

func TestFrequencyJSONTags(t *testing.T) {
	frequency := Frequency{
		ID:    uuid.New(),
		Name:  "Test Frequency",
		Reset: "0 18 * * *",
	}

	if frequency.ID == uuid.Nil {
		t.Error("Expected frequency to have an ID")
	}

	if frequency.Name == "" {
		t.Error("Expected frequency to have a name")
	}

	if frequency.Reset == "" {
		t.Error("Expected frequency to have a reset schedule")
	}
}
