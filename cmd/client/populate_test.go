package main

import (
	"math/rand"
	"os"
	"testing"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForPopulate(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestPopulateWithSampleData(t *testing.T) {
	db := setupTestDBForPopulate(t)

	t.Run("populate with positive count", func(t *testing.T) {
		count := 5
		err := populateWithSampleData(db, count)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		var tasks []models.Task
		result := db.Find(&tasks)
		if result.Error != nil {
			t.Errorf("Failed to query tasks: %v", result.Error)
		}

		if len(tasks) != count {
			t.Errorf("Expected %d tasks, got %d", count, len(tasks))
		}
	})

	t.Run("populate with zero count", func(t *testing.T) {
		db2 := setupTestDBForPopulate(t)
		count := 0
		err := populateWithSampleData(db2, count)

		if err != nil {
			t.Errorf("Expected no error for zero count, got %v", err)
		}

		var tasks []models.Task
		result := db2.Find(&tasks)
		if result.Error != nil {
			t.Errorf("Failed to query tasks: %v", result.Error)
		}

		if len(tasks) != 0 {
			t.Errorf("Expected 0 tasks, got %d", len(tasks))
		}
	})
}

func TestGetRandomTaskName(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	name := getRandomTaskName(r)
	if name == "" {
		t.Error("Expected non-empty task name")
	}

	found := false
	for _, taskName := range taskNames {
		if name == taskName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected task name to be from predefined list, got %s", name)
	}
}

func TestTaskNames(t *testing.T) {
	if len(taskNames) == 0 {
		t.Error("Expected taskNames to contain at least one task")
	}

	for i, taskName := range taskNames {
		if taskName == "" {
			t.Errorf("Expected taskNames[%d] to be non-empty", i)
		}
	}
}

func TestPopulateCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping populate command test in short mode")
	}

	testDBFile := "test_populate.db"
	defer os.Remove(testDBFile)

	args := []string{"--database", testDBFile, "--entries", "3"}
	populateCommand(args)

	db, err := connectToDatabase(testDBFile)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	var tasks []models.Task
	result := db.Find(&tasks)
	if result.Error != nil {
		t.Errorf("Failed to query tasks: %v", result.Error)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks after populate, got %d", len(tasks))
	}
}
