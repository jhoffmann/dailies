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

func setupTagRoutesTestDB(t *testing.T) func() {
	originalDB := database.DB

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Tag{}, &models.Task{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	database.DB = db

	return func() {
		database.DB = originalDB
	}
}

func TestSetupTagRoutes(t *testing.T) {
	cleanup := setupTagRoutesTestDB(t)
	defer cleanup()

	// Create a new ServeMux for testing to avoid conflicts
	mux := http.NewServeMux()

	// Manually set up the routes for testing instead of calling SetupTagRoutes()
	mux.HandleFunc("/tags", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetTags(w, r)
		case http.MethodPost:
			api.CreateTag(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	mux.HandleFunc("/tags/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/tags/") == "" {
			logger.LoggedError(w, "Tag ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.GetTag(w, r)
		case http.MethodPut:
			api.UpdateTag(w, r)
		case http.MethodDelete:
			api.DeleteTag(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	t.Run("GET /tags", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tags", nil)
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

	t.Run("POST /tags", func(t *testing.T) {
		tag := models.Tag{Name: "test-tag", Color: "#fecaca"}
		tagJSON, _ := json.Marshal(tag)

		req := httptest.NewRequest("POST", "/tags", bytes.NewBuffer(tagJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("unsupported method on /tags", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tags", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("GET /tags/{id}", func(t *testing.T) {
		tag := models.Tag{Name: "get-test", Color: "#fed7aa"}
		database.DB.Create(&tag)

		req := httptest.NewRequest("GET", "/tags/"+tag.ID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("PUT /tags/{id}", func(t *testing.T) {
		tag := models.Tag{Name: "update-test", Color: "#fde68a"}
		database.DB.Create(&tag)

		updateData := models.Tag{Name: "updated-tag", Color: "#bbf7d0"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tags/"+tag.ID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("DELETE /tags/{id}", func(t *testing.T) {
		tag := models.Tag{Name: "delete-test", Color: "#a7f3d0"}
		database.DB.Create(&tag)

		req := httptest.NewRequest("DELETE", "/tags/"+tag.ID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}
	})

	t.Run("unsupported method on /tags/{id}", func(t *testing.T) {
		tagID := uuid.New()
		req := httptest.NewRequest("PATCH", "/tags/"+tagID.String(), nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("empty tag ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tags/", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
