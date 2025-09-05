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

func setupTagTestDB(t *testing.T) func() {
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

func TestGetTags(t *testing.T) {
	cleanup := setupTagTestDB(t)
	defer cleanup()

	tag1 := models.Tag{Name: "work", Color: "#fecaca"}
	tag2 := models.Tag{Name: "personal", Color: "#fed7aa"}
	tag3 := models.Tag{Name: "urgent", Color: "#fde68a"}

	database.DB.Create(&tag1)
	database.DB.Create(&tag2)
	database.DB.Create(&tag3)

	t.Run("get all tags", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tags", nil)
		w := httptest.NewRecorder()

		GetTags(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tags []models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &tags)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tags) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(tags))
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tags?name=work", nil)
		w := httptest.NewRecorder()

		GetTags(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tags []models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &tags)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tags) != 1 {
			t.Errorf("Expected 1 tag with 'work' in name, got %d", len(tags))
		}

		if tags[0].Name != "work" {
			t.Errorf("Expected tag name 'work', got %s", tags[0].Name)
		}
	})

	t.Run("filter by partial name", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tags?name=p", nil)
		w := httptest.NewRecorder()

		GetTags(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tags []models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &tags)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(tags) != 1 {
			t.Errorf("Expected 1 tag with 'p' in name, got %d", len(tags))
		}
	})
}

func TestGetTag(t *testing.T) {
	cleanup := setupTagTestDB(t)
	defer cleanup()

	tag := models.Tag{Name: "test", Color: "#fecaca"}
	database.DB.Create(&tag)

	t.Run("get existing tag", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tags/"+tag.ID.String(), nil)
		w := httptest.NewRecorder()

		GetTag(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var returnedTag models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &returnedTag)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if returnedTag.ID != tag.ID {
			t.Errorf("Expected tag ID %s, got %s", tag.ID, returnedTag.ID)
		}

		if returnedTag.Name != "test" {
			t.Errorf("Expected tag name 'test', got %s", returnedTag.Name)
		}
	})

	t.Run("invalid tag ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tags/invalid-uuid", nil)
		w := httptest.NewRecorder()

		GetTag(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("tag not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("GET", "/tags/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		GetTag(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestCreateTag(t *testing.T) {
	cleanup := setupTagTestDB(t)
	defer cleanup()

	t.Run("create valid tag", func(t *testing.T) {
		tag := models.Tag{Name: "new-tag", Color: "#fecaca"}
		tagJSON, _ := json.Marshal(tag)

		req := httptest.NewRequest("POST", "/tags", bytes.NewBuffer(tagJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTag(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var createdTag models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &createdTag)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if createdTag.Name != "new-tag" {
			t.Errorf("Expected tag name 'new-tag', got %s", createdTag.Name)
		}

		if createdTag.ID == uuid.Nil {
			t.Error("Expected tag to have generated ID")
		}
	})

	t.Run("create tag without color", func(t *testing.T) {
		tag := models.Tag{Name: "colorless"}
		tagJSON, _ := json.Marshal(tag)

		req := httptest.NewRequest("POST", "/tags", bytes.NewBuffer(tagJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTag(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var createdTag models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &createdTag)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if createdTag.Color == "" {
			t.Error("Expected color to be assigned")
		}
	})

	t.Run("create tag without name", func(t *testing.T) {
		tag := models.Tag{Color: "#fecaca"}
		tagJSON, _ := json.Marshal(tag)

		req := httptest.NewRequest("POST", "/tags", bytes.NewBuffer(tagJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTag(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/tags", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTag(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestUpdateTag(t *testing.T) {
	cleanup := setupTagTestDB(t)
	defer cleanup()

	tag := models.Tag{Name: "original", Color: "#fecaca"}
	database.DB.Create(&tag)

	t.Run("update existing tag", func(t *testing.T) {
		updateData := models.Tag{Name: "updated", Color: "#fed7aa"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tags/"+tag.ID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTag(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var updatedTag models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &updatedTag)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if updatedTag.Name != "updated" {
			t.Errorf("Expected updated name 'updated', got %s", updatedTag.Name)
		}

		if updatedTag.Color != "#fed7aa" {
			t.Errorf("Expected updated color '#fed7aa', got %s", updatedTag.Color)
		}
	})

	t.Run("update only name", func(t *testing.T) {
		tag2 := models.Tag{Name: "test2", Color: "#fecaca"}
		database.DB.Create(&tag2)

		updateData := models.Tag{Name: "name-only"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tags/"+tag2.ID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTag(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var updatedTag models.Tag
		err := json.Unmarshal(w.Body.Bytes(), &updatedTag)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if updatedTag.Name != "name-only" {
			t.Errorf("Expected updated name 'name-only', got %s", updatedTag.Name)
		}

		if updatedTag.Color != "#fecaca" {
			t.Errorf("Expected color to remain '#fecaca', got %s", updatedTag.Color)
		}
	})

	t.Run("invalid tag ID", func(t *testing.T) {
		updateData := models.Tag{Name: "updated"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tags/invalid-uuid", bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTag(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("tag not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		updateData := models.Tag{Name: "updated"}
		updateJSON, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/tags/"+nonExistentID.String(), bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTag(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/tags/"+tag.ID.String(), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateTag(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestDeleteTag(t *testing.T) {
	cleanup := setupTagTestDB(t)
	defer cleanup()

	tag := models.Tag{Name: "delete-me", Color: "#fecaca"}
	database.DB.Create(&tag)

	t.Run("delete existing tag", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tags/"+tag.ID.String(), nil)
		w := httptest.NewRecorder()

		DeleteTag(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		var deletedTag models.Tag
		result := database.DB.First(&deletedTag, "id = ?", tag.ID)
		if result.Error == nil {
			t.Error("Expected tag to be deleted")
		}
	})

	t.Run("invalid tag ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tags/invalid-uuid", nil)
		w := httptest.NewRecorder()

		DeleteTag(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("tag not found", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("DELETE", "/tags/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		DeleteTag(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}
