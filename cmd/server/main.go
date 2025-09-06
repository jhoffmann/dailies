package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/routes"
	"github.com/jhoffmann/dailies/internal/scheduler"
)

// main initializes the database, sets up routes, and starts the HTTP server.
// It accepts an --address flag to specify the listening address in the form ip:port or :port.
// It accepts a --db flag to specify the database file path.
// If --db is not provided, it will use the DAILIES_DB_PATH environment variable.
func main() {
	address := flag.String("address", ":9001", "Address to listen on in the form ip:port or :port")
	dbPath := flag.String("db", "", "Database file path")
	flag.Parse()

	// Determine database path: command line flag takes precedence over environment variable
	finalDBPath := *dbPath
	if finalDBPath == "" {
		if envPath := os.Getenv("DAILIES_DB_PATH"); envPath != "" {
			finalDBPath = envPath
		} else {
			finalDBPath = "dailies.db"
		}
	}

	database.Init(finalDBPath)

	routes.Setup()

	// Start the background task scheduler
	taskScheduler := scheduler.NewTaskScheduler(database.GetDB())
	taskScheduler.Start()

	log.Printf("Server starting on %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
