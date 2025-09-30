package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Frequency represents a recurring schedule for tasks (e.g., daily, weekly).
type Frequency struct {
	ID        string    `json:"id" gorm:"type:text;primaryKey"`
	Name      string    `json:"name" gorm:"not null;unique"`
	Period    string    `json:"period" gorm:"not null"`
	Tasks     []Task    `json:"tasks,omitempty" gorm:"foreignKey:FrequencyID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate is a GORM hook that generates a UUID for the frequency before creation.
func (f *Frequency) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

// TimeUntilNextReset calculates how long until the next reset based on the cron schedule
// using the specified timezone. Returns a human-readable duration string like "6h", "2d", "12m".
func (f *Frequency) TimeUntilNextReset(location *time.Location, timezone string) (string, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse("TZ=" + timezone + " " + f.Period)
	if err != nil {
		return "", err
	}

	now := time.Now().In(location)
	next := schedule.Next(now)
	duration := next.Sub(now)

	return formatDuration(duration), nil
}

// formatDuration converts a time.Duration to a human-readable string.
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0m"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return "1m" // Show at least 1 minute if less than a minute remains
}
