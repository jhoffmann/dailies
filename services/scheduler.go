// Package services provides business logic and background services for the dailies application.
package services

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"github.com/jhoffmann/dailies/models"
)

// TaskScheduler manages background task resets based on frequency schedules.
// It runs a single cron job every minute that checks all completed tasks with frequencies
// and resets them if their scheduled reset time has passed.
type TaskScheduler struct {
	db        *gorm.DB
	cron      *cron.Cron
	wsManager *WebSocketManager
	location  *time.Location
	timezone  string
}

// NewTaskScheduler creates a new task scheduler instance with the provided database connection and timezone.
func NewTaskScheduler(db *gorm.DB, location *time.Location, timezone string) *TaskScheduler {
	return &TaskScheduler{
		db:       db,
		cron:     cron.New(cron.WithLocation(location)),
		location: location,
		timezone: timezone,
	}
}

// SetWebSocketManager sets the WebSocket manager for broadcasting events
func (ts *TaskScheduler) SetWebSocketManager(wsManager *WebSocketManager) {
	ts.wsManager = wsManager
}

// Start begins the background scheduler that checks for task resets every minute.
// This approach is fully dynamic - it automatically handles tasks and frequencies
// created after the service starts without requiring restart or reconfiguration.
func (ts *TaskScheduler) Start() {
	// Check every minute for tasks that need to be reset
	_, err := ts.cron.AddFunc("* * * * *", func() {
		ts.resetCompletedTasks()
	})
	if err != nil {
		log.Printf("Failed to schedule task reset job: %v", err)
		return
	}

	ts.cron.Start()
	log.Println("Task scheduler started")
}

// Stop stops the background scheduler gracefully.
func (ts *TaskScheduler) Stop() {
	ts.cron.Stop()
	log.Println("Task scheduler stopped")
}

// GetTimezone returns the configured timezone name.
func (ts *TaskScheduler) GetTimezone() string {
	return ts.timezone
}

// GetLocation returns the configured time location.
func (ts *TaskScheduler) GetLocation() *time.Location {
	return ts.location
}

// resetCompletedTasks checks all completed tasks with frequencies and resets them
// if their scheduled reset time has passed. This method runs every minute and handles
// all frequency-based task resets dynamically.
func (ts *TaskScheduler) resetCompletedTasks() {
	var tasks []models.Task

	// Get all completed tasks that have frequencies and are not deleted
	result := ts.db.Preload("Frequency").
		Where("completed = ? AND frequency_id IS NOT NULL AND deleted = ?", true, false).
		Find(&tasks)

	if result.Error != nil {
		log.Printf("Error fetching tasks for reset check: %v", result.Error)
		return
	}

	now := time.Now().In(ts.location)
	resetCount := 0

	for _, task := range tasks {
		if task.Frequency == nil {
			continue
		}

		// Parse the 5-field cron expression (format: "minute hour day month day-of-week")
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

		// Prefix the Period with timezone to ensure it uses the correct timezone
		schedule, err := parser.Parse("TZ=" + ts.timezone + " " + task.Frequency.Period)
		if err != nil {
			log.Printf("Invalid cron expression '%s' for task %s: %v",
				task.Frequency.Period, task.Name, err)
			continue
		}

		// Calculate when this task should next reset after it was completed
		// The task should only reset after the next scheduled reset time following completion
		nextReset := schedule.Next(task.UpdatedAt)

		// If the scheduled reset time has passed, reset the task
		if nextReset.Before(now) || nextReset.Equal(now) {
			err := ts.db.Model(&task).Update("completed", false).Error
			if err != nil {
				log.Printf("Error resetting task %s: %v", task.Name, err)
				continue
			}
			resetCount++

			log.Printf("Reset task '%s' (frequency: %s)", task.Name, task.Frequency.Name)

			// Broadcast the task reset event
			if ts.wsManager != nil {
				// Reload the task to get the latest state for broadcasting
				var updatedTask models.Task
				if err := ts.db.Preload("Tags").Preload("Frequency").First(&updatedTask, "id = ?", task.ID).Error; err == nil {
					ts.wsManager.Broadcast(EventTaskReset, updatedTask)
				}
			}
		}
	}

	if resetCount > 0 {
		log.Printf("Reset %d tasks", resetCount)
	}
}
