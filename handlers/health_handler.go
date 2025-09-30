// Package handlers provides HTTP request handlers for the dailies REST API.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jhoffmann/dailies/config"
	"gorm.io/gorm"
)

// GetHealth returns a Gin handler function that checks the health of the service
// and database connection. It returns HTTP 200 if healthy, HTTP 503 if the database
// connection is not active.
func GetHealth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Failed to get database connection",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection is not active",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// GetTimezone returns a Gin handler function that provides timezone information
// used by the scheduler.
func GetTimezone(appConfig *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := appConfig.GetTimezoneInfo()
		c.JSON(http.StatusOK, info)
	}
}
