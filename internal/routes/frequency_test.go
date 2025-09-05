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

func setupFrequencyRoutesTestDB(t *testing.T) func() {
	originalDB := database.DB

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Frequency{}, &models.Task{}, &models.Tag{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	database.DB = db

	return func() {
		database.DB = originalDB
	}
}

func TestSetupFrequencyRoutes(t *testing.T) {
	cleanup := setupFrequencyRoutesTestDB(t)
	defer cleanup()

	// Create a new ServeMux for testing to avoid conflicts
	mux := http.NewServeMux()

	// Manually set up the frequency routes for testing
	mux.HandleFunc("/frequencies", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetFrequencies(w, r)
		case http.MethodPost:
			api.CreateFrequency(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	mux.HandleFunc("/frequencies/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/frequencies/") == "" {
			logger.LoggedError(w, "Frequency ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.GetFrequency(w, r)
		case http.MethodPut:
			api.UpdateFrequency(w, r)
		case http.MethodDelete:
			api.DeleteFrequency(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	t.Run("GET /frequencies", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies", nil)
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

	t.Run("POST /frequencies", func(t *testing.T) {
		requestBody := map[string]string{
			"name":  "Test Frequency",
			"reset": "0 18 * * *",
		}
		frequencyJSON, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/frequencies", bytes.NewBuffer(frequencyJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		var frequency models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &frequency)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if frequency.Name != "Test Frequency" {
			t.Errorf("Expected name 'Test Frequency', got %s", frequency.Name)
		}
	})

	t.Run("GET /frequencies/{id}", func(t *testing.T) {
		// Create a frequency first
		frequency := models.Frequency{Name: "Get Test", Reset: "0 6 * * *"}
		database.DB.Create(&frequency)

		req := httptest.NewRequest("GET", "/frequencies/"+frequency.ID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var responseFreq models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &responseFreq)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if responseFreq.Name != "Get Test" {
			t.Errorf("Expected name 'Get Test', got %s", responseFreq.Name)
		}
	})

	t.Run("PUT /frequencies/{id}", func(t *testing.T) {
		// Create a frequency first
		frequency := models.Frequency{Name: "Update Test", Reset: "0 18 * * *"}
		database.DB.Create(&frequency)

		updateBody := map[string]string{
			"name": "Updated Name",
		}
		updateJSON, _ := json.Marshal(updateBody)

		req := httptest.NewRequest("PUT", "/frequencies/"+frequency.ID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var responseFreq models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &responseFreq)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if responseFreq.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %s", responseFreq.Name)
		}
	})

	t.Run("DELETE /frequencies/{id}", func(t *testing.T) {
		// Create a frequency first
		frequency := models.Frequency{Name: "Delete Test", Reset: "0 23 * * 1"}
		database.DB.Create(&frequency)

		req := httptest.NewRequest("DELETE", "/frequencies/"+frequency.ID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		// Verify frequency was deleted
		var dbFreq models.Frequency
		result := database.DB.First(&dbFreq, "id = ?", frequency.ID)
		if result.Error == nil {
			t.Error("Expected frequency to be deleted")
		}
	})

	t.Run("Method not allowed on /frequencies", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/frequencies", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("Method not allowed on /frequencies/{id}", func(t *testing.T) {
		frequency := models.Frequency{Name: "Method Test", Reset: "0 18 * * *"}
		database.DB.Create(&frequency)

		req := httptest.NewRequest("PATCH", "/frequencies/"+frequency.ID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("Missing frequency ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies/", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Invalid frequency ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies/invalid-id", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Non-existent frequency", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("GET", "/frequencies/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Invalid JSON on POST", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/frequencies", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Invalid JSON on PUT", func(t *testing.T) {
		frequency := models.Frequency{Name: "JSON Test", Reset: "0 18 * * *"}
		database.DB.Create(&frequency)

		req := httptest.NewRequest("PUT", "/frequencies/"+frequency.ID.String(), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
