package database

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

// Connect initializes the database connection with the provided connection string.
// It includes a retry loop to wait for the database container to become fully available.
func Connect(connStr string) *pgx.Conn {
	var conn *pgx.Conn
	var err error

	// Retry up to 10 times with 2-second intervals
	for i := 0; i < 10; i++ {
		conn, err = pgx.Connect(context.Background(), connStr)
		if err == nil {
			err = conn.Ping(context.Background())
			if err == nil {
				slog.Info("Successfully connected to the database")
				return conn
			}
		}

		slog.Warn("Failed to connect to DB, retrying...", slog.Int("attempt", i+1), slog.String("error", err.Error()))
		time.Sleep(2 * time.Second)
	}
	slog.Error("Could not connect to database after 10 attempts", slog.String("error", err.Error()))
	os.Exit(1)
	return nil
}
