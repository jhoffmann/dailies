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

func setupFrequencyTestDB(t *testing.T) func() {
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

func TestGetFrequencies(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	freq1 := models.Frequency{Name: "Daily", Reset: "0 18 * * *"}
	freq2 := models.Frequency{Name: "Weekly", Reset: "0 23 * * 1"}
	freq3 := models.Frequency{Name: "Monthly", Reset: "0 0 15 * *"}

	database.DB.Create(&freq1)
	database.DB.Create(&freq2)
	database.DB.Create(&freq3)

	t.Run("get all frequencies", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies", nil)
		w := httptest.NewRecorder()

		GetFrequencies(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var frequencies []models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &frequencies)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(frequencies) != 3 {
			t.Errorf("Expected 3 frequencies, got %d", len(frequencies))
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies?name=Week", nil)
		w := httptest.NewRecorder()

		GetFrequencies(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var frequencies []models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &frequencies)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(frequencies) != 1 {
			t.Errorf("Expected 1 frequency with 'Week' in name, got %d", len(frequencies))
		}

		if frequencies[0].Name != "Weekly" {
			t.Errorf("Expected name 'Weekly', got %s", frequencies[0].Name)
		}
	})

	t.Run("no matching filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies?name=NonExistent", nil)
		w := httptest.NewRecorder()

		GetFrequencies(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var frequencies []models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &frequencies)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(frequencies) != 0 {
			t.Errorf("Expected 0 frequencies, got %d", len(frequencies))
		}
	})
}

