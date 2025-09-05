package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTaskRoutesTestDB(t *testing.T) func() {
	originalDB := database.DB

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Task{}, &models.Tag{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	database.DB = db

	return func() {
		database.DB = originalDB
	}
}

func TestSetupTaskRoutes(t *testing.T) {
	cleanup := setupTaskRoutesTestDB(t)
	defer cleanup()

	// Create a new ServeMux for testing to avoid conflicts
	mux := http.NewServeMux()

	// Manually set up the task routes for testing
	mux.HandleFunc("/tasks", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetTasks(w, r)
		case http.MethodPost:
			api.CreateTask(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	mux.HandleFunc("/tasks/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/tasks/") == "" {
			logger.LoggedError(w, "Task ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.GetTask(w, r)
		case http.MethodPut:
			api.UpdateTask(w, r)
		case http.MethodDelete:
			api.DeleteTask(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	t.Run("GET /tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
	})

	t.Run("POST /tasks", func(t *testing.T) {
		task := models.Task{Name: "Test Task"}
		taskJSON, _ := json.Marshal(task)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("unsupported method on /tasks", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("GET /tasks/{id}", func(t *testing.T) {
		task := models.Task{Name: "Test Task"}
		database.DB.Create(&task)

		req := httptest.NewRequest("GET", "/tasks/"+task.ID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("PUT /tasks/{id}", func(t *testing.T) {
		task := models.Task{Name: "Original Task"}
		database.DB.Create(&task)

		updateData := models.Task{Name: "Updated Task", Completed: true}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tasks/"+task.ID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("DELETE /tasks/{id}", func(t *testing.T) {
		task := models.Task{Name: "Task to Delete"}
		database.DB.Create(&task)

		req := httptest.NewRequest("DELETE", "/tasks/"+task.ID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}
	})

	t.Run("unsupported method on /tasks/{id}", func(t *testing.T) {
		taskID := uuid.New()
		req := httptest.NewRequest("PATCH", "/tasks/"+taskID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("empty task ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("GET /tasks with query parameters", func(t *testing.T) {
		// Create some test data
		task1 := models.Task{Name: "Complete Task", Completed: true}
		task2 := models.Task{Name: "Incomplete Task", Completed: false}
		database.DB.Create(&task1)
		database.DB.Create(&task2)

		req := httptest.NewRequest("GET", "/tasks?completed=true", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tasks []models.Task
		err := json.Unmarshal(w.Body.Bytes(), &tasks)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Should have at least the completed task
		foundCompleted := false
		for _, task := range tasks {
			if task.Completed {
				foundCompleted = true
				break
			}
		}
		if !foundCompleted {
			t.Error("Expected to find at least one completed task")
		}
	})

	t.Run("GET /tasks with name filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?name=Complete", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /tasks with sort parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?sort=priority", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}
