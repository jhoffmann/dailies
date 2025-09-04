// Package routes sets up HTTP routes for the web application
package routes

import (
	"net/http"
	"strings"

	"github.com/jhoffmann/dailies/internal/api"
	"github.com/jhoffmann/dailies/internal/logger"
	"github.com/jhoffmann/dailies/internal/ui/web"
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
		web.ServeIndex(w, r)
	}))

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

	// Component routes for HTMX HTML snippets
	http.HandleFunc("/component/tasks", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
			return
		}
		web.GetTasksHTML(w, r)
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

	http.HandleFunc("/tags", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.GetTags(w, r)
		case http.MethodPost:
			api.CreateTag(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))

	http.HandleFunc("/tags/", LogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/tags/") == "" {
			logger.LoggedError(w, "Tag ID is required", http.StatusBadRequest, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.GetTag(w, r)
		case http.MethodPut:
			api.UpdateTag(w, r)
		case http.MethodDelete:
			api.DeleteTag(w, r)
		default:
			logger.LoggedError(w, "Method not allowed", http.StatusMethodNotAllowed, r)
		}
	}))
}
