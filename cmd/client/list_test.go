package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
