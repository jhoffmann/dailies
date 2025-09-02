package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/gorm"
)

var taskNames = []string{
	"Complete daily standup",
	"Review pull requests",
	"Update documentation",
	"Write unit tests",
	"Fix reported bugs",
	"Refactor legacy code",
	"Design new feature",
	"Optimize database queries",
	"Deploy to production",
	"Monitor system metrics",
	"Research new technologies",
	"Code review feedback",
	"Update dependencies",
	"Backup database",
	"Security audit",
	"Performance testing",
	"User acceptance testing",
	"Create API endpoints",
	"Frontend development",
	"Mobile app updates",
	"DevOps improvements",
	"Infrastructure maintenance",
	"Team meeting",
	"Client presentation",
	"Project planning",
}

// populateCommand handles the populate subcommand to add sample data to the database.
func populateCommand(args []string) {
	populateFlags := flag.NewFlagSet("populate", flag.ExitOnError)
	database := populateFlags.String("database", "dailies.db", "Path to the SQLite database file")
	entries := populateFlags.Int("entries", 10, "Number of sample entries to create")

	populateFlags.Parse(args)

	if *entries <= 0 {
		log.Fatal("Number of entries must be greater than 0")
	}

	db, err := connectToDatabase(*database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Connected to database: %s", *database)

	err = populateWithSampleData(db, *entries)
	if err != nil {
		log.Fatalf("Failed to populate database: %v", err)
	}

	log.Printf("Successfully created %d sample tasks", *entries)
}

// populateWithSampleData creates the specified number of sample tasks in the database.
func populateWithSampleData(db *gorm.DB, count int) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range count {
		task := models.Task{
			Name:      getRandomTaskName(r),
			Completed: r.Float32() < 0.3, // 30% chance of being completed
		}

		result := db.Create(&task)
		if result.Error != nil {
			return fmt.Errorf("failed to create task %d: %w", i+1, result.Error)
		}
	}

	return nil
}

// getRandomTaskName returns a random task name from the predefined list.
func getRandomTaskName(r *rand.Rand) string {
	return taskNames[r.Intn(len(taskNames))]
}
