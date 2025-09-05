// Package routes sets up HTTP routes for the web application
package routes

import (
	"net/http"
	"strings"

	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/ui/web"
)

// SetupTaskRoutes configures HTTP routes for task management.
// Registers both API endpoints and HTML component endpoints for tasks.
func SetupTaskRoutes() {
	// Task API endpoints
	http.HandleFunc("/tasks", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetTasks(w, r)
		case http.MethodPost:
			api.CreateTask(w, r)
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
			api.GetTask(w, r)
		case http.MethodPut:
			api.UpdateTask(w, r)
		case http.MethodDelete:
			api.DeleteTask(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	// Task component routes for HTMX HTML snippets
	http.HandleFunc("/component/tasks", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}
		web.GetTasksHTML(w, r)
	}))

	http.HandleFunc("/component/create/task", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			web.GetTaskCreateHTML(w, r)
		case http.MethodPost:
			web.CreateTaskHTML(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	http.HandleFunc("/component/delete/task/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}
		web.GetTaskConfirmDeleteHTML(w, r)
	}))

	http.HandleFunc("/component/tasks/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/component/tasks/")
		if path == "" {
			logger.LoggedError(w, "Task ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			// Check if this is an edit request
			if strings.HasSuffix(path, "/edit") {
				web.GetTaskEditHTML(w, r)
			} else {
				web.GetTaskHTML(w, r)
			}
		case http.MethodPut:
			web.UpdateTaskHTML(w, r)
		case http.MethodDelete:
			web.DeleteTaskHTML(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))
}
