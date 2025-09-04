// Package routes sets up HTTP routes for the web application
package routes

import (
	"net/http"
	"strings"

	"github.com/jhoffmann/dailies/internal/handlers"
	"github.com/jhoffmann/dailies/internal/logger"
)

// Setup configures HTTP routes for the application.
// Registers handlers for static files, root path, and task management API endpoints.
func Setup() {
	// Static file server for CSS and JS
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Root path serves the main HTML template
	http.HandleFunc("/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			logger.LoggedError(w, "Not Found", http.StatusNotFound, r)
			return
		}
		handlers.ServeIndex(w, r)
	}))

	http.HandleFunc("/tasks", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTasks(w, r)
		case http.MethodPost:
			handlers.CreateTask(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	http.HandleFunc("/tasks/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/tasks/") == "" {
			logger.LoggedError(w, "Task ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handlers.GetTask(w, r)
		case http.MethodPut:
			handlers.UpdateTask(w, r)
		case http.MethodDelete:
			handlers.DeleteTask(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	// Component routes for HTMX HTML snippets
	http.HandleFunc("/component/tasks", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}
		handlers.GetTasksHTML(w, r)
	}))

	http.HandleFunc("/component/tasks/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/component/tasks/")
		if path == "" {
			logger.LoggedError(w, "Task ID is required", http.StatusBadRequest, r)
			return
		}

		// Handle DELETE requests for HTMX
		if r.Method == http.MethodDelete && strings.HasSuffix(path, "/delete") {
			handlers.DeleteTaskHTML(w, r)
			return
		}

		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}

		// Check if this is an edit request
		if strings.HasSuffix(path, "/edit") {
			handlers.GetTaskEditHTML(w, r)
		} else {
			handlers.GetTaskHTML(w, r)
		}
	}))
}
