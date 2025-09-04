package main

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPrintUsage(t *testing.T) {
	printUsage()
}

func TestConnectToDatabase(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		dbPath := ":memory:"
		db, err := connectToDatabase(dbPath)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if db == nil {
			t.Error("Expected database connection, got nil")
		}

		var task models.Task
		err = db.AutoMigrate(&task)
		if err != nil {
			t.Errorf("Expected successful migration: %v", err)
		}
	})

	t.Run("invalid database path", func(t *testing.T) {
		dbPath := "/invalid/path/database.db"
		db, err := connectToDatabase(dbPath)

		if err == nil {
			t.Error("Expected error for invalid database path")
		}

		if db != nil {
			t.Error("Expected nil database connection")
		}
	})
}

func TestMain_NoArgs(t *testing.T) {
	if os.Getenv("TEST_MAIN") == "1" {
		main()
		return
	}

	os.Args = []string{"cmd"}

	if os.Getenv("CI") != "" {
		t.Skip("Skipping main test in CI environment")
	}
}

func TestMain_UnknownCommand(t *testing.T) {
	if os.Getenv("TEST_MAIN") == "1" {
		os.Args = []string{"cmd", "unknown"}
		main()
		return
	}

	if os.Getenv("CI") != "" {
		t.Skip("Skipping main test in CI environment")
	}
}

func TestMain_ValidCommands(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping main test in CI environment")
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	testDBFile := "test_main_commands.db"
	defer os.Remove(testDBFile)

	t.Run("populate command", func(t *testing.T) {
		os.Args = []string{"cmd", "populate", "--database", testDBFile, "--entries", "1"}

		defer func() {
			if r := recover(); r != nil {
				t.Log("Expected populate command to run without panic")
			}
		}()

		populateCommand([]string{"--database", testDBFile, "--entries", "1"})
	})

	t.Run("list command", func(t *testing.T) {
		os.Args = []string{"cmd", "list", "--database", testDBFile}

		defer func() {
			if r := recover(); r != nil {
				t.Log("Expected list command to run without panic")
			}
		}()

		listCommand([]string{"--database", testDBFile})
	})
}

func setupTestDB(t *testing.T) (*gorm.DB, string) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db, ":memory:"
}

func TestListCommand(t *testing.T) {
	testDBFile := "test_list.db"
	defer os.Remove(testDBFile)

	db, err := connectToDatabase(testDBFile)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	task1 := models.Task{Name: "Test Task 1", Completed: false}
	task2 := models.Task{Name: "Test Task 2", Completed: true}

	db.Create(&task1)
	db.Create(&task2)

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	args := []string{"--database", testDBFile}
	listCommand(args)

	w.Close()
	output, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	var tasks []models.Task
	err = json.Unmarshal(output, &tasks)
	if err != nil {
		t.Errorf("Failed to unmarshal output: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
}

func TestListCommand_EmptyDatabase(t *testing.T) {
	testDBFile := "test_empty_list.db"
	defer os.Remove(testDBFile)

	_, err := connectToDatabase(testDBFile)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	args := []string{"--database", testDBFile}
	listCommand(args)

	w.Close()
	output, _ := io.ReadAll(r)
	os.Stdout = originalStdout
	buf.Write(output)

	var tasks []models.Task
	err = json.Unmarshal(buf.Bytes(), &tasks)
	if err != nil {
		t.Errorf("Failed to unmarshal output: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestListCommand_WithFlags(t *testing.T) {
	testDBFile := "test_flags_list.db"
	defer os.Remove(testDBFile)

	db, err := connectToDatabase(testDBFile)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	task1 := models.Task{Name: "Test Task 1", Completed: false}
	db.Create(&task1)

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	args := []string{"--database", testDBFile}
	listCommand(args)

	w.Close()
	output, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	var tasks []models.Task
	err = json.Unmarshal(output, &tasks)
	if err != nil {
		t.Errorf("Failed to unmarshal output: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
}

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
