package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) func() {
	originalDB := database.DB

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	database.DB = db

	return func() {
		database.DB = originalDB
	}
}

func TestGetTasks(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	task1 := models.Task{Name: "Test Task 1", Completed: false}
	task2 := models.Task{Name: "Test Task 2", Completed: true}
	task3 := models.Task{Name: "Another Task", Completed: false}

	database.DB.Create(&task1)
	database.DB.Create(&task2)
	database.DB.Create(&task3)

	t.Run("get all tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		w := httptest.NewRecorder()

		GetTasks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tasks []models.Task
		err := json.Unmarshal(w.Body.Bytes(), &tasks)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(tasks))
		}
	})

	t.Run("filter by completed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?completed=true", nil)
		w := httptest.NewRecorder()

		GetTasks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tasks []models.Task
		err := json.Unmarshal(w.Body.Bytes(), &tasks)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tasks) != 1 {
			t.Errorf("Expected 1 completed task, got %d", len(tasks))
		}

		if !tasks[0].Completed {
			t.Error("Expected task to be completed")
		}
	})

	t.Run("filter by completed false", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?completed=false", nil)
		w := httptest.NewRecorder()

		GetTasks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tasks []models.Task
		err := json.Unmarshal(w.Body.Bytes(), &tasks)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tasks) != 2 {
			t.Errorf("Expected 2 incomplete tasks, got %d", len(tasks))
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?name=Test", nil)
		w := httptest.NewRecorder()

		GetTasks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tasks []models.Task
		err := json.Unmarshal(w.Body.Bytes(), &tasks)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tasks) != 2 {
			t.Errorf("Expected 2 tasks with 'Test' in name, got %d", len(tasks))
		}
	})

	t.Run("invalid completed parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?completed=invalid", nil)
		w := httptest.NewRecorder()

		GetTasks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tasks []models.Task
		err := json.Unmarshal(w.Body.Bytes(), &tasks)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tasks) != 3 {
			t.Errorf("Expected all 3 tasks when completed param is invalid, got %d", len(tasks))
		}
	})
}

func TestGetTask(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	task := models.Task{Name: "Test Task", Completed: false}
	database.DB.Create(&task)

	t.Run("get existing task", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/"+task.ID.String(), nil)
		w := httptest.NewRecorder()

		GetTask(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var returnedTask models.Task
		err := json.Unmarshal(w.Body.Bytes(), &returnedTask)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if returnedTask.ID != task.ID {
			t.Errorf("Expected task ID %s, got %s", task.ID, returnedTask.ID)
		}
	})

	t.Run("invalid task ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/invalid-uuid", nil)
		w := httptest.NewRecorder()

		GetTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("GET", "/tasks/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		GetTask(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestCreateTask(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("create valid task", func(t *testing.T) {
		task := models.Task{Name: "New Test Task"}
		taskJSON, _ := json.Marshal(task)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTask(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var createdTask models.Task
		err := json.Unmarshal(w.Body.Bytes(), &createdTask)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if createdTask.Name != "New Test Task" {
			t.Errorf("Expected task name 'New Test Task', got %s", createdTask.Name)
		}

		if createdTask.ID == uuid.Nil {
			t.Error("Expected task to have generated ID")
		}
	})

	t.Run("create task without name", func(t *testing.T) {
		task := models.Task{Name: ""}
		taskJSON, _ := json.Marshal(task)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestUpdateTask(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	task := models.Task{Name: "Original Task", Completed: false}
	database.DB.Create(&task)

	t.Run("update existing task", func(t *testing.T) {
		updateData := models.Task{Name: "Updated Task", Completed: true}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tasks/"+task.ID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTask(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var updatedTask models.Task
		err := json.Unmarshal(w.Body.Bytes(), &updatedTask)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if updatedTask.Name != "Updated Task" {
			t.Errorf("Expected updated name 'Updated Task', got %s", updatedTask.Name)
		}

		if !updatedTask.Completed {
			t.Error("Expected task to be marked as completed")
		}
	})

	t.Run("update only name", func(t *testing.T) {
		task2 := models.Task{Name: "Task 2", Completed: true}
		database.DB.Create(&task2)

		updateData := models.Task{Name: "Updated Name Only"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tasks/"+task2.ID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTask(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var updatedTask models.Task
		err := json.Unmarshal(w.Body.Bytes(), &updatedTask)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if updatedTask.Name != "Updated Name Only" {
			t.Errorf("Expected updated name, got %s", updatedTask.Name)
		}

		if updatedTask.Completed {
			t.Error("Expected completed status to be reset to false")
		}
	})

	t.Run("invalid task ID", func(t *testing.T) {
		updateData := models.Task{Name: "Updated Task"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tasks/invalid-uuid", bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		updateData := models.Task{Name: "Updated Task"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tasks/"+nonExistentID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTask(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/tasks/"+task.ID.String(), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestDeleteTask(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	task := models.Task{Name: "Task to Delete", Completed: false}
	database.DB.Create(&task)

	t.Run("delete existing task", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/"+task.ID.String(), nil)
		w := httptest.NewRecorder()

		DeleteTask(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		var deletedTask models.Task
		result := database.DB.First(&deletedTask, "id = ?", task.ID)
		if result.Error == nil {
			t.Error("Expected task to be deleted")
		}
	})

	t.Run("invalid task ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/invalid-uuid", nil)
		w := httptest.NewRecorder()

		DeleteTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("DELETE", "/tasks/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		DeleteTask(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}
