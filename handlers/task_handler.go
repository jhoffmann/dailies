// Package handlers provides HTTP request handlers for the REST API endpoints.
package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jhoffmann/dailies/models"
	"gorm.io/gorm"
)

// GetTasks returns a handler function for retrieving all tasks with optional filtering.
func GetTasks(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tasks []models.Task
		query := db.Preload("Tags").Preload("Frequency").Where("deleted = ?", false)

		// Filter by completion status
		if completed := c.Query("completed"); completed != "" {
			if comp, err := strconv.ParseBool(completed); err == nil {
				query = query.Where("completed = ?", comp)
			}
		}

		// Filter by name (partial matching)
		if name := c.Query("name"); name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}

		// Filter by tag IDs
		if tagIds := c.Query("tag_ids"); tagIds != "" {
			ids := strings.Split(tagIds, ",")
			query = query.Joins("JOIN task_tags ON tasks.id = task_tags.task_id").
				Where("task_tags.tag_id IN ?", ids).
				Distinct()
		}

		// Filter by tag names
		if tagNames := c.Query("tag"); tagNames != "" {
			names := strings.Split(tagNames, ",")
			query = query.Joins("JOIN task_tags ON tasks.id = task_tags.task_id").
				Joins("JOIN tags ON task_tags.tag_id = tags.id").
				Where("tags.name IN ?", names).
				Distinct()
		}

		// Sorting
		sort := c.DefaultQuery("sort", "created_at")
		switch sort {
		case "completed":
			query = query.Order("tasks.completed ASC, tasks.priority ASC")
		case "priority":
			query = query.Order("tasks.priority ASC")
		case "name":
			query = query.Order("tasks.name")
		default:
			query = query.Order("tasks.created_at ASC")
		}

		if err := query.Find(&tasks).Error; err != nil {
			log.Println("Error fetching tasks:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
			return
		}

		c.JSON(http.StatusOK, tasks)
	}
}

// GetTask returns a handler function for retrieving a specific task by ID.
func GetTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var task models.Task

		if err := db.Preload("Tags").Preload("Frequency").Where("deleted = ?", false).First(&task, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}
			log.Println("Error fetching task:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
			return
		}

		c.JSON(http.StatusOK, task)
	}
}

// CreateTaskRequest represents the request payload for creating a task.
type CreateTaskRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	FrequencyID *string  `json:"frequency_id,omitempty"`
	TagIDs      []string `json:"tag_ids,omitempty"`
}

// CreateTask returns a handler function for creating a new task.
func CreateTask(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate priority range
		if req.Priority != nil && (*req.Priority < 1 || *req.Priority > 5) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Priority must be between 1 and 5"})
			return
		}

		// Validate frequency exists if provided
		if req.FrequencyID != nil {
			var frequency models.Frequency
			if err := db.First(&frequency, "id = ?", *req.FrequencyID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Frequency not found"})
					return
				}
				log.Println("Error validating frequency:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate frequency"})
				return
			}
		}

		// Create task
		task := models.Task{
			Name:        req.Name,
			Description: req.Description,
			Priority:    req.Priority,
			FrequencyID: req.FrequencyID,
		}

		// Handle tags if provided
		var tags []models.Tag
		if len(req.TagIDs) > 0 {
			if err := db.Find(&tags, "id IN ?", req.TagIDs).Error; err != nil {
				log.Println("Error fetching tags:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
				return
			}
			if len(tags) != len(req.TagIDs) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "One or more tags not found"})
				return
			}
		}

		if err := db.Create(&task).Error; err != nil {
			log.Println("Error creating task:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		// Associate tags
		if len(tags) > 0 {
			if err := db.Model(&task).Association("Tags").Append(&tags); err != nil {
				log.Println("Error associating tags:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate tags"})
				return
			}
		}

		// Reload with associations
		if err := db.Preload("Tags").Preload("Frequency").First(&task, "id = ?", task.ID).Error; err != nil {
			log.Println("Error reloading task:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload task"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("task_create", task)
			}
		}

		c.JSON(http.StatusCreated, task)
	}
}

