package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jhoffmann/dailies/models"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// GetFrequencies returns a handler function for retrieving all frequencies with optional filtering.
func GetFrequencies(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var frequencies []models.Frequency
		query := db.Model(&models.Frequency{})

		// Filter by name (partial matching)
		if name := c.Query("name"); name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}

		// Default sorting by name
		query = query.Order("name")

		if err := query.Preload("Tasks").Find(&frequencies).Error; err != nil {
			log.Println("Error fetching frequencies:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch frequencies"})
			return
		}

		c.JSON(http.StatusOK, frequencies)
	}
}

// GetFrequency returns a handler function for retrieving a specific frequency by ID.
func GetFrequency(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var frequency models.Frequency

		if err := db.Preload("Tasks").First(&frequency, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Frequency not found"})
				return
			}
			log.Println("Error fetching frequency:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch frequency"})
			return
		}

		c.JSON(http.StatusOK, frequency)
	}
}

// CreateFrequencyRequest represents the request payload for creating a frequency.
type CreateFrequencyRequest struct {
	Name   string `json:"name" binding:"required"`
	Period string `json:"period" binding:"required"`
}

// validateCronExpression validates that a cron expression is valid.
func validateCronExpression(expr string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err := parser.Parse(expr)
	return err
}

// CreateFrequency returns a handler function for creating a new frequency.
func CreateFrequency(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateFrequencyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate cron expression
		if err := validateCronExpression(req.Period); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cron expression: " + err.Error()})
			return
		}

		frequency := models.Frequency{
			Name:   strings.TrimSpace(req.Name),
			Period: strings.TrimSpace(req.Period),
		}

		if err := db.Create(&frequency).Error; err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
				c.JSON(http.StatusConflict, gin.H{"error": "Frequency with this name already exists"})
				return
			}
			log.Println("Error creating frequency:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create frequency"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("frequency_create", frequency)
			}
		}

		c.JSON(http.StatusCreated, frequency)
	}
}

// UpdateFrequencyRequest represents the request payload for updating a frequency.
type UpdateFrequencyRequest struct {
	Name   *string `json:"name,omitempty"`
	Period *string `json:"period,omitempty"`
}

// UpdateFrequency returns a handler function for updating an existing frequency.
func UpdateFrequency(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req UpdateFrequencyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var frequency models.Frequency
		if err := db.First(&frequency, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Frequency not found"})
				return
			}
			log.Println("Error fetching frequency:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch frequency"})
			return
		}

		// Validate cron expression if provided
		if req.Period != nil {
			if err := validateCronExpression(*req.Period); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cron expression: " + err.Error()})
				return
			}
		}

		// Update fields
		updates := make(map[string]any)
		if req.Name != nil {
			updates["name"] = strings.TrimSpace(*req.Name)
		}
		if req.Period != nil {
			updates["period"] = strings.TrimSpace(*req.Period)
		}

		if len(updates) > 0 {
			if err := db.Model(&frequency).Updates(updates).Error; err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
					c.JSON(http.StatusConflict, gin.H{"error": "Frequency with this name already exists"})
					return
				}
				log.Println("Error updating frequency:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update frequency"})
				return
			}
		}

		// Reload the frequency
		if err := db.First(&frequency, "id = ?", id).Error; err != nil {
			log.Println("Error reloading frequency:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload frequency"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("frequency_update", frequency)
			}
		}

		c.JSON(http.StatusOK, frequency)
	}
}

// DeleteFrequency returns a handler function for deleting a frequency.
func DeleteFrequency(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var frequency models.Frequency
		if err := db.First(&frequency, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Frequency not found"})
				return
			}
			log.Println("Error fetching frequency:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch frequency"})
			return
		}

		// Clear frequency_id from associated tasks
		if err := db.Model(&models.Task{}).Where("frequency_id = ?", id).Update("frequency_id", nil).Error; err != nil {
			log.Println("Error clearing frequency references from tasks:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear frequency references"})
			return
		}

		// Get frequency data before deletion for WebSocket event
		var frequencyForEvent models.Frequency
		if len(wsManager) > 0 && wsManager[0] != nil {
			db.First(&frequencyForEvent, "id = ?", id)
		}

		if err := db.Delete(&frequency).Error; err != nil {
			log.Println("Error deleting frequency:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete frequency"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("frequency_delete", frequencyForEvent)
			}
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

// FrequencyTimer represents the response structure for the timers endpoint.
type FrequencyTimer struct {
	Name           string `json:"name"`
	TimeUntilReset string `json:"time_until_reset"`
}

// GetFrequencyTimers returns a handler function for retrieving timer information
// for all frequencies using the specified timezone.
func GetFrequencyTimers(db *gorm.DB, location *time.Location, timezone string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var frequencies []models.Frequency
		if err := db.Order("name").Find(&frequencies).Error; err != nil {
			log.Println("Error fetching frequencies:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch frequencies"})
			return
		}

		var timers []FrequencyTimer
		for _, freq := range frequencies {
			timeUntilReset, err := freq.TimeUntilNextReset(location, timezone)
			if err != nil {
				log.Printf("Error calculating time until reset for frequency %s: %v", freq.Name, err)
				continue
			}

			timers = append(timers, FrequencyTimer{
				Name:           freq.Name,
				TimeUntilReset: timeUntilReset,
			})
		}

		c.JSON(http.StatusOK, timers)
	}
}
