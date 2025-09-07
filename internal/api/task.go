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
//
//	@Summary		List tasks
//	@Description	Get tasks with optional filtering and sorting
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			completed	query		boolean	false	"Filter by completion status"
//	@Param			name		query		string	false	"Filter by task name (partial matching)"
//	@Param			tag_ids		query		string	false	"Filter by tag IDs (comma-separated)"
//	@Param			sort		query		string	false	"Sort field: completed, priority, name"
//	@Success		200			{array}		models.Task
//	@Failure		400			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/tasks [get]
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
//
//	@Summary		Get task by ID
//	@Description	Get a single task by its ID
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Task ID"
//	@Success		200	{object}	models.Task
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/tasks/{id} [get]
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
//
//	@Summary		Create a new task
//	@Description	Create a new task with optional tags and frequency
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			task	body		object{name=string,tag_ids=[]string,frequency_id=string,priority=integer}	true	"Task data"
//	@Success		201		{object}	models.Task
//	@Failure		400		{object}	map[string]string
//	@Router			/tasks [post]
func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var taskData struct {
		Name        string      `json:"name"`
		TagIDs      []uuid.UUID `json:"tag_ids"`
		FrequencyID *uuid.UUID  `json:"frequency_id"`
		Priority    *int        `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	task, err := CreateTaskWithTagsAndFrequency(taskData.Name, taskData.TagIDs, taskData.FrequencyID, taskData.Priority)
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
func CreateTaskWithTagsAndFrequency(name string, tagIDs []uuid.UUID, frequencyID *uuid.UUID, priority *int) (*models.Task, error) {
	if name == "" {
		return nil, errors.New("task name is required")
	}

	task := models.Task{
		Name:        name,
		FrequencyID: frequencyID,
	}

	// Set priority if provided
	if priority != nil {
		task.Priority = *priority
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
	return CreateTaskWithTagsAndFrequency(name, tagIDs, nil, nil)
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
func UpdateTaskByID(taskID uuid.UUID, updateData *models.Task, tagIDs *[]uuid.UUID) (*models.Task, error) {
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

	// Update tag associations if provided
	if tagIDs != nil {
		err = task.UpdateTags(database.GetDB(), *tagIDs)
		if err != nil {
			return nil, err
		}
	}

	// Reload to get updated relationships
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
//
//	@Summary		Update task
//	@Description	Update an existing task by ID
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string																		true	"Task ID"
//	@Param			task	body		object{name=string,completed=boolean,priority=integer,frequency_id=string,tag_ids=[]string}	true	"Task update data"
//	@Success		200		{object}	models.Task
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/tasks/{id} [put]
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
		Name        string      `json:"name,omitempty"`
		Completed   *bool       `json:"completed,omitempty"`
		Priority    *int        `json:"priority,omitempty"`
		FrequencyID *uuid.UUID  `json:"frequency_id"`
		TagIDs      []uuid.UUID `json:"tag_ids,omitempty"`
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

	// Handle tag updates - pass pointer to tagIDs if provided, nil otherwise
	var tagIDsPtr *[]uuid.UUID
	if requestData.TagIDs != nil {
		tagIDsPtr = &requestData.TagIDs
	}

	task, err := UpdateTaskByID(taskID, &updateData, tagIDsPtr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		} else {
			logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		}
		return
	}

	// Broadcast WebSocket notification
	websocket.BroadcastTaskUpdated(task.ID.String(), task.Name, *requestData.Completed)
	websocket.BroadcastTaskListRefresh()

	json.NewEncoder(w).Encode(task)
}

// DeleteTask handles DELETE requests to remove a task by ID.
//
//	@Summary		Delete task
//	@Description	Delete a task by ID
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Task ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/tasks/{id} [delete]
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
//
//	@Summary		Health check
//	@Description	Check application and database health
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/healthz [get]
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
