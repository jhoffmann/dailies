package database

import (
	"os"
	"testing"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/gorm"
)

func TestInit(t *testing.T) {
	t.Run("initializes database with default path", func(t *testing.T) {
		originalDB := DB
		defer func() {
			DB = originalDB
		}()

		DB = nil
		Init("dailies.db")

		if DB == nil {
			t.Error("Expected DB to be initialized")
		}

		var task models.Task
		err := DB.AutoMigrate(&task)
		if err != nil {
			t.Errorf("Expected migration to work: %v", err)
		}
	})

	t.Run("initializes database with custom path", func(t *testing.T) {
		testDB := "test_dailies.db"
		defer os.Remove(testDB)

		originalDB := DB
		defer func() {
			DB = originalDB
		}()

		DB = nil
		Init(testDB)

		if DB == nil {
			t.Error("Expected DB to be initialized")
		}

		if _, err := os.Stat(testDB); os.IsNotExist(err) {
			t.Error("Expected database file to be created")
		}
	})
}

func TestGetDB(t *testing.T) {
	originalDB := DB
	defer func() {
		DB = originalDB
	}()

	testDB := &gorm.DB{}
	DB = testDB

	result := GetDB()
	if result != testDB {
		t.Error("Expected GetDB to return the set database instance")
	}
}

func TestGetDB_Nil(t *testing.T) {
	originalDB := DB
	defer func() {
		DB = originalDB
	}()

	DB = nil
	result := GetDB()
	if result != nil {
		t.Error("Expected GetDB to return nil when DB is nil")
	}
}

func TestInit_WithExistingFile(t *testing.T) {
	testDB := "test_existing.db"
	defer os.Remove(testDB)

	originalDB := DB
	defer func() {
		DB = originalDB
	}()

	DB = nil
	Init(testDB)

	if DB == nil {
		t.Error("Expected DB to be initialized")
	}

	if _, err := os.Stat(testDB); os.IsNotExist(err) {
		t.Error("Expected database file to be created")
	}
}
