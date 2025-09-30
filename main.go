// Package main provides the entry point for the dailies REST API server.
// It sets up the database connection, configures the Gin router with middleware,
// and starts the HTTP server on port 8080 (default).
package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jhoffmann/dailies/config"
	"github.com/jhoffmann/dailies/handlers"
	"github.com/jhoffmann/dailies/middleware"
	"github.com/jhoffmann/dailies/services"
)

// main initializes the application, sets up the database connection,
// configures the HTTP router with all endpoints, and starts the server.
func main() {
	// Parse configuration from flags and environment variables
	appConfig, err := config.ParseFlags()
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}
	log.Printf("Using timezone: %s", appConfig.Timezone)

	db, err := config.SetupDatabase(appConfig.DBPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize and start WebSocket manager
	wsManager := services.NewWebSocketManager()
	go wsManager.Run()

	// Initialize and start the task scheduler
	scheduler := services.NewTaskScheduler(db, appConfig.Location, appConfig.Timezone)
	scheduler.SetWebSocketManager(wsManager)
	scheduler.Start()
	defer scheduler.Stop()

	r := gin.Default()

	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
		tasks := api.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks(db))
			tasks.GET("/:id", handlers.GetTask(db))
			tasks.POST("", handlers.CreateTask(db, wsManager))
			tasks.PUT("/:id", handlers.UpdateTask(db, wsManager))
			tasks.DELETE("/:id", handlers.DeleteTask(db, wsManager))
		}

		frequencies := api.Group("/frequencies")
		{
			frequencies.GET("", handlers.GetFrequencies(db))
			frequencies.GET("/timers", handlers.GetFrequencyTimers(db, appConfig.Location, appConfig.Timezone))
			frequencies.GET("/:id", handlers.GetFrequency(db))
			frequencies.POST("", handlers.CreateFrequency(db, wsManager))
			frequencies.PUT("/:id", handlers.UpdateFrequency(db, wsManager))
			frequencies.DELETE("/:id", handlers.DeleteFrequency(db, wsManager))
		}

		tags := api.Group("/tags")
		{
			tags.GET("", handlers.GetTags(db))
			tags.GET("/:id", handlers.GetTag(db))
			tags.POST("", handlers.CreateTag(db, wsManager))
			tags.PUT("/:id", handlers.UpdateTag(db, wsManager))
			tags.DELETE("/:id", handlers.DeleteTag(db, wsManager))
		}
	}

	r.GET("/health", handlers.GetHealth(db))
	r.GET("/ws", wsManager.HandleWebSocket())

	// Add timezone endpoint
	api.GET("/timezone", handlers.GetTimezone(appConfig))

	log.Printf("Starting server on :%d", appConfig.Port)
	if err := r.Run(fmt.Sprintf(":%d", appConfig.Port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
