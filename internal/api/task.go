// Package api implements HTTP handlers for task management
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
	"github.com/jhoffmann/dailies/internal/websocket"
)

// GetTasks handles GET requests to retrieve tasks with optional filtering and sorting.
// Supports query parameters: completed (boolean), name (string for partial matching), tag_ids (comma-separated UUIDs), and sort (completed, priority, name).
func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var completedFilter *bool
	completed := r.URL.Query().Get("completed")
	if completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			completedFilter = &completedBool
		}
	}

	var tagIDFilter []uuid.UUID
	tagIDsParam := r.URL.Query().Get("tag_ids")
	if tagIDsParam != "" {
		tagIDStrings := strings.SplitSeq(tagIDsParam, ",")
		for tagIDStr := range tagIDStrings {
			tagIDStr = strings.TrimSpace(tagIDStr)
			if tagID, err := uuid.Parse(tagIDStr); err == nil {
				tagIDFilter = append(tagIDFilter, tagID)
			} else {
				logger.LoggedError(w, "Invalid tag ID format: "+tagIDStr, http.StatusBadRequest, r)
				return
			}
		}
	}

	nameFilter := r.URL.Query().Get("name")
	sortField := r.URL.Query().Get("sort")

	tasks, err := GetTasksWithFilter(completedFilter, nameFilter, tagIDFilter, sortField)
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

	task, err := GetTaskByID(taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	json.NewEncoder(w).Encode(task)
}

// CreateTask handles POST requests to create a new task.
// Requires a JSON body with a task name and optional tag_ids array and frequency_id.
func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var taskData struct {
		Name        string      `json:"name"`
		TagIDs      []uuid.UUID `json:"tag_ids"`
		FrequencyID *uuid.UUID  `json:"frequency_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	task, err := CreateTaskWithTagsAndFrequency(taskData.Name, taskData.TagIDs, taskData.FrequencyID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusBadRequest, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)

	// Broadcast task list refresh for new task
	websocket.BroadcastTaskListRefresh()
}

// CreateTaskWithTagsAndFrequency creates a new task with associated tags and optional frequency.
// This function contains the business logic for task creation with tag and frequency associations.
func CreateTaskWithTagsAndFrequency(name string, tagIDs []uuid.UUID, frequencyID *uuid.UUID) (*models.Task, error) {
	if name == "" {
		return nil, errors.New("task name is required")
	}

	task := models.Task{
		Name:        name,
		FrequencyID: frequencyID,
	}

	// Validate frequency exists if provided
	if frequencyID != nil {
		var frequency models.Frequency
		if err := frequency.LoadByID(database.GetDB(), *frequencyID); err != nil {
			return nil, errors.New("frequency not found")
		}
	}

	// Load associated tags if any were selected
	if len(tagIDs) > 0 {
		var tags []models.Tag
		if err := database.GetDB().Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
			return nil, errors.New("failed to load tags")
		}

		// Validate that all requested tags exist
		if len(tags) != len(tagIDs) {
			return nil, errors.New("one or more tags not found")
		}

		task.Tags = tags
	}

	err := task.Create(database.GetDB())
	if err != nil {
		return nil, err
	}

	// Reload task with tags and frequency for the response
	err = task.LoadByID(database.GetDB(), task.ID)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// CreateTaskWithTags creates a new task with associated tags (legacy function for backward compatibility).
// This function contains the business logic for task creation with tag associations.
func CreateTaskWithTags(name string, tagIDs []uuid.UUID) (*models.Task, error) {
	return CreateTaskWithTagsAndFrequency(name, tagIDs, nil)
}

// GetTasksWithFilter retrieves tasks with optional filtering and sorting.
// This function contains the business logic for task retrieval.
func GetTasksWithFilter(completedFilter *bool, nameFilter string, tagIDFilter []uuid.UUID, sortField string) ([]models.Task, error) {
	return models.GetTasks(database.GetDB(), completedFilter, nameFilter, tagIDFilter, sortField)
}

// GetTaskByID retrieves a single task by ID.
// This function contains the business logic for single task retrieval.
func GetTaskByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := task.LoadByID(database.GetDB(), taskID)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTaskByID updates an existing task by ID.
// This function contains the business logic for task updates.
func UpdateTaskByID(taskID uuid.UUID, updateData *models.Task) (*models.Task, error) {
	var task models.Task
	err := task.LoadByID(database.GetDB(), taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.New("task not found")
		}
		return nil, err
	}

	// Validate frequency exists if provided in update data
	if updateData.FrequencyID != nil {
		var frequency models.Frequency
		if err := frequency.LoadByID(database.GetDB(), *updateData.FrequencyID); err != nil {
			return nil, errors.New("frequency not found")
		}
	}

	err = task.Update(database.GetDB(), updateData)
	if err != nil {
		return nil, err
	}

	// Reload to get updated frequency relationship
	err = task.LoadByID(database.GetDB(), task.ID)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// DeleteTaskByID deletes a task by ID.
// This function contains the business logic for task deletion.
func DeleteTaskByID(taskID uuid.UUID) error {
	var task models.Task
	task.ID = taskID
	return task.Delete(database.GetDB())
}

// UpdateTask handles PUT requests to update an existing task by ID.
// Accepts a JSON body with fields to update (name, completed status, priority, and frequency_id).
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/tasks/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	// Use a custom struct to handle the JSON properly
	var requestData struct {
		Name        string     `json:"name,omitempty"`
		Completed   *bool      `json:"completed,omitempty"`
		Priority    *int       `json:"priority,omitempty"`
		FrequencyID *uuid.UUID `json:"frequency_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	// Convert to Task struct for update - be explicit about field handling
	updateData := models.Task{}

	if requestData.Name != "" {
		updateData.Name = requestData.Name
	}
	if requestData.Completed != nil {
		updateData.Completed = *requestData.Completed
	}
	if requestData.Priority != nil {
		updateData.Priority = *requestData.Priority
	}

	// Handle frequency update explicitly
	updateData.FrequencyID = requestData.FrequencyID

	task, err := UpdateTaskByID(taskID, &updateData)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		} else {
			logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		}
		return
	}

	// Broadcast WebSocket notification if task completion status changed
	if requestData.Completed != nil {
		websocket.BroadcastTaskUpdated(task.ID.String(), task.Name, *requestData.Completed)
		websocket.BroadcastTaskListRefresh()
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

	err = DeleteTaskByID(taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	// Broadcast task list refresh for deleted task
	websocket.BroadcastTaskListRefresh()

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
