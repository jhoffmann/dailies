package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jhoffmann/dailies/internal/models"
)

// listCommand handles the list subcommand to retrieve and display all tasks in JSON format.
func listCommand(args []string) {
	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	database := listFlags.String("database", "dailies.db", "Path to the SQLite database file")

	listFlags.Parse(args)

	db, err := connectToDatabase(*database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	var tasks []models.Task
	result := db.Find(&tasks)
	if result.Error != nil {
		log.Fatalf("Failed to retrieve tasks: %v", result.Error)
	}

	jsonData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal tasks to JSON: %v", err)
	}

	fmt.Fprint(os.Stdout, string(jsonData))
}