// UpdateTaskRequest represents the request payload for updating a task.
type UpdateTaskRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Completed   *bool    `json:"completed,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	FrequencyID *string  `json:"frequency_id,omitempty"`
	TagIDs      []string `json:"tag_ids,omitempty"`
}

// UpdateTask returns a handler function for updating an existing task.
func UpdateTask(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req UpdateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var task models.Task
		if err := db.Where("deleted = ?", false).First(&task, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}
			log.Println("Error fetching task:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
			return
		}

		// Handle priority: 0 means remove, 1-5 means set, anything else is invalid
		removePriority := false
		if req.Priority != nil {
			if *req.Priority == 0 {
				removePriority = true
			} else if *req.Priority < 1 || *req.Priority > 5 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Priority must be between 1 and 5"})
				return
			}
		}

		// Handle empty string frequency ID (treat as removal)
		removeFrequency := false
		if req.FrequencyID != nil && *req.FrequencyID == "" {
			removeFrequency = true
		}

		// Validate frequency exists if provided and not empty
		if req.FrequencyID != nil && *req.FrequencyID != "" {
			var frequency models.Frequency
			if err := db.First(&frequency, "id = ?", *req.FrequencyID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Frequency not found"})
					return
				}
				log.Println("Error validating frequency:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate frequency"})
				return
			}
		}

		// Update fields
		updates := make(map[string]any)
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}
		if req.Completed != nil {
			updates["completed"] = *req.Completed
		}
		// Handle priority: set to nil to remove, or set to value (1-5)
		if removePriority {
			updates["priority"] = nil
		} else if req.Priority != nil {
			updates["priority"] = *req.Priority
		}
		// Handle frequency_id: set to nil to remove, or set to ID value
		if removeFrequency {
			updates["frequency_id"] = nil
		} else if req.FrequencyID != nil {
			updates["frequency_id"] = *req.FrequencyID
		}

		if len(updates) > 0 {
			if err := db.Model(&task).Updates(updates).Error; err != nil {
				log.Println("Error updating task:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
				return
			}
		}

		// Handle tag associations
		if req.TagIDs != nil {
			var tags []models.Tag
			if len(req.TagIDs) > 0 {
				if err := db.Find(&tags, "id IN ?", req.TagIDs).Error; err != nil {
					log.Println("Error fetching tags:", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
					return
				}
				if len(tags) != len(req.TagIDs) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "One or more tags not found"})
					return
				}
			}

			// Replace all tag associations
			if err := db.Model(&task).Association("Tags").Replace(&tags); err != nil {
				log.Println("Error updating tag associations:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag associations"})
				return
			}
		}

		// Reload with associations
		if err := db.Preload("Tags").Preload("Frequency").First(&task, "id = ?", task.ID).Error; err != nil {
			log.Println("Error reloading task:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload task"})
			return
		}

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("task_update", task)
			}
		}

		c.JSON(http.StatusOK, task)
	}
}

// DeleteTask returns a handler function for soft deleting a task.
func DeleteTask(db *gorm.DB, wsManager ...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var task models.Task
		if err := db.Preload("Tags").Preload("Frequency").Where("deleted = ?", false).First(&task, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}
			log.Println("Error fetching task:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
			return
		}

		// Soft delete by setting the deleted flag
		if err := db.Model(&task).Update("deleted", true).Error; err != nil {
			log.Println("Error soft deleting task:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
			return
		}

		// Update the task object for the WebSocket event
		task.Deleted = true

		// Broadcast WebSocket event
		if len(wsManager) > 0 && wsManager[0] != nil {
			if ws, ok := wsManager[0].(interface {
				Broadcast(eventType any, data any)
			}); ok {
				ws.Broadcast("task_delete", task)
			}
		}

		c.JSON(http.StatusNoContent, nil)
	}
}
