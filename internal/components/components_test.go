package components

import (
	"strings"
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "just now",
			input:    now,
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			input:    now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			input:    now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			input:    now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			input:    now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			input:    now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "3 days ago",
			input:    now.Add(-3 * 24 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "6 days ago",
			input:    now.Add(-6 * 24 * time.Hour),
			expected: "6 days ago",
		},
		{
			name:     "1 week ago - shows date",
			input:    now.Add(-7 * 24 * time.Hour),
			expected: now.Add(-7 * 24 * time.Hour).Format("2006-01-02"),
		},
		{
			name:     "future date - shows formatted date",
			input:    now.Add(24 * time.Hour),
			expected: now.Add(24 * time.Hour).Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDate(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatDateFull(t *testing.T) {
	// Test with a known date
	testTime := time.Date(2025, 1, 15, 14, 30, 45, 0, time.UTC)

	result := formatDateFull(testTime)
	expected := "2025-01-15 2:30pm UTC"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRaisePriority(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "priority 5 wraps to 4",
			input:    5,
			expected: 4,
		},
		{
			name:     "priority 4 becomes 3",
			input:    4,
			expected: 3,
		},
		{
			name:     "priority 3 becomes 2",
			input:    3,
			expected: 2,
		},
		{
			name:     "priority 2 becomes 1",
			input:    2,
			expected: 1,
		},
		{
			name:     "priority 1 wraps to 5",
			input:    1,
			expected: 5,
		},
		{
			name:     "priority 0 wraps to 5 (edge case)",
			input:    0,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := raisePriority(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestTimeUntilReset(t *testing.T) {
	tests := []struct {
		name     string
		cronExpr string
		contains string // Part of expected string since exact timing varies
		isEmpty  bool
	}{
		{
			name:     "empty cron expression",
			cronExpr: "",
			isEmpty:  true,
		},
		{
			name:     "invalid cron expression",
			cronExpr: "invalid",
			isEmpty:  true,
		},
		{
			name:     "daily at midnight",
			cronExpr: "0 0 * * *",
			contains: "", // Will contain "hour" or "minute" - varies by test time
			isEmpty:  false,
		},
		{
			name:     "every minute (for testing)",
			cronExpr: "* * * * *",
			contains: "minute",
			isEmpty:  false,
		},
		{
			name:     "weekly on Sunday",
			cronExpr: "0 0 * * 0",
			contains: "", // Will contain "day", "hour", or "minute"
			isEmpty:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeUntilReset(tt.cronExpr)

			if tt.isEmpty {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
			} else {
				if result == "" {
					t.Error("expected non-empty string, got empty")
				}
				if tt.contains != "" && !strings.Contains(result, tt.contains) {
					t.Errorf("expected result to contain %q, got %q", tt.contains, result)
				}
				// Verify the result matches expected patterns
				validPatterns := []string{
					"< 1 minute",
					"1 minute",
					"minutes",
					"1 hour",
					"hours",
					"1 day",
					"days",
				}
				matched := false
				for _, pattern := range validPatterns {
					if strings.Contains(result, pattern) || result == pattern {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("result %q doesn't match expected patterns", result)
				}
			}
		})
	}
}

func TestNewComponentRenderer(t *testing.T) {
	cr := NewComponentRenderer()

	if cr == nil {
		t.Fatal("expected non-nil ComponentRenderer")
	}

	if cr.templates == nil {
		t.Fatal("expected non-nil templates map")
	}

	// Check that some templates are loaded
	if len(cr.templates) == 0 {
		t.Error("expected templates to be loaded")
	}

	// Check that specific templates exist (based on the files we saw)
	expectedTemplates := []string{
		"taskList",
		"taskView",
		"taskCreate",
		"taskEdit",
		"tagCreate",
		"frequencySelection",
	}

	for _, templateName := range expectedTemplates {
		if _, exists := cr.templates[templateName]; !exists {
			t.Errorf("expected template %q to be loaded", templateName)
		}
	}
}

func TestComponentRenderer_Render(t *testing.T) {
	cr := NewComponentRenderer()

	t.Run("non-existent component", func(t *testing.T) {
		_, err := cr.Render("nonExistentComponent", nil)
		if err == nil {
			t.Error("expected error for non-existent component")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected 'not found' in error message, got: %v", err)
		}
	})

	// Note: Testing actual template rendering would require knowing the template structure
	// and providing appropriate data. Since templates can be complex and change frequently,
	// we focus on testing the error cases and basic functionality here.

	t.Run("template exists check", func(t *testing.T) {
		// Just verify that if a template exists, Render doesn't immediately error
		// We can't easily test successful rendering without mock templates or knowing exact structure
		if len(cr.templates) > 0 {
			// Get any template name
			var templateName string
			for name := range cr.templates {
				templateName = name
				break
			}

			// Call render - it might error due to template expecting specific data structure,
			// but it shouldn't error due to template not being found
			_, err := cr.Render(templateName, nil)
			if err != nil && strings.Contains(err.Error(), "not found") {
				t.Errorf("template %q should exist but got 'not found' error", templateName)
			}
		}
	})
}