func TestGetFrequency(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	frequency := models.Frequency{Name: "Test Frequency", Reset: "0 18 * * *"}
	database.DB.Create(&frequency)

	t.Run("get existing frequency", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies/"+frequency.ID.String(), nil)
		w := httptest.NewRecorder()

		GetFrequency(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var responseFreq models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &responseFreq)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if responseFreq.Name != "Test Frequency" {
			t.Errorf("Expected name 'Test Frequency', got %s", responseFreq.Name)
		}

		if responseFreq.Reset != "0 18 * * *" {
			t.Errorf("Expected reset '0 18 * * *', got %s", responseFreq.Reset)
		}
	})

	t.Run("get non-existent frequency", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("GET", "/frequencies/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		GetFrequency(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid frequency ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/frequencies/invalid-id", nil)
		w := httptest.NewRecorder()

		GetFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestCreateFrequency(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	t.Run("create valid frequency", func(t *testing.T) {
		requestBody := map[string]string{
			"name":  "Test Frequency",
			"reset": "0 6 * * *",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/frequencies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateFrequency(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var responseFreq models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &responseFreq)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if responseFreq.Name != "Test Frequency" {
			t.Errorf("Expected name 'Test Frequency', got %s", responseFreq.Name)
		}

		if responseFreq.Reset != "0 6 * * *" {
			t.Errorf("Expected reset '0 6 * * *', got %s", responseFreq.Reset)
		}
	})

	t.Run("create frequency without name", func(t *testing.T) {
		requestBody := map[string]string{
			"reset": "0 18 * * *",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/frequencies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("create frequency without reset", func(t *testing.T) {
		requestBody := map[string]string{
			"name": "Test Frequency",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/frequencies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("create frequency with duplicate name", func(t *testing.T) {
		existingFreq := models.Frequency{Name: "Duplicate", Reset: "0 18 * * *"}
		database.DB.Create(&existingFreq)

		requestBody := map[string]string{
			"name":  "Duplicate",
			"reset": "0 6 * * *",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/frequencies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/frequencies", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestUpdateFrequency(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	frequency := models.Frequency{Name: "Original", Reset: "0 18 * * *"}
	database.DB.Create(&frequency)

	t.Run("update frequency name", func(t *testing.T) {
		updateBody := map[string]string{
			"name": "Updated Name",
		}
		body, _ := json.Marshal(updateBody)

		req := httptest.NewRequest("PUT", "/frequencies/"+frequency.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateFrequency(w, req)

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

	t.Run("update frequency reset", func(t *testing.T) {
		updateBody := map[string]string{
			"reset": "0 6 * * *",
		}
		body, _ := json.Marshal(updateBody)

		req := httptest.NewRequest("PUT", "/frequencies/"+frequency.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateFrequency(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var responseFreq models.Frequency
		err := json.Unmarshal(w.Body.Bytes(), &responseFreq)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if responseFreq.Reset != "0 6 * * *" {
			t.Errorf("Expected reset '0 6 * * *', got %s", responseFreq.Reset)
		}
	})

	t.Run("update non-existent frequency", func(t *testing.T) {
		nonExistentID := uuid.New()
		updateBody := map[string]string{
			"name": "Updated",
		}
		body, _ := json.Marshal(updateBody)

		req := httptest.NewRequest("PUT", "/frequencies/"+nonExistentID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateFrequency(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("update with duplicate name", func(t *testing.T) {
		otherFreq := models.Frequency{Name: "Other", Reset: "0 23 * * 1"}
		database.DB.Create(&otherFreq)

		updateBody := map[string]string{
			"name": "Other",
		}
		body, _ := json.Marshal(updateBody)

		req := httptest.NewRequest("PUT", "/frequencies/"+frequency.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid frequency ID", func(t *testing.T) {
		updateBody := map[string]string{
			"name": "Updated",
		}
		body, _ := json.Marshal(updateBody)

		req := httptest.NewRequest("PUT", "/frequencies/invalid-id", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/frequencies/"+frequency.ID.String(), bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		UpdateFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestDeleteFrequency(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	t.Run("delete existing frequency", func(t *testing.T) {
		frequency := models.Frequency{Name: "Delete Me", Reset: "0 18 * * *"}
		database.DB.Create(&frequency)

		req := httptest.NewRequest("DELETE", "/frequencies/"+frequency.ID.String(), nil)
		w := httptest.NewRecorder()

		DeleteFrequency(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		var dbFreq models.Frequency
		result := database.DB.First(&dbFreq, "id = ?", frequency.ID)
		if result.Error == nil {
			t.Error("Expected frequency to be deleted")
		}
	})

	t.Run("delete non-existent frequency", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("DELETE", "/frequencies/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		DeleteFrequency(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid frequency ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/frequencies/invalid-id", nil)
		w := httptest.NewRecorder()

		DeleteFrequency(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestCreateFrequencyWithData(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	t.Run("valid frequency creation", func(t *testing.T) {
		frequency, err := CreateFrequencyWithData("Test", "0 18 * * *")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if frequency.Name != "Test" {
			t.Errorf("Expected name 'Test', got %s", frequency.Name)
		}

		if frequency.Reset != "0 18 * * *" {
			t.Errorf("Expected reset '0 18 * * *', got %s", frequency.Reset)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := CreateFrequencyWithData("", "0 18 * * *")
		if err == nil {
			t.Error("Expected error for empty name")
		}

		if err.Error() != "frequency name is required" {
			t.Errorf("Expected 'frequency name is required', got %s", err.Error())
		}
	})

	t.Run("empty reset", func(t *testing.T) {
		_, err := CreateFrequencyWithData("Test", "")
		if err == nil {
			t.Error("Expected error for empty reset")
		}

		if err.Error() != "frequency reset is required" {
			t.Errorf("Expected 'frequency reset is required', got %s", err.Error())
		}
	})
}

func TestGetFrequenciesWithFilter(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	freq1 := models.Frequency{Name: "Daily", Reset: "0 18 * * *"}
	freq2 := models.Frequency{Name: "Weekly", Reset: "0 23 * * 1"}
	database.DB.Create(&freq1)
	database.DB.Create(&freq2)

	t.Run("no filter", func(t *testing.T) {
		frequencies, err := GetFrequenciesWithFilter("")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(frequencies) != 2 {
			t.Errorf("Expected 2 frequencies, got %d", len(frequencies))
		}
	})

	t.Run("with filter", func(t *testing.T) {
		frequencies, err := GetFrequenciesWithFilter("Daily")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(frequencies) != 1 {
			t.Errorf("Expected 1 frequency, got %d", len(frequencies))
		}

		if frequencies[0].Name != "Daily" {
			t.Errorf("Expected name 'Daily', got %s", frequencies[0].Name)
		}
	})
}

func TestGetFrequencyByID(t *testing.T) {
	cleanup := setupFrequencyTestDB(t)
	defer cleanup()

	frequency := models.Frequency{Name: "Test", Reset: "0 18 * * *"}
	database.DB.Create(&frequency)

	t.Run("existing frequency", func(t *testing.T) {
		result, err := GetFrequencyByID(frequency.ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.Name != "Test" {
			t.Errorf("Expected name 'Test', got %s", result.Name)
		}
	})

	t.Run("non-existent frequency", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := GetFrequencyByID(nonExistentID)
		if err == nil {
			t.Error("Expected error for non-existent frequency")
		}
	})
}
