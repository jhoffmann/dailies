package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jhoffmann/dailies/internal/models"
)

func TestParseHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "hostname only - adds default port",
			input:    "api.example.com",
			expected: "http://api.example.com:9001",
			hasError: false,
		},
		{
			name:     "port only - adds localhost",
			input:    ":8080",
			expected: "http://localhost:8080",
			hasError: false,
		},
		{
			name:     "hostname and port",
			input:    "localhost:9001",
			expected: "http://localhost:9001",
			hasError: false,
		},
		{
			name:     "custom hostname and port",
			input:    "example.com:3000",
			expected: "http://example.com:3000",
			hasError: false,
		},
		{
			name:     "localhost only - adds default port",
			input:    "localhost",
			expected: "http://localhost:9001",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseHost(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewDailiesClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "valid URL",
			baseURL: "http://localhost:9001",
		},
		{
			name:    "different URL",
			baseURL: "https://api.example.com:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewDailiesClient(tt.baseURL)

			if client == nil {
				t.Error("expected non-nil client")
			}
			if client.baseURL != tt.baseURL {
				t.Errorf("expected baseURL %q, got %q", tt.baseURL, client.baseURL)
			}
			if client.client == nil {
				t.Error("expected non-nil HTTP client")
			}
		})
	}
}

