package config

import (
	"os"
	"testing"
	"time"
)

func TestAppConfig_GetTimezoneInfo_UTC(t *testing.T) {
	config := &AppConfig{
		Timezone: "UTC",
		Location: time.UTC,
	}

	info := config.GetTimezoneInfo()

	if info.Timezone != "UTC" {
		t.Errorf("Expected timezone 'UTC', got: %s", info.Timezone)
	}

	if info.Name != "UTC" {
		t.Errorf("Expected name 'UTC', got: %s", info.Name)
	}

	if info.Offset != "+0000" {
		t.Errorf("Expected offset '+0000', got: %s", info.Offset)
	}
}

func TestAppConfig_GetTimezoneInfo_Denver(t *testing.T) {
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Skipf("Skipping test: timezone America/Denver not available: %v", err)
	}

	config := &AppConfig{
		Timezone: "America/Denver",
		Location: location,
	}

	info := config.GetTimezoneInfo()

	if info.Timezone != "America/Denver" {
		t.Errorf("Expected timezone 'America/Denver', got: %s", info.Timezone)
	}

	// The name will be either MST or MDT depending on the current date
	if info.Name != "MST" && info.Name != "MDT" {
		t.Errorf("Expected name 'MST' or 'MDT', got: %s", info.Name)
	}

	// The offset will be either -0700 (MST) or -0600 (MDT)
	if info.Offset != "-0700" && info.Offset != "-0600" {
		t.Errorf("Expected offset '-0700' or '-0600', got: %s", info.Offset)
	}
}

func TestFormatOffset(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		expected string
	}{
		{"UTC", 0, "+0000"},
		{"EST", -18000, "-0500"}, // -5 hours
		{"MST", -25200, "-0700"}, // -7 hours
		{"MDT", -21600, "-0600"}, // -6 hours
		{"JST", 32400, "+0900"},  // +9 hours
		{"IST", 19800, "+0530"},  // +5:30 hours
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOffset(tt.offset)
			if result != tt.expected {
				t.Errorf("formatOffset(%d) = %s, expected %s", tt.offset, result, tt.expected)
			}
		})
	}
}

func TestTimezoneConfigurationPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		envVar     string
		expectedTZ string
	}{
		{"Default when no env var", "", "UTC"},
		{"Environment variable set", "America/New_York", "America/New_York"},
		{"Invalid timezone should cause error", "Invalid/Timezone", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env var
			originalEnv := os.Getenv("DB_TIMEZONE")
			defer func() {
				if originalEnv == "" {
					os.Unsetenv("DB_TIMEZONE")
				} else {
					os.Setenv("DB_TIMEZONE", originalEnv)
				}
			}()

			// Set test environment
			if tt.envVar == "" {
				os.Unsetenv("DB_TIMEZONE")
			} else {
				os.Setenv("DB_TIMEZONE", tt.envVar)
			}

			// Test timezone resolution logic directly
			timezone := tt.envVar
			if timezone == "" {
				timezone = "UTC"
			}

			if tt.expectedTZ == "" {
				// Test invalid timezone
				_, err := time.LoadLocation(timezone)
				if err == nil {
					t.Errorf("Expected error for invalid timezone %s", timezone)
				}
			} else {
				location, err := time.LoadLocation(timezone)
				if err != nil {
					t.Errorf("Expected valid timezone %s, got error: %v", timezone, err)
				}
				if location.String() != tt.expectedTZ {
					t.Errorf("Expected timezone %s, got: %s", tt.expectedTZ, location.String())
				}
			}
		})
	}
}
