package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/routes"
)

// main initializes the database, sets up routes, and starts the HTTP server.
// It accepts an --address flag to specify the listening address in the form ip:port or :port.
// It accepts a --db flag to specify the database file path.
func main() {
	address := flag.String("address", ":9001", "Address to listen on in the form ip:port or :port")
	dbPath := flag.String("db", "testing.db", "Database file path")
	flag.Parse()

	database.Init(*dbPath)

	routes.Setup()

	log.Printf("Server starting on %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
