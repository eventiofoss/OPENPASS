package database

import (
	"log/slog"
	"os"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect initializes the database connection with the provided connection string.
// It includes a retry loop to wait for the database container to become fully available,
// and automatically runs migrations on all registered models.
func Connect(connStr string) *gorm.DB {
	var db *gorm.DB
	var err error

	// Retry up to 10 times with 2-second intervals
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			slog.Info("Successfully connected to the database")

			// Run auto-migrations
			if err := autoMigrate(db); err != nil {
				slog.Error("Auto-migration failed", slog.String("error", err.Error()))
				os.Exit(1)
			}

			return db
		}

		slog.Warn("Failed to connect to DB, retrying...", slog.Int("attempt", i+1), slog.String("error", err.Error()))
		time.Sleep(2 * time.Second)
	}
	slog.Error("Could not connect to database after 10 attempts", slog.String("error", err.Error()))
	os.Exit(1)
	return nil
}

// autoMigrate runs all database migrations for registered models.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Organizer{},
		&models.Event{},
		&models.FormField{},
		&models.Attendee{},
		&models.Payment{},
		&models.CheckIn{},
		&models.Export{},
	)
}
