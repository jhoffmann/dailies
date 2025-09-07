// Package web contains HTTP handlers for serving HTML templates and components
package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
)

// GetTasksHTML returns HTML snippet for task list (for HTMX).
// Supports query parameters: completed (boolean), name (string for partial matching), tag_ids (comma-separated or multiple), and sort (completed, priority, name).
func GetTasksHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var completedFilter *bool
	completed := r.URL.Query().Get("completed")
	if completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			completedFilter = &completedBool
		}
	}

	var tagIDFilter []uuid.UUID
	// Handle multiple tag_ids parameters from checkboxes
	tagIDs := r.URL.Query()["tag_ids"]
	for _, tagIDStr := range tagIDs {
		if tagID, err := uuid.Parse(strings.TrimSpace(tagIDStr)); err == nil {
			tagIDFilter = append(tagIDFilter, tagID)
		}
	}

	nameFilter := r.URL.Query().Get("name")
	sortField := r.URL.Query().Get("sort")

	// Use the API layer for business logic
	tasks, err := api.GetTasksWithFilter(completedFilter, nameFilter, tagIDFilter, sortField)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	html, err := componentRenderer.Render("taskList", tasks)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}

// GetTaskHTML returns HTML snippet for single task view.
func GetTaskHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	id := r.URL.Path[len("/component/tasks/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	// Use the API layer for business logic
	task, err := api.GetTaskByID(taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	html, err := componentRenderer.Render("taskView", task)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}

// GetTaskEditHTML returns HTML form for editing a task.
func GetTaskEditHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	path := r.URL.Path
	idEnd := len(path) - len("/edit")
	id := path[len("/component/tasks/"):idEnd]

	taskID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	// Use the API layer for business logic
	task, err := api.GetTaskByID(taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	html, err := componentRenderer.Render("taskEdit", task)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}

// DeleteTaskHTML handles DELETE requests for HTMX task deletion.
// Returns empty content to remove the task from DOM.
func DeleteTaskHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	path := r.URL.Path
	id := path[len("/component/tasks/"):]

	taskID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	// Use the API layer for business logic
	err = api.DeleteTaskByID(taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

// UpdateTaskHTML handles PUT requests for HTMX task updates.
// Returns updated taskView HTML snippet.
func UpdateTaskHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	path := r.URL.Path
	id := path[len("/component/tasks/"):]

	taskID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	var updateData models.Task
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	// Use the API layer for business logic - no tag updates from web UI currently
	task, err := api.UpdateTaskByID(taskID, &updateData, nil)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	html, err := componentRenderer.Render("taskView", task)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}

// GetTaskCreateHTML returns HTML form for creating a new task.
func GetTaskCreateHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html, err := componentRenderer.Render("taskCreate", nil)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}

// CreateTaskHTML handles POST requests to create a new task and return taskView HTML.
func CreateTaskHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var taskData struct {
		Name        string     `json:"name"`
		TagIDs      any        `json:"tag_ids"`
		FrequencyID *uuid.UUID `json:"frequency_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		logger.LoggedError(w, "Invalid JSON", http.StatusBadRequest, r)
		return
	}

	// Handle tag_ids as either string or array
	var tagIDs []uuid.UUID
	if taskData.TagIDs != nil {
		switch v := taskData.TagIDs.(type) {
		case string:
			// Single tag ID as string
			if tagID, err := uuid.Parse(v); err == nil {
				tagIDs = []uuid.UUID{tagID}
			}
		case []any:
			// Array of tag IDs
			for _, tagIDInterface := range v {
				if tagIDStr, ok := tagIDInterface.(string); ok {
					if tagID, err := uuid.Parse(tagIDStr); err == nil {
						tagIDs = append(tagIDs, tagID)
					}
				}
			}
		}
	}

	// Use the API layer for business logic
	task, err := api.CreateTaskWithTagsAndFrequency(taskData.Name, tagIDs, taskData.FrequencyID, nil)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusBadRequest, r)
		return
	}

	html, err := componentRenderer.Render("taskView", task)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(html))
}

// GetTaskConfirmDeleteHTML returns HTML confirmation modal for deleting a task.
func GetTaskConfirmDeleteHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	id := r.URL.Path[len("/component/delete/task/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		logger.LoggedError(w, "Invalid task ID", http.StatusBadRequest, r)
		return
	}

	// Use the API layer for business logic
	task, err := api.GetTaskByID(taskID)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusNotFound, r)
		return
	}

	html, err := componentRenderer.Render("taskConfirmDelete", task)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}
