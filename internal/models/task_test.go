package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTaskStruct(t *testing.T) {
	task := Task{
		Name:      "Test Task",
		Completed: true,
	}

	if task.Name != "Test Task" {
		t.Errorf("Expected name 'Test Task', got %s", task.Name)
	}

	if !task.Completed {
		t.Errorf("Expected completed to be true")
	}
}

func setupTaskTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&Task{}, &Tag{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestBeforeCreate(t *testing.T) {
	db := setupTaskTestDB(t)

	t.Run("generates UUID when ID is nil", func(t *testing.T) {
		task := Task{Name: "Test Task"}

		if task.ID != uuid.Nil {
			t.Errorf("Expected ID to be nil initially")
		}

		err := task.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if task.ID == uuid.Nil {
			t.Errorf("Expected ID to be generated")
		}
	})

	t.Run("preserves existing UUID", func(t *testing.T) {
		existingID := uuid.New()
		task := Task{
			ID:   existingID,
			Name: "Test Task",
		}

		err := task.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if task.ID != existingID {
			t.Errorf("Expected ID to remain %s, got %s", existingID, task.ID)
		}
	})
}

func TestTaskCreate(t *testing.T) {
	db := setupTaskTestDB(t)

	t.Run("creates task with name", func(t *testing.T) {
		task := Task{Name: "Test Task"}

		err := task.Create(db)
		if err != nil {
			t.Errorf("Create returned error: %v", err)
		}

		if task.ID == uuid.Nil {
			t.Error("Expected ID to be generated")
		}

		var dbTask Task
		result := db.First(&dbTask, "id = ?", task.ID)
		if result.Error != nil {
			t.Errorf("Task not found in database: %v", result.Error)
		}

		if dbTask.Name != "Test Task" {
			t.Errorf("Expected name 'Test Task', got %s", dbTask.Name)
		}
	})

	t.Run("fails to create task without name", func(t *testing.T) {
		task := Task{Name: ""}

		err := task.Create(db)
		if err == nil {
			t.Error("Expected error when creating task without name")
		}

		if err.Error() != "task name is required" {
			t.Errorf("Expected 'task name is required', got %s", err.Error())
		}
	})
}

func TestTaskSave(t *testing.T) {
	db := setupTaskTestDB(t)

	task := Task{Name: "Test Task", Completed: false}
	err := task.Create(db)
	if err != nil {
		t.Fatalf("Failed to create initial task: %v", err)
	}

	task.Completed = true
	err = task.Save(db)
	if err != nil {
		t.Errorf("Save returned error: %v", err)
	}

	var dbTask Task
	result := db.First(&dbTask, "id = ?", task.ID)
	if result.Error != nil {
		t.Errorf("Task not found: %v", result.Error)
	}

	if !dbTask.Completed {
		t.Error("Expected task to be marked as completed")
	}
}

func TestTaskLoadByID(t *testing.T) {
	db := setupTaskTestDB(t)

	originalTask := Task{Name: "Test Task", Priority: 2}
	err := originalTask.Create(db)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	t.Run("loads existing task", func(t *testing.T) {
		var loadedTask Task
		err := loadedTask.LoadByID(db, originalTask.ID)
		if err != nil {
			t.Errorf("LoadByID returned error: %v", err)
		}

		if loadedTask.Name != "Test Task" {
			t.Errorf("Expected name 'Test Task', got %s", loadedTask.Name)
		}

		if loadedTask.Priority != 2 {
			t.Errorf("Expected priority 2, got %d", loadedTask.Priority)
		}
	})

	t.Run("loads task with tags", func(t *testing.T) {
		taskWithTags := Task{Name: "Tagged Task"}
		tag := Tag{Name: "work", Color: "#fecaca"}
		db.Create(&tag)
		taskWithTags.Tags = []Tag{tag}
		db.Create(&taskWithTags)

		var loadedTask Task
		err := loadedTask.LoadByID(db, taskWithTags.ID)
		if err != nil {
			t.Errorf("LoadByID returned error: %v", err)
		}

		if len(loadedTask.Tags) != 1 {
			t.Errorf("Expected 1 tag, got %d", len(loadedTask.Tags))
		}

		if loadedTask.Tags[0].Name != "work" {
			t.Errorf("Expected tag name 'work', got %s", loadedTask.Tags[0].Name)
		}
	})

	t.Run("fails to load non-existent task", func(t *testing.T) {
		var loadedTask Task
		nonExistentID := uuid.New()
		err := loadedTask.LoadByID(db, nonExistentID)
		if err == nil {
			t.Error("Expected error when loading non-existent task")
		}

		if err.Error() != "task not found" {
			t.Errorf("Expected 'task not found', got %s", err.Error())
		}
	})
}

func TestTaskUpdate(t *testing.T) {
	db := setupTaskTestDB(t)

	task := Task{Name: "Original Task", Priority: 3, Completed: false}
	err := task.Create(db)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	t.Run("updates name and completed status", func(t *testing.T) {
		updateData := Task{Name: "Updated Task", Completed: true}
		err := task.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if task.Name != "Updated Task" {
			t.Errorf("Expected name 'Updated Task', got %s", task.Name)
		}

		if !task.Completed {
			t.Error("Expected task to be marked as completed")
		}
	})

	t.Run("updates priority", func(t *testing.T) {
		updateData := Task{Priority: 1}
		err := task.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if task.Priority != 1 {
			t.Errorf("Expected priority 1, got %d", task.Priority)
		}
	})

	t.Run("ignores invalid priority", func(t *testing.T) {
		originalPriority := task.Priority
		updateData := Task{Priority: 6} // Invalid priority
		err := task.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if task.Priority != originalPriority {
			t.Errorf("Expected priority to remain %d, got %d", originalPriority, task.Priority)
		}
	})

	t.Run("updates only completed status", func(t *testing.T) {
		originalName := task.Name
		updateData := Task{Completed: false}
		err := task.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if task.Name != originalName {
			t.Errorf("Expected name to remain %s, got %s", originalName, task.Name)
		}

		if task.Completed {
			t.Error("Expected task to be marked as incomplete")
		}
	})
}

func TestTaskDelete(t *testing.T) {
	db := setupTaskTestDB(t)

	t.Run("deletes existing task", func(t *testing.T) {
		task := Task{Name: "Delete Me"}
		err := task.Create(db)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		err = task.Delete(db)
		if err != nil {
			t.Errorf("Delete returned error: %v", err)
		}

		var dbTask Task
		result := db.First(&dbTask, "id = ?", task.ID)
		if result.Error == nil {
			t.Error("Expected task to be deleted")
		}
	})

	t.Run("fails to delete non-existent task", func(t *testing.T) {
		task := Task{ID: uuid.New()}
		err := task.Delete(db)
		if err == nil {
			t.Error("Expected error when deleting non-existent task")
		}

		if err.Error() != "task not found" {
			t.Errorf("Expected 'task not found', got %s", err.Error())
		}
	})
}

func TestGetTasks(t *testing.T) {
	db := setupTaskTestDB(t)

	task1 := Task{Name: "Incomplete Task", Completed: false, Priority: 1}
	task2 := Task{Name: "Complete Task", Completed: true, Priority: 2}
	task3 := Task{Name: "Another Task", Completed: false, Priority: 3}

	db.Create(&task1)
	db.Create(&task2)
	db.Create(&task3)

	t.Run("gets all tasks", func(t *testing.T) {
		tasks, err := GetTasks(db, nil, "", []uuid.UUID{}, "")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(tasks))
		}
	})

	t.Run("filters by completed true", func(t *testing.T) {
		completed := true
		tasks, err := GetTasks(db, &completed, "", nil, "")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) != 1 {
			t.Errorf("Expected 1 completed task, got %d", len(tasks))
		}

		if !tasks[0].Completed {
			t.Error("Expected task to be completed")
		}
	})

	t.Run("filters by completed false", func(t *testing.T) {
		completed := false
		tasks, err := GetTasks(db, &completed, "", nil, "")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) != 2 {
			t.Errorf("Expected 2 incomplete tasks, got %d", len(tasks))
		}

		for _, task := range tasks {
			if task.Completed {
				t.Error("Expected all tasks to be incomplete")
			}
		}
	})

	t.Run("filters by name", func(t *testing.T) {
		tasks, err := GetTasks(db, nil, "Complete", nil, "")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) != 2 {
			t.Errorf("Expected 2 tasks with 'Complete' in name, got %d", len(tasks))
		}
	})

	t.Run("sorts by priority", func(t *testing.T) {
		tasks, err := GetTasks(db, nil, "", nil, "priority")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) < 2 {
			t.Error("Expected at least 2 tasks for sorting test")
		}

		// Should be ordered by completed first, then priority
		incompleteCount := 0
		for _, task := range tasks {
			if !task.Completed {
				incompleteCount++
			} else {
				break
			}
		}

		if incompleteCount < 1 {
			t.Error("Expected incomplete tasks to come first")
		}
	})

	t.Run("sorts by name", func(t *testing.T) {
		tasks, err := GetTasks(db, nil, "", nil, "name")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) < 2 {
			t.Error("Expected at least 2 tasks for sorting test")
		}
	})

	t.Run("sorts by completed", func(t *testing.T) {
		tasks, err := GetTasks(db, nil, "", nil, "completed")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) < 2 {
			t.Error("Expected at least 2 tasks for sorting test")
		}
	})

	t.Run("default sorting", func(t *testing.T) {
		tasks, err := GetTasks(db, nil, "", nil, "invalid")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(tasks))
		}
	})

	t.Run("filters by tag IDs", func(t *testing.T) {
		// Create a tag and associate it with task1
		tag := Tag{Name: "work"}
		db.Create(&tag)

		// Associate tag with task1
		db.Model(&task1).Association("Tags").Append(&tag)

		// Test filtering by tag ID
		tasks, err := GetTasks(db, nil, "", []uuid.UUID{tag.ID}, "")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) != 1 {
			t.Errorf("Expected 1 task with specific tag, got %d", len(tasks))
		}

		if tasks[0].ID != task1.ID {
			t.Errorf("Expected task ID %s, got %s", task1.ID, tasks[0].ID)
		}

		// Test with non-existent tag
		nonExistentTagID := uuid.New()
		tasks, err = GetTasks(db, nil, "", []uuid.UUID{nonExistentTagID}, "")
		if err != nil {
			t.Errorf("GetTasks returned error: %v", err)
		}

		if len(tasks) != 0 {
			t.Errorf("Expected 0 tasks with non-existent tag, got %d", len(tasks))
		}
	})
}

