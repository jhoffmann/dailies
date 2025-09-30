// Package config provides database configuration and setup functionality.
package config

import (
	"log"

	"github.com/jhoffmann/dailies/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupDatabase initializes and configures the SQLite database connection.
// It takes a dbPath parameter specifying the path to the database file,
// runs migrations, and returns a configured GORM database instance.
func SetupDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// migrate runs database migrations for all models and creates necessary indexes.
func migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.Frequency{},
		&models.Tag{},
		&models.Task{},
	)
	if err != nil {
		return err
	}

	if err := addIndexes(db); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// addIndexes creates database indexes to improve query performance.
func addIndexes(db *gorm.DB) error {
	// Single column indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_completed ON tasks(completed)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_frequency_id ON tasks(frequency_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_deleted ON tasks(deleted)").Error; err != nil {
		return err
	}

	// Composite indexes for common query patterns with soft deletes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_deleted_completed ON tasks(deleted, completed)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_deleted_completed_frequency ON tasks(deleted, completed, frequency_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_deleted_id ON tasks(deleted, id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_frequencies_name ON frequencies(name)").Error; err != nil {
		return err
	}

	return nil
}
