package config

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetupDatabase(t *testing.T) {
	// Use a temporary file for testing
	tempFile := "test_database.db"
	defer os.Remove(tempFile)

	db, err := SetupDatabase(tempFile)
	if err != nil {
		t.Fatalf("SetupDatabase failed: %v", err)
	}

	if db == nil {
		t.Fatal("Expected database instance, got nil")
	}

	// Verify the database file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Expected database file to be created")
	}

	// Verify tables exist by checking if we can query them
	result := db.Exec("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('tasks', 'frequencies', 'tags')")
	if result.Error != nil {
		t.Errorf("Failed to query tables: %v", result.Error)
	}
}

func TestSetupDatabaseWithInMemory(t *testing.T) {
	db, err := SetupDatabase(":memory:")
	if err != nil {
		t.Fatalf("SetupDatabase with in-memory failed: %v", err)
	}

	if db == nil {
		t.Fatal("Expected database instance, got nil")
	}

	// Verify we can perform basic operations
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get SQL DB: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestSetupDatabaseWithInvalidPath(t *testing.T) {
	// Try to create database in non-existent directory
	invalidPath := "/non/existent/directory/test.db"

	_, err := SetupDatabase(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid database path, got nil")
	}
}

func TestMigrate(t *testing.T) {
	// Create in-memory database for testing
	db, err := gorm.Open(sqliteDialector(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = migrate(db)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Verify tables were created
	tables := []string{"tasks", "frequencies", "tags", "task_tags"}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("Expected table %s to exist", table)
		}
	}
}

func TestAddIndexes(t *testing.T) {
	// Create in-memory database for testing
	db, err := gorm.Open(sqliteDialector(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// First migrate to create tables
	err = migrate(db)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Check if indexes exist by querying sqlite_master
	var count int64
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'").Scan(&count)

	if count == 0 {
		t.Error("Expected indexes to be created")
	}
}

// Helper function to create SQLite dialector for testing
func sqliteDialector(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}