func TestTaskJSONTags(t *testing.T) {
	task := Task{
		ID:           uuid.New(),
		Name:         "Test Task",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Completed:    true,
	}

	if task.ID == uuid.Nil {
		t.Error("Expected task to have an ID")
	}

	if task.Name == "" {
		t.Error("Expected task to have a name")
	}

	if task.DateCreated.IsZero() {
		t.Error("Expected task to have a creation date")
	}

	if task.DateModified.IsZero() {
		t.Error("Expected task to have a modification date")
	}
}

func TestTaskFrequencyUpdate(t *testing.T) {
	db := setupTaskTestDB(t)

	// Create a frequency first
	frequency := Frequency{Name: "Test Freq", Reset: "0 18 * * *"}
	db.Create(&frequency)

	// Create task with frequency
	task := Task{Name: "Test Task", FrequencyID: &frequency.ID}
	db.Create(&task)

	t.Run("remove frequency from task using Update", func(t *testing.T) {
		// Update to remove frequency
		updateData := Task{FrequencyID: nil}
		err := task.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update failed: %v", err)
		}

		// Check if frequency was removed
		if task.FrequencyID != nil {
			t.Errorf("Expected FrequencyID to be nil, got %v", task.FrequencyID)
		}

		// Reload from database
		var reloadedTask Task
		err = reloadedTask.LoadByID(db, task.ID)
		if err != nil {
			t.Errorf("LoadByID failed: %v", err)
		}

		if reloadedTask.FrequencyID != nil {
			t.Errorf("Expected reloaded FrequencyID to be nil, got %v", reloadedTask.FrequencyID)
		}
	})

}
