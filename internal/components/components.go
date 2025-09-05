// Package components provides functionality to render HTML components using Go's html/template package
package components

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
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
				if minutes == 1 {
					return "1 minute ago"
				}
				return fmt.Sprintf("%d minutes ago", minutes)
			}
			if hours == 1 {
				return "1 hour ago"
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

// raisePriority decrements a priority value by 1, wrapping from 1 to 5.
// Used for cycling task priorities in the UI when clicking the priority badge.
func raisePriority(i int) int {
	i = i - 1
	if i == 0 {
		i = 5
	}
	return i
}

// timeUntilReset calculates the time remaining until the next occurrence of a cron expression.
// Takes a standard cron expression (minute hour day month day-of-week) and returns a human-readable
// string like "2 days", "4 hours", or "30 minutes". Returns empty string for invalid expressions.
func timeUntilReset(cronExpr string) string {
	if cronExpr == "" {
		return ""
	}

	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return ""
	}

	now := time.Now()
	nextReset := schedule.Next(now)
	duration := nextReset.Sub(now)

	if duration < 0 {
		return ""
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}

	if hours > 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}

	if minutes > 0 {
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}

	return "< 1 minute"
}

// NewComponentRenderer creates and initializes a new ComponentRenderer with all HTML templates loaded.
// It sets up template functions and parses all component templates from the embedded filesystem.
func NewComponentRenderer() *ComponentRenderer {
	cr := &ComponentRenderer{templates: make(map[string]*template.Template)}

	// Create template function map
	funcMap := template.FuncMap{
		"formatDate":     formatDate,
		"formatDateFull": formatDateFull,
		"toggle":         func(b bool) bool { return !b },
		"raise":          raisePriority,
		"timeUntilReset": timeUntilReset,
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

// Render executes a component template with the provided data and returns the resulting HTML string.
// Returns an error if the component template is not found or template execution fails.
func (cr *ComponentRenderer) Render(component string, data any) (string, error) {
	tmpl, exists := cr.templates[component]
	if !exists {
		return "", fmt.Errorf("component %s not found", component)
	}

	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}
