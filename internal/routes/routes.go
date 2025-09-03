package routes

import (
	"net/http"
	"strings"

	"github.com/jhoffmann/dailies/internal/handlers"
)

// Setup configures HTTP routes for the application.
// Registers handlers for static files, root path, and task management API endpoints.
func Setup() {
	// Static file server for CSS and JS
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Root path serves the main HTML template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handlers.ServeIndex(w, r)
	})

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTasks(w, r)
		case http.MethodPost:
			handlers.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/tasks/") == "" {
			http.Error(w, "Task ID is required", http.StatusBadRequest)
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
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
