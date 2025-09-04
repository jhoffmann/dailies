// Package api implements HTTP handlers for task management
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
)

// GetTasks handles GET requests to retrieve tasks with optional filtering and sorting.
// Supports query parameters: completed (boolean), name (string for partial matching), and sort (completed, priority, name).
func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var completedFilter *bool
	completed := r.URL.Query().Get("completed")
	if completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			completedFilter = &completedBool
		}
	}

	nameFilter := r.URL.Query().Get("name")
	sortField := r.URL.Query().Get("sort")

	tasks, err := models.GetTasks(database.GetDB(), completedFilter, nameFilter, sortField)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
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
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	var task models.Task
	err = task.LoadByID(database.GetDB(), taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
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
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	err := task.Create(database.GetDB())
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusBadRequest, r)
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
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	var task models.Task
	err = task.LoadByID(database.GetDB(), taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	var updateData models.Task
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	err = task.Update(database.GetDB(), &updateData)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	json.NewEncoder(w).Encode(task)
}

// DeleteTask handles DELETE requests to remove a task by ID.
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/tasks/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	var task models.Task
	task.ID = taskID
	err = task.Delete(database.GetDB())
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck handles GET requests to check application health.
// Returns HTTP 200 with JSON body {'health': 'Ok'} if database connection is working.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := database.GetDB()
	if db == nil {
		logger.LoggedError(w, "Database not initialized", http.StatusInternalServerError, r)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.LoggedError(w, "Failed to get database connection", http.StatusInternalServerError, r)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		logger.LoggedError(w, "Database connection failed", http.StatusInternalServerError, r)
		return
	}

	response := map[string]string{"status": "UP"}
	json.NewEncoder(w).Encode(response)
}
