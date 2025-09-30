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

func TestGetTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.GET("/tags", GetTags(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var tags []models.Tag
	err := json.Unmarshal(w.Body.Bytes(), &tags)
	if err != nil {
		t.Errorf("Expected valid JSON array, got error: %v", err)
	}
}

func TestGetTagNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.GET("/tags/:id", GetTag(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags/non-existent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Tag not found") {
		t.Error("Expected 'Tag not found' error message")
	}
}

func TestCreateTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.POST("/tags", CreateTag(db))

	requestBody := `{"name": "Work", "color": "#ff0000"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tags", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestCreateTagInvalidColor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.POST("/tags", CreateTag(db))

	requestBody := `{"name": "Invalid", "color": "red"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tags", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteTagNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.DELETE("/tags/:id", DeleteTag(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tags/non-existent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGenerateRandomColor(t *testing.T) {
	// Test that the function generates valid hex colors
	for i := 0; i < 100; i++ {
		color := generateRandomColor()

		// Check format: starts with # and has 7 characters total
		if len(color) != 7 {
			t.Errorf("Expected color length 7, got %d for color: %s", len(color), color)
		}

		if color[0] != '#' {
			t.Errorf("Expected color to start with #, got: %s", color)
		}

		// Validate it's a proper hex color using our own validator
		if !validateHexColor(color) {
			t.Errorf("Generated color failed validation: %s", color)
		}
	}

	// Test that multiple calls generate different colors (probabilistically)
	colors := make(map[string]bool)
	duplicateFound := false

	for i := 0; i < 50; i++ {
		color := generateRandomColor()
		if colors[color] {
			duplicateFound = true
			break
		}
		colors[color] = true
	}

	// It's extremely unlikely to get duplicates in 50 tries with 16M possibilities
	// but we won't fail the test if it happens, just log it
	if duplicateFound {
		t.Logf("Duplicate color found in 50 generations (this is statistically unlikely but possible)")
	}
}

func TestValidateHexColor(t *testing.T) {
	// Test valid hex colors
	validColors := []string{
		"#000000",
		"#ffffff",
		"#FFFFFF",
		"#123456",
		"#abcdef",
		"#ABCDEF",
		"#ff0000",
		"#00FF00",
		"#0000ff",
	}

	for _, color := range validColors {
		if !validateHexColor(color) {
			t.Errorf("Expected %s to be valid, but validation failed", color)
		}
	}

	// Test invalid hex colors
	invalidColors := []string{
		"",
		"#",
		"#12345",   // too short
		"#1234567", // too long
		"123456",   // missing #
		"#gggggg",  // invalid hex characters
		"#GGGGGG",  // invalid hex characters
		"red",      // not hex at all
		"#12 45",   // contains space
		"#12\t45",  // contains tab
	}

	for _, color := range invalidColors {
		if validateHexColor(color) {
			t.Errorf("Expected %s to be invalid, but validation passed", color)
		}
	}
}

func TestGetTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test tag
	tag := models.Tag{Name: "Work", Color: "#ff0000"}
	db.Create(&tag)

	r := gin.New()
	r.GET("/tags/:id", GetTag(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags/"+tag.ID, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var responseTag models.Tag
	err := json.Unmarshal(w.Body.Bytes(), &responseTag)
	if err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if responseTag.Name != tag.Name {
		t.Errorf("Expected name %s, got %s", tag.Name, responseTag.Name)
	}
}

func TestUpdateTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test tag
	tag := models.Tag{Name: "Work", Color: "#ff0000"}
	db.Create(&tag)

	r := gin.New()
	r.PUT("/tags/:id", UpdateTag(db))

	requestBody := `{"name": "Updated Work", "color": "#00ff00"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tags/"+tag.ID, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var updatedTag models.Tag
	err := json.Unmarshal(w.Body.Bytes(), &updatedTag)
	if err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if updatedTag.Name != "Updated Work" {
		t.Errorf("Expected name 'Updated Work', got %s", updatedTag.Name)
	}

	if updatedTag.Color != "#00ff00" {
		t.Errorf("Expected color '#00ff00', got %s", updatedTag.Color)
	}
}

func TestUpdateTagNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	r := gin.New()
	r.PUT("/tags/:id", UpdateTag(db))

	requestBody := `{"name": "Updated"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tags/non-existent", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateTagInvalidColor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test tag
	tag := models.Tag{Name: "Work", Color: "#ff0000"}
	db.Create(&tag)

	r := gin.New()
	r.PUT("/tags/:id", UpdateTag(db))

	requestBody := `{"color": "invalid-color"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tags/"+tag.ID, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateTagDuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create two test tags
	tag1 := models.Tag{Name: "Work", Color: "#ff0000"}
	tag2 := models.Tag{Name: "Personal", Color: "#00ff00"}
	db.Create(&tag1)
	db.Create(&tag2)

	r := gin.New()
	r.PUT("/tags/:id", UpdateTag(db))

	// Try to update tag2 to have the same name as tag1
	requestBody := `{"name": "Work"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tags/"+tag2.ID, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestDeleteTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestHandlerDB(t)

	// Create a test tag
	tag := models.Tag{Name: "Work", Color: "#ff0000"}
	db.Create(&tag)

	r := gin.New()
	r.DELETE("/tags/:id", DeleteTag(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tags/"+tag.ID, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify tag was deleted
	var deletedTag models.Tag
	result := db.First(&deletedTag, "id = ?", tag.ID)
	if result.Error == nil {
		t.Error("Expected tag to be deleted, but it still exists")
	}
}
