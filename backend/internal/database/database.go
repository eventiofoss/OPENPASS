package database

import (
	"log/slog"
	"os"
	"strings"
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

	gormLogLevel := logger.Info
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
		gormLogLevel = logger.Error
	}

	// Retry up to 10 times with 2-second intervals
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
			Logger: logger.Default.LogMode(gormLogLevel),
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
	// Rename legacy organizers table to users if this is an upgraded deployment.
	if err := db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'organizers'
			) AND NOT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'users'
			) THEN
				ALTER TABLE organizers RENAME TO users;
			END IF;
		END $$;
	`).Error; err != nil {
		return err
	}

	// Drop old role constraints so GORM can apply the users role check.
	if err := db.Exec("ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS chk_organizers_role").Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS organizers_role").Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_role").Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.FormField{},
		&models.Attendee{},
		&models.Payment{},
		&models.CheckIn{},
		&models.Export{},
	)
}
