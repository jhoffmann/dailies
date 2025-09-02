package main

import (
	"fmt"
	"os"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// main is the entry point for the client application that handles subcommands.
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "populate":
		populateCommand(args)
	case "list":
		listCommand(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays usage information for the client application.
func printUsage() {
	fmt.Println("Usage: go run cmd/client/main.go <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  populate --database <path> --entries <count>  Populate database with sample data")
	fmt.Println("  list --database <path>                        List all tasks in JSON format")
}

// connectToDatabase establishes a connection to the SQLite database and performs migrations.
func connectToDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
