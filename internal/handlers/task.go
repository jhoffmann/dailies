package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/models"
)

// GetTasks handles GET requests to retrieve tasks with optional filtering.
// Supports query parameters: completed (boolean) and name (string for partial matching).
func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var tasks []models.Task
	query := database.GetDB()

	completed := r.URL.Query().Get("completed")
	if completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			query = query.Where("completed = ?", completedBool)
		}
	}

	name := r.URL.Query().Get("name")
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	result := query.Find(&tasks)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

// GetTask handles GET requests to retrieve a single task by ID.
func GetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/tasks/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	result := database.GetDB().First(&task, "id = ?", taskID)
	if result.Error != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(task)
}

// CreateTask handles POST requests to create a new task.
// Requires a JSON body with a task name.
func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.Name == "" {
		http.Error(w, "Task name is required", http.StatusBadRequest)
		return
	}

	result := database.GetDB().Create(&task)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// UpdateTask handles PUT requests to update an existing task by ID.
// Accepts a JSON body with fields to update (name and/or completed status).
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/tasks/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var existingTask models.Task
	result := database.GetDB().First(&existingTask, "id = ?", taskID)
	if result.Error != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var updateData models.Task
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if updateData.Name != "" {
		existingTask.Name = updateData.Name
	}
	existingTask.Completed = updateData.Completed

	result = database.GetDB().Save(&existingTask)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existingTask)
}

// DeleteTask handles DELETE requests to remove a task by ID.
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/tasks/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	result := database.GetDB().Delete(&models.Task{}, "id = ?", taskID)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
