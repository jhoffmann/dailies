package scheduler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jhoffmann/dailies/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Task{}, &models.Tag{}, &models.Frequency{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestResetCompletedTasks(t *testing.T) {
	db := setupTestDB(t)

	// Create a frequency that should trigger immediately (every minute)
	frequency := models.Frequency{
		ID:    uuid.New(),
		Name:  "Every Minute",
		Reset: "* * * * *", // Every minute
	}
	err := frequency.Create(db)
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	// Create a completed task with this frequency, but set DateModified to 2 minutes ago
	// so it should be reset
	task := models.Task{
		ID:           uuid.New(),
		Name:         "Test Task",
		Completed:    true,
		FrequencyID:  &frequency.ID,
		DateModified: time.Now().Add(-2 * time.Minute),
	}
	err = task.Create(db)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Create another completed task without frequency (should not be reset)
	taskWithoutFreq := models.Task{
		ID:        uuid.New(),
		Name:      "Task Without Frequency",
		Completed: true,
	}
	err = taskWithoutFreq.Create(db)
	if err != nil {
		t.Fatalf("Failed to create task without frequency: %v", err)
	}

	// Create scheduler and run reset check
	scheduler := NewTaskScheduler(db)
	scheduler.resetCompletedTasks()

	// Check that the task with frequency was reset
	var updatedTask models.Task
	err = updatedTask.LoadByID(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to load updated task: %v", err)
	}
	if updatedTask.Completed {
		t.Errorf("Expected task to be reset (completed = false), but it was still completed")
	}

	// Check that the task without frequency was not reset
	var unchangedTask models.Task
	err = unchangedTask.LoadByID(db, taskWithoutFreq.ID)
	if err != nil {
		t.Fatalf("Failed to load unchanged task: %v", err)
	}
	if !unchangedTask.Completed {
		t.Errorf("Expected task without frequency to remain completed, but it was reset")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	db := setupTestDB(t)

	scheduler := NewTaskScheduler(db)

	// Test starting the scheduler
	scheduler.Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test stopping the scheduler
	scheduler.Stop()
}

func TestTaskNotResetBeforeScheduledTime(t *testing.T) {
	db := setupTestDB(t)

	// Create a frequency that runs once daily at midnight
	frequency := models.Frequency{
		ID:    uuid.New(),
		Name:  "Daily at Midnight",
		Reset: "0 0 * * *", // Daily at 00:00
	}
	err := frequency.Create(db)
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	// Create a completed task that was completed 1 hour ago
	// Since the next reset isn't until midnight, it should NOT be reset
	task := models.Task{
		ID:           uuid.New(),
		Name:         "Daily Task",
		Completed:    true,
		FrequencyID:  &frequency.ID,
		DateModified: time.Now().Add(-1 * time.Hour), // Completed 1 hour ago
	}
	err = task.Create(db)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Run scheduler reset check
	scheduler := NewTaskScheduler(db)
	scheduler.resetCompletedTasks()

	// Task should still be completed since next reset is at midnight
	var unchangedTask models.Task
	err = unchangedTask.LoadByID(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to load task: %v", err)
	}
	if !unchangedTask.Completed {
		t.Errorf("Expected task to remain completed until scheduled reset time, but it was reset prematurely")
	}
}

func TestInvalidCronExpression(t *testing.T) {
	db := setupTestDB(t)

	// Create a frequency with invalid cron expression
	frequency := models.Frequency{
		ID:    uuid.New(),
		Name:  "Invalid Cron",
		Reset: "invalid cron expression",
	}
	err := frequency.Create(db)
	if err != nil {
		t.Fatalf("Failed to create frequency: %v", err)
	}

	task := models.Task{
		ID:          uuid.New(),
		Name:        "Test Task with Invalid Cron",
		Completed:   true,
		FrequencyID: &frequency.ID,
	}
	err = task.Create(db)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// This should not panic or error, just log and continue
	scheduler := NewTaskScheduler(db)
	scheduler.resetCompletedTasks()

	// Task should remain completed since cron expression was invalid
	var unchangedTask models.Task
	err = unchangedTask.LoadByID(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to load task: %v", err)
	}
	if !unchangedTask.Completed {
		t.Errorf("Expected task with invalid cron to remain completed")
	}
}
