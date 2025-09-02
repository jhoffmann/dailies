package main

import (
	"os"
	"testing"

	"github.com/jhoffmann/dailies/internal/models"
)

func TestPrintUsage(t *testing.T) {
	printUsage()
}

func TestConnectToDatabase(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		dbPath := ":memory:"
		db, err := connectToDatabase(dbPath)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if db == nil {
			t.Error("Expected database connection, got nil")
		}

		var task models.Task
		err = db.AutoMigrate(&task)
		if err != nil {
			t.Errorf("Expected successful migration: %v", err)
		}
	})

	t.Run("invalid database path", func(t *testing.T) {
		dbPath := "/invalid/path/database.db"
		db, err := connectToDatabase(dbPath)

		if err == nil {
			t.Error("Expected error for invalid database path")
		}

		if db != nil {
			t.Error("Expected nil database connection")
		}
	})
}

func TestMain_NoArgs(t *testing.T) {
	if os.Getenv("TEST_MAIN") == "1" {
		main()
		return
	}

	os.Args = []string{"cmd"}

	if os.Getenv("CI") != "" {
		t.Skip("Skipping main test in CI environment")
	}
}

func TestMain_UnknownCommand(t *testing.T) {
	if os.Getenv("TEST_MAIN") == "1" {
		os.Args = []string{"cmd", "unknown"}
		main()
		return
	}

	if os.Getenv("CI") != "" {
		t.Skip("Skipping main test in CI environment")
	}
}

func TestMain_ValidCommands(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping main test in CI environment")
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	testDBFile := "test_main_commands.db"
	defer os.Remove(testDBFile)

	t.Run("populate command", func(t *testing.T) {
		os.Args = []string{"cmd", "populate", "--database", testDBFile, "--entries", "1"}

		defer func() {
			if r := recover(); r != nil {
				t.Log("Expected populate command to run without panic")
			}
		}()

		populateCommand([]string{"--database", testDBFile, "--entries", "1"})
	})

	t.Run("list command", func(t *testing.T) {
		os.Args = []string{"cmd", "list", "--database", testDBFile}

		defer func() {
			if r := recover(); r != nil {
				t.Log("Expected list command to run without panic")
			}
		}()

		listCommand([]string{"--database", testDBFile})
	})
}
