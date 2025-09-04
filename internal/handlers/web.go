package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/components"
	"github.com/jhoffmann/dailies/internal/database"
	"github.com/jhoffmann/dailies/internal/models"
)

var componentRenderer = components.NewComponentRenderer()

// ServeIndex handles GET requests to serve the main HTML template.
// Renders the index.html template for the single-page application.
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// GetTasksHTML returns HTML snippet for task list (for HTMX).
// Supports query parameters: completed (boolean) and name (string for partial matching).
func GetTasksHTML(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetTasksHTML called: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html")

	var completedFilter *bool
	completed := r.URL.Query().Get("completed")
	if completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			completedFilter = &completedBool
		}
	}

	nameFilter := r.URL.Query().Get("name")

	tasks, err := models.GetTasks(database.GetDB(), completedFilter, nameFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html, err := componentRenderer.Render("taskList", tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(html))
}

// GetTaskHTML returns HTML snippet for single task view.
func GetTaskHTML(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetTaskHTML called: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html")

	id := r.URL.Path[len("/component/tasks/"):]
	taskID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	err = task.LoadByID(database.GetDB(), taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	html, err := componentRenderer.Render("taskView", task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(html))
}

// GetTaskEditHTML returns HTML form for editing a task.
func GetTaskEditHTML(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetTaskEditHTML called: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/html")

	path := r.URL.Path
	idEnd := len(path) - len("/edit")
	id := path[len("/component/tasks/"):idEnd]

	taskID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	err = task.LoadByID(database.GetDB(), taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	html, err := componentRenderer.Render("taskEdit", task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(html))
}
