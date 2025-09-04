// Package web contains HTTP handlers for serving HTML templates and components
package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/components"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/models"
)

var componentRenderer = components.NewComponentRenderer()

// ServeIndex handles GET requests to serve the main HTML template.
// Renders the index.html template for the single-page application.
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		logger.LoggedError(w, "Error loading template", http.StatusInternalServerError, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, nil); err != nil {
		logger.LoggedError(w, "Error rendering template", http.StatusInternalServerError, r)
		return
	}
}

// GetTasksHTML returns HTML snippet for task list (for HTMX).
// Supports query parameters: completed (boolean), name (string for partial matching), and sort (completed, priority, name).
func GetTasksHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var completedFilter *bool
	completed := r.URL.Query().Get("completed")
	if completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			completedFilter = &completedBool
		}
	}

	nameFilter := r.URL.Query().Get("name")
	sortField := r.URL.Query().Get("sort")
	if sortField == "" {
		sortField = "priority"
	}

	tasks, err := models.GetTasks(database.GetDB(), completedFilter, nameFilter, sortField)
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

	var task models.Task
	err = task.LoadByID(database.GetDB(), taskID)
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

	var task models.Task
	err = task.LoadByID(database.GetDB(), taskID)
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

	var task models.Task
	task.ID = taskID
	err = task.Delete(database.GetDB())
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

	html, err := componentRenderer.Render("taskView", task)
	if err != nil {
		logger.LoggedError(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	w.Write([]byte(html))
}
