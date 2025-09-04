// Package components provides functionality to render HTML components using Go's html/template package
package components

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"
)

//go:embed templates/*.html
var componentTemplates embed.FS

type ComponentRenderer struct {
	templates map[string]*template.Template
}

// formatDate formats a time.Time for display with relative time for recent dates
func formatDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	// If less than a week old, show relative time
	if diff < 7*24*time.Hour && diff > 0 {
		days := int(diff.Hours() / 24)
		if days == 0 {
			hours := int(diff.Hours())
			if hours == 0 {
				minutes := int(diff.Minutes())
				if minutes == 0 {
					return "just now"
				}
				return fmt.Sprintf("%d minutes ago", minutes)
			}
			return fmt.Sprintf("%d hours ago", hours)
		}
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// For older dates, show in YYYY-MM-DD format
	return t.Format("2006-01-02")
}

// formatDateFull formats a time.Time for tooltip display
func formatDateFull(t time.Time) string {
	return t.Format("2006-01-02 3:04pm MST")
}

func raisePriority(i int) int {
	i = i - 1
	if i == 0 {
		i = 5
	}
	return i
}

func NewComponentRenderer() *ComponentRenderer {
	cr := &ComponentRenderer{templates: make(map[string]*template.Template)}

	// Create template function map
	funcMap := template.FuncMap{
		"formatDate":     formatDate,
		"formatDateFull": formatDateFull,
		"toggle":         func(b bool) bool { return !b },
		"raise":          raisePriority,
	}

	// Load all component templates together so they can reference each other
	files, _ := componentTemplates.ReadDir("templates")
	templatePaths := make([]string, 0, len(files))
	for _, file := range files {
		templatePaths = append(templatePaths, "templates/"+file.Name())
	}

	// Parse all templates together
	if len(templatePaths) > 0 {
		allTemplates := template.Must(template.New("").Funcs(funcMap).ParseFS(componentTemplates, templatePaths...))

		// Store each template by name
		for _, file := range files {
			name := strings.TrimSuffix(file.Name(), ".html")
			cr.templates[name] = allTemplates.Lookup(file.Name())
		}
	}

	return cr
}

func (cr *ComponentRenderer) Render(component string, data any) (string, error) {
	tmpl, exists := cr.templates[component]
	if !exists {
		return "", fmt.Errorf("component %s not found", component)
	}

	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}
