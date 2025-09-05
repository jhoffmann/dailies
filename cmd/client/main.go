package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
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

var tagNames = []string{
	"work",
	"personal",
	"urgent",
	"low-priority",
	"development",
	"testing",
	"documentation",
	"meeting",
	"maintenance",
	"security",
	"performance",
	"frontend",
	"backend",
	"devops",
}

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

	err = db.AutoMigrate(&models.Task{}, &models.Tag{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// listCommand handles the list subcommand to retrieve and display all tasks in JSON format.
func listCommand(args []string) {
	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	database := listFlags.String("database", "dailies.db", "Path to the SQLite database file")

	listFlags.Parse(args)

	db, err := connectToDatabase(*database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	tasks, err := models.GetTasks(db, nil, "", []uuid.UUID{}, "")
	if err != nil {
		log.Fatalf("Failed to retrieve tasks: %v", err)
	}

	jsonData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal tasks to JSON: %v", err)
	}

	fmt.Fprint(os.Stdout, string(jsonData))
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

// populateWithSampleData clears existing tasks and tags, then creates sample data.
func populateWithSampleData(db *gorm.DB, count int) error {
	// Clear existing data
	db.Exec("DELETE FROM task_tags")

	result := db.Where("1 = 1").Delete(&models.Task{})
	if result.Error != nil {
		return fmt.Errorf("failed to clear existing tasks: %w", result.Error)
	}
	log.Printf("Cleared %d existing tasks", result.RowsAffected)

	result = db.Where("1 = 1").Delete(&models.Tag{})
	if result.Error != nil {
		return fmt.Errorf("failed to clear existing tags: %w", result.Error)
	}
	log.Printf("Cleared %d existing tags", result.RowsAffected)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create sample tags
	var createdTags []models.Tag
	for _, tagName := range tagNames {
		tag := models.Tag{Name: tagName}
		err := tag.Create(db)
		if err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tagName, err)
		}
		createdTags = append(createdTags, tag)
	}
	log.Printf("Created %d sample tags", len(createdTags))

	// Create sample tasks with random tags
	for i := range count {
		task := models.Task{
			Name:      getRandomTaskName(r),
			Completed: r.Float32() < 0.3, // 30% chance of being completed
			Priority:  r.Intn(5) + 1,     // Random priority 1-5
		}

		// Add random tags to each task (0-3 tags per task)
		numTags := r.Intn(4)
		if numTags > 0 {
			tagIndices := r.Perm(len(createdTags))
			for j := 0; j < numTags && j < len(tagIndices); j++ {
				task.Tags = append(task.Tags, createdTags[tagIndices[j]])
			}
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
