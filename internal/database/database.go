// Package database handles the database connection and migration
package database

import (
	"log"

	"github.com/jhoffmann/dailies/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var dbPath string

// Init initializes the database connection and performs auto-migration.
// It accepts a database file path parameter.
func Init(path string) {
	dbPath = path

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

// GetDBPath returns the database file path.
func GetDBPath() string {
	return dbPath
}
