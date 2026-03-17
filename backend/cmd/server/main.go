package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"github.com/v4sud3v/eventio/backend/internal/database"
)

// setupLogger initializes a structured JSON logger
func setupLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func main() {
	setupLogger()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		slog.Error("DATABASE_URL environment variable is required but not set")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, err := pgx.Connect(ctx, connStr)
		if err != nil {
			slog.Error("Database connection failed during healthcheck", slog.String("error", err.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Database connection failed",
			})
		}
		defer conn.Close(ctx)

		err = conn.Ping(ctx)
		if err != nil {
			slog.Error("Database ping failed during healthcheck", slog.String("error", err.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Database ping failed",
			})
		}

		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Database connection successful",
		})
	})

	// Global check on boot to wait for DB
	dbConn := database.Connect(connStr)
	if dbConn != nil {
		dbConn.Close(context.Background())
	}

	// Graceful shutdown setup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		slog.Info("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	listenAddr := fmt.Sprintf(":%s", port)
	slog.Info("Starting Eventio Backend", slog.String("port", port))
	if err := app.Listen(listenAddr); err != nil {
		slog.Error("Server encountered an error", slog.String("error", err.Error()))
	}

	slog.Info("Server was successfully shutdown.")
}
