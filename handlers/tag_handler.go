package handlers

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jhoffmann/dailies/models"
	"gorm.io/gorm"
)

// generateRandomColor generates a random hex color code.
func generateRandomColor() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return fmt.Sprintf("#%02x%02x%02x", bytes[0], bytes[1], bytes[2])
}

// validateHexColor validates that a string is a valid hex color.
func validateHexColor(color string) bool {
	hexPattern := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	return hexPattern.MatchString(color)
}

// GetTags returns a handler function for retrieving all tags with optional filtering.
func GetTags(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tags []models.Tag
		query := db.Model(&models.Tag{})

		// Filter by name (partial matching)
		if name := c.Query("name"); name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}

		// Default sorting by name
		query = query.Order("name")

		if err := query.Preload("Tasks").Find(&tags).Error; err != nil {
			log.Println("Error fetching tags:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
			return
		}

		c.JSON(http.StatusOK, tags)
	}
}

// GetTag returns a handler function for retrieving a specific tag by ID.
func GetTag(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var tag models.Tag

		if err := db.Preload("Tasks").First(&tag, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
				return
			}
			log.Println("Error fetching tag:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag"})
			return
		}

		c.JSON(http.StatusOK, tag)
	}
}

// CreateTagRequest represents the request payload for creating a tag.
type CreateTagRequest struct {
	Name  string  `json:"name" binding:"required"`
	Color *string `json:"color,omitempty"`
}

// CreateTag returns a handler function for creating a new tag.
func CreateTag(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateTagRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate color if provided, otherwise generate one
		var color string
		if req.Color != nil {
			color = strings.TrimSpace(*req.Color)
			if !validateHexColor(color) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Color must be a valid hex color (e.g., #ff0000)"})
				return
			}
		} else {
			color = generateRandomColor()
		}

		tag := models.Tag{
			Name:  strings.TrimSpace(req.Name),
			Color: color,
		}

		if err := db.Create(&tag).Error; err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
				c.JSON(http.StatusConflict, gin.H{"error": "Tag with this name already exists"})
				return
			}
			log.Println("Error creating tag:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("tag_create", tag)
			}
		}

		c.JSON(http.StatusCreated, tag)
	}
}

// UpdateTagRequest represents the request payload for updating a tag.
type UpdateTagRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// UpdateTag returns a handler function for updating an existing tag.
func UpdateTag(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req UpdateTagRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var tag models.Tag
		if err := db.First(&tag, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
				return
			}
			log.Println("Error fetching tag:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag"})
			return
		}

		// Validate color if provided
		if req.Color != nil {
			color := strings.TrimSpace(*req.Color)
			if !validateHexColor(color) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Color must be a valid hex color (e.g., #ff0000)"})
				return
			}
		}

		// Update fields
		updates := make(map[string]any)
		if req.Name != nil {
			updates["name"] = strings.TrimSpace(*req.Name)
		}
		if req.Color != nil {
			updates["color"] = strings.TrimSpace(*req.Color)
		}

		if len(updates) > 0 {
			if err := db.Model(&tag).Updates(updates).Error; err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
					c.JSON(http.StatusConflict, gin.H{"error": "Tag with this name already exists"})
					return
				}
				log.Println("Error updating tag:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag"})
				return
			}
		}

		// Reload the tag
		if err := db.First(&tag, "id = ?", id).Error; err != nil {
			log.Println("Error reloading tag:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload tag"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("tag_update", tag)
			}
		}

		c.JSON(http.StatusOK, tag)
	}
}

// DeleteTag returns a handler function for deleting a tag.
func DeleteTag(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var tag models.Tag
		if err := db.First(&tag, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
				return
			}
			log.Println("Error fetching tag:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag"})
			return
		}

		// Store tag data for WebSocket event before deletion
		tagForEvent := tag

		// Clear tag associations from tasks (this will be handled by CASCADE, but being explicit)
		if err := db.Model(&tag).Association("Tasks").Clear(); err != nil {
			log.Println("Error clearing tag associations:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear tag associations"})
			return
		}

		if err := db.Delete(&tag).Error; err != nil {
			log.Println("Error deleting tag:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tag"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("tag_delete", tagForEvent)
			}
		}

		c.JSON(http.StatusNoContent, nil)
	}
}
