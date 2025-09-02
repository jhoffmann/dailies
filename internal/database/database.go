package database

import (
	"log"
	"os"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init initializes the database connection and performs auto-migration.
// It uses the DB_PATH environment variable or defaults to "dailies.db".
func Init() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dailies.db"
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&models.Task{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Printf("Database connected and migrated successfully at %s", dbPath)
}

// GetDB returns the global database instance.
func GetDB() *gorm.DB {
	return DB
}
