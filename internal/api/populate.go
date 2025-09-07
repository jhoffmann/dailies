// Package api provides HTTP handlers for population operations
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
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

var tagNames = []string{
	"work",
	"personal",
	"urgent",
	"development",
	"testing",
	"documentation",
	"security",
	"performance",
	"frontend",
	"backend",
	"devops",
}

// PopulateSampleData handles the populate sample data endpoint
func PopulateSampleData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		return
	}

	// Get count parameter from query string, default to 10
	countStr := r.URL.Query().Get("count")
	count := 10
	if countStr != "" {
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			logger.LoggedError(w, "Invalid count parameter", http.StatusBadRequest, r)
			return
		}
	}

	db := database.GetDB()
	if db == nil {
		logger.LoggedError(w, "Database connection not available", http.StatusInternalServerError, r)
		return
	}

	err := populateWithSampleData(db, count)
	if err != nil {
		logger.LoggedError(w, fmt.Sprintf("Failed to populate database: %v", err), http.StatusInternalServerError, r)
		return
	}

	response := map[string]any{
		"success": true,
		"message": fmt.Sprintf("Successfully created %d sample tasks", count),
		"count":   count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// populateWithSampleData clears existing tasks, tags, and frequencies, then creates sample data.
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

	result = db.Where("1 = 1").Delete(&models.Frequency{})
	if result.Error != nil {
		return fmt.Errorf("failed to clear existing frequencies: %w", result.Error)
	}
	log.Printf("Cleared %d existing frequencies", result.RowsAffected)

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

	// Create sample frequencies
	frequencies := []struct {
		name  string
		reset string
	}{
		{"hourly", "0 * * * *"},
		{"daily", "0 0 * * *"},
		{"weekly", "0 0 * * 1"},
		{"quarter", "*/15 * * * *"},
		{"fast", "*/2 * * * *"},
	}

	var createdFrequencies []models.Frequency
	for _, freq := range frequencies {
		frequency := models.Frequency{Name: freq.name, Reset: freq.reset}
		err := frequency.Create(db)
		if err != nil {
			return fmt.Errorf("failed to create frequency %s: %w", freq.name, err)
		}
		createdFrequencies = append(createdFrequencies, frequency)
	}
	log.Printf("Created %d sample frequencies", len(createdFrequencies))

	// Create sample tasks with random tags and frequencies
	for i := range count {
		task := models.Task{
			Name:      getRandomTaskName(r),
			Completed: r.Float32() < 0.3, // 30% chance of being completed
			Priority:  r.Intn(5) + 1,     // Random priority 1-5
		}

		// Assign random frequency to each task (70% chance of having a frequency)
		if r.Float32() < 0.7 && len(createdFrequencies) > 0 {
			frequencyIndex := r.Intn(len(createdFrequencies))
			task.FrequencyID = &createdFrequencies[frequencyIndex].ID
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
