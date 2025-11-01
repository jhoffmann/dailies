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

func TestUpdateTaskRemoveFrequency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	frequency := models.Frequency{
		Name:   "Daily",
		Period: "0 0 * * *",
	}
	db.Create(&frequency)

	task := models.Task{
		Name:        "Test Task",
		FrequencyID: &frequency.ID,
	}
	db.Create(&task)

	var createdTask models.Task
	db.Preload("Frequency").First(&createdTask, "id = ?", task.ID)
	if createdTask.FrequencyID == nil || *createdTask.FrequencyID != frequency.ID {
		t.Fatal("Task should have frequency assigned")
	}

	updateData := map[string]any{
		"frequency_id": "",
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/api/tasks/"+task.ID, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.PUT("/api/tasks/:id", UpdateTask(db))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var updatedTask models.Task
	db.Preload("Frequency").First(&updatedTask, "id = ?", task.ID)
	if updatedTask.FrequencyID != nil {
		t.Errorf("Expected frequency_id to be nil, got %v", *updatedTask.FrequencyID)
	}
	if updatedTask.Frequency != nil {
		t.Error("Expected frequency to be nil")
	}
}

func TestUpdateTaskRemovePriority(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	priority := 3
	task := models.Task{
		Name:     "Test Task",
		Priority: &priority,
	}
	db.Create(&task)

	var createdTask models.Task
	db.First(&createdTask, "id = ?", task.ID)
	if createdTask.Priority == nil || *createdTask.Priority != 3 {
		t.Fatal("Task should have priority assigned")
	}

	updateData := map[string]any{
		"priority": 0,
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/api/tasks/"+task.ID, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.PUT("/api/tasks/:id", UpdateTask(db))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var updatedTask models.Task
	db.First(&updatedTask, "id = ?", task.ID)
	if updatedTask.Priority != nil {
		t.Errorf("Expected priority to be nil, got %v", *updatedTask.Priority)
	}
}

func TestGetTasksFilterByTagNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	tag1 := models.Tag{Name: "warframe", Color: "#ff0000"}
	tag2 := models.Tag{Name: "games", Color: "#00ff00"}
	tag3 := models.Tag{Name: "work", Color: "#0000ff"}
	db.Create(&tag1)
	db.Create(&tag2)
	db.Create(&tag3)

	task1 := models.Task{Name: "Task 1"}
	task2 := models.Task{Name: "Task 2"}
	task3 := models.Task{Name: "Task 3"}
	db.Create(&task1)
	db.Create(&task2)
	db.Create(&task3)

	db.Model(&task1).Association("Tags").Append(&tag1)
	db.Model(&task2).Association("Tags").Append(&tag2)
	db.Model(&task3).Association("Tags").Append(&tag3)

	r := gin.New()
	r.GET("/tasks", GetTasks(db))

	// Test single tag name
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?tag=warframe", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var tasks []models.Task
	json.Unmarshal(w.Body.Bytes(), &tasks)
	if len(tasks) != 1 || tasks[0].Name != "Task 1" {
		t.Errorf("Expected 1 task with name 'Task 1', got %d tasks", len(tasks))
	}

	// Test multiple tag names
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/tasks?tag=warframe,games", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	json.Unmarshal(w.Body.Bytes(), &tasks)
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
}
