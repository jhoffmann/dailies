package services

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jhoffmann/dailies/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.Task{}, &models.Frequency{}, &models.Tag{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func setupTestScheduler(t *testing.T) (*TaskScheduler, *gorm.DB) {
	db := setupTestDB(t)
	location, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatalf("Failed to load UTC timezone: %v", err)
	}
	scheduler := NewTaskScheduler(db, location, "UTC")
	return scheduler, db
}

func TestNewTaskScheduler(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	if scheduler.db != db {
		t.Error("Expected scheduler to store database reference")
	}

	if scheduler.cron == nil {
		t.Error("Expected scheduler to initialize cron instance")
	}

	if scheduler.timezone != "UTC" {
		t.Error("Expected scheduler to store timezone")
	}
}

func TestTaskSchedulerStartStop(t *testing.T) {
	scheduler, _ := setupTestScheduler(t)

	// Test start
	scheduler.Start()

	// Test stop
	scheduler.Stop()
}

func TestResetCompletedTasksWithValidCronExpression(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	// Create a frequency with daily reset at midnight
	frequency := &models.Frequency{
		Name:   "Daily",
		Period: "0 0 * * *", // Daily at midnight
	}
	err := db.Create(frequency).Error
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	// Create a completed task with this frequency, updated yesterday
	yesterday := time.Now().Add(-24 * time.Hour)
	task := &models.Task{
		Name:        "Test Task",
		Completed:   true,
		FrequencyID: &frequency.ID,
		UpdatedAt:   yesterday,
	}
	err = db.Create(task).Error
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Run the reset function
	scheduler.resetCompletedTasks()

	// Reload the task and check if it was reset
	err = db.First(task, "id = ?", task.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task: %v", err)
	}

	if task.Completed {
		t.Error("Expected task to be reset (completed=false), but it's still completed")
	}
}

func TestResetCompletedTasksWithInvalidCronExpression(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	// Create a frequency with invalid cron expression
	frequency := &models.Frequency{
		Name:   "Invalid",
		Period: "invalid cron", // Invalid expression
	}
	err := db.Create(frequency).Error
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	// Create a completed task with this frequency
	task := &models.Task{
		Name:        "Test Task",
		Completed:   true,
		FrequencyID: &frequency.ID,
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}
	err = db.Create(task).Error
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Run the reset function (should not crash)
	scheduler.resetCompletedTasks()

	// Reload the task and check if it wasn't reset due to invalid cron
	err = db.First(task, "id = ?", task.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task: %v", err)
	}

	if !task.Completed {
		t.Error("Expected task to remain completed due to invalid cron expression")
	}
}

func TestResetCompletedTasksNotYetDue(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	// Create a frequency with daily reset at midnight
	frequency := &models.Frequency{
		Name:   "Daily",
		Period: "0 0 * * *", // Daily at midnight
	}
	err := db.Create(frequency).Error
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	// Create a completed task with this frequency, updated just now
	// This should not be reset because the next reset time hasn't passed
	task := &models.Task{
		Name:        "Test Task",
		Completed:   true,
		FrequencyID: &frequency.ID,
		UpdatedAt:   time.Now(),
	}
	err = db.Create(task).Error
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Run the reset function
	scheduler.resetCompletedTasks()

	// Reload the task and check if it wasn't reset
	err = db.First(task, "id = ?", task.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task: %v", err)
	}

	if !task.Completed {
		t.Error("Expected task to remain completed because reset time hasn't passed")
	}
}

func TestResetCompletedTasksWithoutFrequency(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	// Create a completed task without frequency
	task := &models.Task{
		Name:      "Test Task",
		Completed: true,
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	err := db.Create(task).Error
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Run the reset function
	scheduler.resetCompletedTasks()

	// Reload the task and check if it wasn't reset
	err = db.First(task, "id = ?", task.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task: %v", err)
	}

	if !task.Completed {
		t.Error("Expected task without frequency to remain completed")
	}
}

func TestResetCompletedTasksMultipleTasks(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	// Create a frequency with daily reset at midnight
	frequency := &models.Frequency{
		Name:   "Daily",
		Period: "0 0 * * *",
	}
	err := db.Create(frequency).Error
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	yesterday := time.Now().Add(-24 * time.Hour)

	// Create multiple completed tasks with this frequency
	task1 := &models.Task{
		Name:        "Task 1",
		Completed:   true,
		FrequencyID: &frequency.ID,
		UpdatedAt:   yesterday,
	}
	task2 := &models.Task{
		Name:        "Task 2",
		Completed:   true,
		FrequencyID: &frequency.ID,
		UpdatedAt:   yesterday,
	}
	task3 := &models.Task{
		Name:        "Task 3",
		Completed:   false, // Not completed, should remain unchanged
		FrequencyID: &frequency.ID,
		UpdatedAt:   yesterday,
	}

	err = db.Create([]*models.Task{task1, task2, task3}).Error
	if err != nil {
		t.Fatalf("Failed to create tasks: %v", err)
	}

	// Run the reset function
	scheduler.resetCompletedTasks()

	// Reload tasks and check results
	err = db.First(task1, "id = ?", task1.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task1: %v", err)
	}
	err = db.First(task2, "id = ?", task2.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task2: %v", err)
	}
	err = db.First(task3, "id = ?", task3.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task3: %v", err)
	}

	if task1.Completed {
		t.Error("Expected task1 to be reset")
	}
	if task2.Completed {
		t.Error("Expected task2 to be reset")
	}
	if task3.Completed {
		t.Error("Expected task3 to remain incomplete")
	}
}

func TestResetCompletedTasksWithHourlyFrequency(t *testing.T) {
	scheduler, db := setupTestScheduler(t)

	// Create a frequency with hourly reset
	frequency := &models.Frequency{
		Name:   "Hourly",
		Period: "0 * * * *", // Every hour
	}
	err := db.Create(frequency).Error
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	// Create a completed task updated 2 hours ago
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	task := &models.Task{
		Name:        "Hourly Task",
		Completed:   true,
		FrequencyID: &frequency.ID,
		UpdatedAt:   twoHoursAgo,
	}
	err = db.Create(task).Error
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Run the reset function
	scheduler.resetCompletedTasks()

	// Reload the task and check if it was reset
	err = db.First(task, "id = ?", task.ID).Error
	if err != nil {
		t.Fatalf("Failed to reload task: %v", err)
	}

	if task.Completed {
		t.Error("Expected hourly task to be reset after 2 hours")
	}
}
