package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/eventiofoss/eventio/backend/internal/api"
	"github.com/eventiofoss/eventio/backend/internal/database"
	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/eventiofoss/eventio/backend/internal/service"
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

	// Global check on boot to wait for DB and run migrations
	db := database.Connect(connStr)

	app.Get("/api/health", func(c *fiber.Ctx) error {
		sqlDB, err := db.DB()
		if err != nil {
			slog.Error("Failed to get database instance during healthcheck", slog.String("error", err.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Database connection failed",
			})
		}

		if err := sqlDB.Ping(); err != nil {
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

	// Build dependency chain: repository → service → handler
	orgRepo := repository.NewOrganizerRepository(db)
	authSvc := service.NewAuthService(orgRepo)

	eventRepo := repository.NewEventRepository(db)
	eventSvc := service.NewEventService(eventRepo)

	formRepo := repository.NewFormFieldRepository(db)
	formSvc := service.NewFormService(formRepo, eventRepo)

	h := &api.Handler{
		DB:     db,
		Auth:   authSvc,
		Events: eventSvc,
		Forms:  formSvc,
	}

	// Authentication routes group
	authGroup := app.Group("/api/auth")
	authLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many authentication attempts. Please retry in a minute.",
			})
		},
	})
	authGroup.Post("/register", authLimiter, h.Register)
	authGroup.Post("/login", authLimiter, h.Login)
	authGroup.Get("/me", middleware.RequireAuth(), h.Me)
	authGroup.Post(
		"/logout",
		middleware.RequireAuth(),
		h.Logout,
	)

	// Events routes group (protected)
	eventsGroup := app.Group("/api/events", middleware.RequireAuth())
	eventsGroup.Post("/", h.CreateEvent)
	eventsGroup.Get("/", h.ListEvents)
	eventsGroup.Get("/:id", h.GetEvent)
	eventsGroup.Patch("/:id", h.UpdateEvent)
	eventsGroup.Delete("/:id", h.DeleteEvent)
	eventsGroup.Post("/:id/forms", h.SetFormFields)
	eventsGroup.Get("/:id/forms", h.GetFormFields)

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
