// Package scheduler provides background task scheduling for resetting completed tasks
package scheduler

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"github.com/jhoffmann/dailies/internal/models"
)

// TaskScheduler manages background task resets based on frequency schedules
type TaskScheduler struct {
	db   *gorm.DB
	cron *cron.Cron
}

// NewTaskScheduler creates a new task scheduler instance
func NewTaskScheduler(db *gorm.DB) *TaskScheduler {
	return &TaskScheduler{
		db:   db,
		cron: cron.New(),
	}
}

// Start begins the background scheduler that checks for task resets every minute
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

// Stop stops the background scheduler
func (ts *TaskScheduler) Stop() {
	ts.cron.Stop()
	log.Println("Task scheduler stopped")
}

// resetCompletedTasks checks all tasks with frequencies and resets completed ones if their reset time has passed
func (ts *TaskScheduler) resetCompletedTasks() {
	var tasks []models.Task

	// Get all completed tasks that have frequencies
	result := ts.db.Preload("Frequency").
		Where("completed = ? AND frequency_id IS NOT NULL", true).
		Find(&tasks)

	if result.Error != nil {
		log.Printf("Error fetching tasks for reset check: %v", result.Error)
		return
	}

	now := time.Now()
	resetCount := 0

	for _, task := range tasks {
		if task.Frequency == nil {
			continue
		}

		// Parse the cron expression
		schedule, err := cron.ParseStandard(task.Frequency.Reset)
		if err != nil {
			log.Printf("Invalid cron expression '%s' for task %s: %v",
				task.Frequency.Reset, task.Name, err)
			continue
		}

		// Calculate when this task should next reset after it was completed
		// The task should only reset after the next scheduled reset time following completion
		nextReset := schedule.Next(task.DateModified)

		// If the scheduled reset time has passed, reset the task
		if nextReset.Before(now) || nextReset.Equal(now) {
			err := ts.db.Model(&task).Update("completed", false).Error
			if err != nil {
				log.Printf("Error resetting task %s: %v", task.Name, err)
				continue
			}
			resetCount++
			log.Printf("Reset task '%s' (frequency: %s)", task.Name, task.Frequency.Name)
		}
	}

	if resetCount > 0 {
		log.Printf("Reset %d completed tasks", resetCount)
	}
}
