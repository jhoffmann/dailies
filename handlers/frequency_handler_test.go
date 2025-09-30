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
)

func TestGetFrequencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.GET("/frequencies", GetFrequencies(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/frequencies", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var frequencies []models.Frequency
	err := json.Unmarshal(w.Body.Bytes(), &frequencies)
	if err != nil {
		t.Errorf("Expected valid JSON array, got error: %v", err)
	}
}

func TestGetFrequencyNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.GET("/frequencies/:id", GetFrequency(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/frequencies/non-existent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Frequency not found") {
		t.Error("Expected 'Frequency not found' error message")
	}
}

func TestCreateFrequency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.POST("/frequencies", CreateFrequency(db))

	requestBody := `{"name": "Daily", "period": "0 0 * * *"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/frequencies", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestCreateFrequencyInvalidCron(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.POST("/frequencies", CreateFrequency(db))

	requestBody := `{"name": "Invalid", "period": "invalid cron"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/frequencies", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteFrequencyNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.DELETE("/frequencies/:id", DeleteFrequency(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/frequencies/non-existent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetFrequency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test frequency
	frequency := models.Frequency{Name: "Daily", Period: "0 0 * * *"}
	db.Create(&frequency)

	r := gin.New()
	r.GET("/frequencies/:id", GetFrequency(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/frequencies/"+frequency.ID, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var responseFrequency models.Frequency
	err := json.Unmarshal(w.Body.Bytes(), &responseFrequency)
	if err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if responseFrequency.Name != frequency.Name {
		t.Errorf("Expected name %s, got %s", frequency.Name, responseFrequency.Name)
	}
}

func TestUpdateFrequency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test frequency
	frequency := models.Frequency{Name: "Daily", Period: "0 0 * * *"}
	db.Create(&frequency)

	r := gin.New()
	r.PUT("/frequencies/:id", UpdateFrequency(db))

	requestBody := `{"name": "Updated Daily", "period": "0 12 * * *"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/frequencies/"+frequency.ID, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var updatedFrequency models.Frequency
	err := json.Unmarshal(w.Body.Bytes(), &updatedFrequency)
	if err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if updatedFrequency.Name != "Updated Daily" {
		t.Errorf("Expected name 'Updated Daily', got %s", updatedFrequency.Name)
	}

	if updatedFrequency.Period != "0 12 * * *" {
		t.Errorf("Expected period '0 12 * * *', got %s", updatedFrequency.Period)
	}
}

func TestUpdateFrequencyNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.PUT("/frequencies/:id", UpdateFrequency(db))

	requestBody := `{"name": "Updated"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/frequencies/non-existent", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateFrequencyInvalidCron(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test frequency
	frequency := models.Frequency{Name: "Daily", Period: "0 0 * * *"}
	db.Create(&frequency)

	r := gin.New()
	r.PUT("/frequencies/:id", UpdateFrequency(db))

	requestBody := `{"period": "invalid-cron"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/frequencies/"+frequency.ID, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateFrequencyDuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create two test frequencies
	freq1 := models.Frequency{Name: "Daily", Period: "0 0 * * *"}
	freq2 := models.Frequency{Name: "Weekly", Period: "0 0 * * 0"}
	db.Create(&freq1)
	db.Create(&freq2)

	r := gin.New()
	r.PUT("/frequencies/:id", UpdateFrequency(db))

	// Try to update freq2 to have the same name as freq1
	requestBody := `{"name": "Daily"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/frequencies/"+freq2.ID, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestDeleteFrequency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test frequency
	frequency := models.Frequency{Name: "Daily", Period: "0 0 * * *"}
	db.Create(&frequency)

	r := gin.New()
	r.DELETE("/frequencies/:id", DeleteFrequency(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/frequencies/"+frequency.ID, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify frequency was deleted
	var deletedFrequency models.Frequency
	result := db.First(&deletedFrequency, "id = ?", frequency.ID)
	if result.Error == nil {
		t.Error("Expected frequency to be deleted, but it still exists")
	}
}