func TestTaskToResponse(t *testing.T) {
	// Create test UUIDs
	taskID := uuid.New()
	freqID := uuid.New()
	tagID1 := uuid.New()
	tagID2 := uuid.New()

	// Create test timestamps
	created := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	modified := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     models.Task
		expected TaskResponse
	}{
		{
			name: "basic task without frequency or tags",
			task: models.Task{
				ID:           taskID,
				Name:         "Test Task",
				DateCreated:  created,
				DateModified: modified,
				Completed:    false,
				Priority:     3,
			},
			expected: TaskResponse{
				ID:           taskID.String(),
				Name:         "Test Task",
				DateCreated:  created.Format(time.RFC3339),
				DateModified: modified.Format(time.RFC3339),
				Completed:    false,
				Priority:     3,
			},
		},
		{
			name: "task with frequency and tags",
			task: models.Task{
				ID:           taskID,
				Name:         "Complex Task",
				DateCreated:  created,
				DateModified: modified,
				Completed:    true,
				Priority:     5,
				FrequencyID:  &freqID,
				Frequency: &models.Frequency{
					ID:    freqID,
					Name:  "Daily",
					Reset: "0 0 * * *",
				},
				Tags: []models.Tag{
					{
						ID:    tagID1,
						Name:  "work",
						Color: "#ff0000",
					},
					{
						ID:    tagID2,
						Name:  "urgent",
						Color: "#00ff00",
					},
				},
			},
			expected: TaskResponse{
				ID:           taskID.String(),
				Name:         "Complex Task",
				DateCreated:  created.Format(time.RFC3339),
				DateModified: modified.Format(time.RFC3339),
				Completed:    true,
				Priority:     5,
				FrequencyID:  stringPtr(freqID.String()),
				Frequency: &FrequencyResponse{
					ID:    freqID.String(),
					Name:  "Daily",
					Reset: "0 0 * * *",
				},
				Tags: []TagResponse{
					{
						ID:    tagID1.String(),
						Name:  "work",
						Color: "#ff0000",
					},
					{
						ID:    tagID2.String(),
						Name:  "urgent",
						Color: "#00ff00",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := taskToResponse(tt.task)

			if result.ID != tt.expected.ID {
				t.Errorf("ID: expected %q, got %q", tt.expected.ID, result.ID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name: expected %q, got %q", tt.expected.Name, result.Name)
			}
			if result.DateCreated != tt.expected.DateCreated {
				t.Errorf("DateCreated: expected %q, got %q", tt.expected.DateCreated, result.DateCreated)
			}
			if result.DateModified != tt.expected.DateModified {
				t.Errorf("DateModified: expected %q, got %q", tt.expected.DateModified, result.DateModified)
			}
			if result.Completed != tt.expected.Completed {
				t.Errorf("Completed: expected %v, got %v", tt.expected.Completed, result.Completed)
			}
			if result.Priority != tt.expected.Priority {
				t.Errorf("Priority: expected %d, got %d", tt.expected.Priority, result.Priority)
			}

			// Check frequency ID
			if tt.expected.FrequencyID == nil && result.FrequencyID != nil {
				t.Error("FrequencyID: expected nil, got non-nil")
			} else if tt.expected.FrequencyID != nil && result.FrequencyID == nil {
				t.Error("FrequencyID: expected non-nil, got nil")
			} else if tt.expected.FrequencyID != nil && result.FrequencyID != nil && *result.FrequencyID != *tt.expected.FrequencyID {
				t.Errorf("FrequencyID: expected %q, got %q", *tt.expected.FrequencyID, *result.FrequencyID)
			}

			// Check frequency
			if tt.expected.Frequency == nil && result.Frequency != nil {
				t.Error("Frequency: expected nil, got non-nil")
			} else if tt.expected.Frequency != nil && result.Frequency == nil {
				t.Error("Frequency: expected non-nil, got nil")
			} else if tt.expected.Frequency != nil && result.Frequency != nil {
				if result.Frequency.ID != tt.expected.Frequency.ID {
					t.Errorf("Frequency.ID: expected %q, got %q", tt.expected.Frequency.ID, result.Frequency.ID)
				}
				if result.Frequency.Name != tt.expected.Frequency.Name {
					t.Errorf("Frequency.Name: expected %q, got %q", tt.expected.Frequency.Name, result.Frequency.Name)
				}
				if result.Frequency.Reset != tt.expected.Frequency.Reset {
					t.Errorf("Frequency.Reset: expected %q, got %q", tt.expected.Frequency.Reset, result.Frequency.Reset)
				}
			}

			// Check tags
			if len(result.Tags) != len(tt.expected.Tags) {
				t.Errorf("Tags length: expected %d, got %d", len(tt.expected.Tags), len(result.Tags))
			} else {
				for i, expectedTag := range tt.expected.Tags {
					if i >= len(result.Tags) {
						break
					}
					resultTag := result.Tags[i]
					if resultTag.ID != expectedTag.ID {
						t.Errorf("Tag[%d].ID: expected %q, got %q", i, expectedTag.ID, resultTag.ID)
					}
					if resultTag.Name != expectedTag.Name {
						t.Errorf("Tag[%d].Name: expected %q, got %q", i, expectedTag.Name, resultTag.Name)
					}
					if resultTag.Color != expectedTag.Color {
						t.Errorf("Tag[%d].Color: expected %q, got %q", i, expectedTag.Color, resultTag.Color)
					}
				}
			}
		})
	}
}

func TestTagToResponse(t *testing.T) {
	tagID := uuid.New()

	tests := []struct {
		name     string
		tag      models.Tag
		expected TagResponse
	}{
		{
			name: "basic tag",
			tag: models.Tag{
				ID:    tagID,
				Name:  "work",
				Color: "#ff0000",
			},
			expected: TagResponse{
				ID:    tagID.String(),
				Name:  "work",
				Color: "#ff0000",
			},
		},
		{
			name: "tag with different color",
			tag: models.Tag{
				ID:    tagID,
				Name:  "personal",
				Color: "#00ff00",
			},
			expected: TagResponse{
				ID:    tagID.String(),
				Name:  "personal",
				Color: "#00ff00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tagToResponse(tt.tag)

			if result.ID != tt.expected.ID {
				t.Errorf("ID: expected %q, got %q", tt.expected.ID, result.ID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name: expected %q, got %q", tt.expected.Name, result.Name)
			}
			if result.Color != tt.expected.Color {
				t.Errorf("Color: expected %q, got %q", tt.expected.Color, result.Color)
			}
		})
	}
}

func TestFrequencyToResponse(t *testing.T) {
	freqID := uuid.New()

	tests := []struct {
		name     string
		freq     models.Frequency
		expected FrequencyResponse
	}{
		{
			name: "daily frequency",
			freq: models.Frequency{
				ID:    freqID,
				Name:  "Daily",
				Reset: "0 0 * * *",
			},
			expected: FrequencyResponse{
				ID:    freqID.String(),
				Name:  "Daily",
				Reset: "0 0 * * *",
			},
		},
		{
			name: "weekly frequency",
			freq: models.Frequency{
				ID:    freqID,
				Name:  "Weekly",
				Reset: "0 0 * * 0",
			},
			expected: FrequencyResponse{
				ID:    freqID.String(),
				Name:  "Weekly",
				Reset: "0 0 * * 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frequencyToResponse(tt.freq)

			if result.ID != tt.expected.ID {
				t.Errorf("ID: expected %q, got %q", tt.expected.ID, result.ID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name: expected %q, got %q", tt.expected.Name, result.Name)
			}
			if result.Reset != tt.expected.Reset {
				t.Errorf("Reset: expected %q, got %q", tt.expected.Reset, result.Reset)
			}
		})
	}
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
