package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jhoffmann/dailies/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetHealth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database with working connection
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate to ensure database is properly set up
	err = db.AutoMigrate(&models.Task{}, &models.Tag{}, &models.Frequency{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	r := gin.New()
	r.GET("/health", GetHealth(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check response body
	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}
}

func TestGetHealth_DatabaseConnectionFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup database and then close the connection to simulate failure
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Close the database connection to simulate failure
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.Close()

	r := gin.New()
	r.GET("/health", GetHealth(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	// Check response body
	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%v'", response["status"])
	}

	if response["message"] != "Database connection is not active" {
		t.Errorf("Expected message 'Database connection is not active', got '%v'", response["message"])
	}
}

func TestGetHealth_InvalidDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a database with an invalid configuration to simulate DB() failure
	db, err := gorm.Open(sqlite.Open("/invalid/path/to/database.db"), &gorm.Config{})
	if err != nil {
		// If we can't even open the invalid database, create one and corrupt it
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}

		// Close underlying connection and set to nil to simulate DB() failure
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	r := gin.New()
	r.GET("/health", GetHealth(db))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	// Should return service unavailable for any database error
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	// Check response body contains error status
	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got '%v'", response["status"])
	}

	// Should contain either connection failure or DB access failure message
	message, exists := response["message"]
	if !exists {
		t.Error("Expected error message in response")
	}

	messageStr, ok := message.(string)
	if !ok {
		t.Error("Expected message to be a string")
	}

	expectedMessages := []string{
		"Failed to get database connection",
		"Database connection is not active",
	}

	found := false
	for _, expected := range expectedMessages {
		if messageStr == expected {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected one of %v, got '%v'", expectedMessages, messageStr)
	}
}
