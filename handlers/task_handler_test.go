package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jhoffmann/dailies/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestHandlerDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto migrate tables
	err = db.AutoMigrate(&models.Task{}, &models.Tag{}, &models.Frequency{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestGetTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.GET("/tasks", GetTasks(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check response is valid JSON array
	var tasks []models.Task
	err := json.Unmarshal(w.Body.Bytes(), &tasks)
	if err != nil {
		t.Errorf("Expected valid JSON array, got error: %v", err)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.GET("/tasks/:id", GetTask(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks/non-existent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Task not found") {
		t.Errorf("Expected 'Task not found' error message")
	}
}

func TestCreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.POST("/tasks", CreateTask(db))

	requestBody := `{"name": "Test Task", "description": "A test task"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var task models.Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	if err != nil {
		t.Errorf("Expected valid JSON response, got error: %v", err)
	}

	if task.Name != "Test Task" {
		t.Errorf("Expected task name 'Test Task', got '%s'", task.Name)
	}

	if task.ID == "" {
		t.Error("Expected task ID to be generated")
	}
}

func TestCreateTaskValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.POST("/tasks", CreateTask(db))

	requestBody := `{"description": "Missing name field"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateTaskNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.PUT("/tasks/:id", UpdateTask(db))

	requestBody := `{"name": "Updated Task"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tasks/non-existent-id", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.DELETE("/tasks/:id", DeleteTask(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tasks/non-existent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
