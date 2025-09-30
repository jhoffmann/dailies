// Package config provides configuration management for the dailies application.
package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// AppConfig holds all application configuration settings.
type AppConfig struct {
	// Database settings
	DBPath string

	// Server settings
	Port int

	// Timezone settings
	Timezone string
	Location *time.Location
}

// ParseFlags parses command line flags and environment variables to create application configuration.
// It follows the precedence: CLI flags > Environment variables > Default values.
func ParseFlags() (*AppConfig, error) {
	dbPath := flag.String("db-path", "", "Path to database file")
	apiPort := flag.Int("port", 8080, "The port to listen to")
	dbTimezone := flag.String("tz", "", "Timezone for scheduler (e.g., America/Denver, UTC)")

	flag.Parse()

	config := &AppConfig{}

	// Resolve database path: CLI flag > env var > default
	if *dbPath != "" {
		config.DBPath = *dbPath
	} else if envPath := os.Getenv("DB_PATH"); envPath != "" {
		config.DBPath = envPath
	} else {
		config.DBPath = "dailies.db"
	}

	// Resolve port
	config.Port = *apiPort

	// Resolve timezone: CLI flag > env var > default
	if *dbTimezone != "" {
		config.Timezone = *dbTimezone
	} else if envTimezone := os.Getenv("DB_TIMEZONE"); envTimezone != "" {
		config.Timezone = envTimezone
	} else {
		config.Timezone = "UTC"
	}

	// Load and validate timezone
	location, err := time.LoadLocation(config.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone '%s': %w", config.Timezone, err)
	}
	config.Location = location

	return config, nil
}

// GetTimezoneInfo returns timezone information for API responses.
func (c *AppConfig) GetTimezoneInfo() TimezoneInfo {
	now := time.Now().In(c.Location)
	_, offset := now.Zone()

	return TimezoneInfo{
		Timezone: c.Timezone,
		Offset:   formatOffset(offset),
		Name:     now.Format("MST"),
	}
}

// TimezoneInfo represents timezone configuration information for API responses.
type TimezoneInfo struct {
	Timezone string `json:"timezone"`
	Offset   string `json:"offset"`
	Name     string `json:"name"`
}

// formatOffset converts offset in seconds to string format like "+0600" or "-0700".
func formatOffset(offset int) string {
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	sign := "+"
	if offset < 0 {
		sign = "-"
		hours = -hours
		minutes = -minutes
	}

	return fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
}
