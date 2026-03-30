package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

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

	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		slog.Error("DATABASE_URL environment variable is required but not set")
		os.Exit(1)
	}

	if err := middleware.EnsureJWTSecretConfigured(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	isProd := strings.EqualFold(os.Getenv("APP_ENV"), "production")
	if isProd && len(trustedProxies) == 0 {
		slog.Error("APP_ENV=production requires TRUSTED_PROXIES to be set for accurate client IP handling behind reverse proxies")
		os.Exit(1)
	}

	slog.Info(
		"Proxy trust configuration",
		slog.Bool("trusted_proxy_check_enabled", len(trustedProxies) > 0),
		slog.Int("trusted_proxy_count", len(trustedProxies)),
		slog.String("app_env", strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))),
	)

	app := fiber.New(fiber.Config{
		DisableStartupMessage:   true,
		ReadTimeout:             15 * time.Second,
		WriteTimeout:            15 * time.Second,
		IdleTimeout:             60 * time.Second,
		BodyLimit:               1 * 1024 * 1024,
		EnableTrustedProxyCheck: len(trustedProxies) > 0,
		TrustedProxies:          trustedProxies,
		ProxyHeader:             fiber.HeaderXForwardedFor,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(helmet.New())
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cross-Origin-Opener-Policy", "same-origin")
		c.Set("Cross-Origin-Resource-Policy", "same-site")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		return c.Next()
	})
	csrfCookieDomain := resolveCSRFCookieDomain(isProd)
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "eventio_csrf",
		CookieHTTPOnly: false,
		CookieSecure:   isProd,
		CookieDomain:   csrfCookieDomain,
		CookieSameSite: "Strict",
		Expiration:     30 * time.Minute,
		Next: func(c *fiber.Ctx) bool {
			// Webhooks come from external gateways and cannot provide our CSRF token.
			return strings.HasPrefix(c.Path(), "/api/webhooks/")
		},
	}))

	// Global check on boot to wait for DB and run migrations
	db := database.Connect(connStr)

	app.Get("/api/health", func(c *fiber.Ctx) error {
		sqlDB, err := db.DB()
		if err != nil {
			slog.Error("Failed to get database instance during healthcheck", slog.String("error", err.Error()))
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
			})
		}

		if err := sqlDB.Ping(); err != nil {
			slog.Error("Database ping failed during healthcheck", slog.String("error", err.Error()))
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
			})
		}

		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Build dependency chain: repository → service → handler
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)

	eventRepo := repository.NewEventRepository(db)
	eventSvc := service.NewEventService(eventRepo)

	formRepo := repository.NewFormFieldRepository(db)
	formSvc := service.NewFormService(formRepo, eventRepo)

	attendeeRepo := repository.NewAttendeeRepository(db)
	registrationSvc := service.NewRegistrationService(attendeeRepo, userRepo)

	paymentGateway, paymentEnabled := configuredPaymentGateway()
	var paymentSvc *service.PaymentService
	if paymentEnabled {
		paymentProvider, err := service.NewPaymentProvider(paymentGateway)
		if err != nil {
			slog.Error(
				"Failed to configure payment provider",
				slog.String("gateway", paymentGateway),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		paymentSvc = service.NewPaymentService(db, paymentProvider)
		slog.Info(
			"Payment processing enabled",
			slog.String("gateway", paymentGateway),
		)
	} else {
		slog.Info(
			"Payment processing disabled",
			slog.String(
				"reason",
				"ACTIVE_PAYMENT_GATEWAY environment variable is unset",
			),
		)
	}
	var backgroundWorkers sync.WaitGroup
	startPaymentSweeper(
		appCtx,
		&backgroundWorkers,
		paymentSvc,
		paymentSweepInterval,
	)

	analyticsRepo := repository.NewAnalyticsRepository(db)
	analyticsSvc := service.NewAnalyticsService(
		analyticsRepo, eventRepo,
	)

	checkinRepo := repository.NewCheckInRepository(db)
	qrSvc := service.NewQRService(checkinRepo)

	h := &api.Handler{
		DB:           db,
		Auth:         authSvc,
		Events:       eventSvc,
		Forms:        formSvc,
		Registration: registrationSvc,
		Payment:      paymentSvc,
		Analytics:    analyticsSvc,
		QR:           qrSvc,
	}

	// Authentication routes group
	authGroup := app.Group("/api/auth")
	if isProd {
		slog.Info("Auth limiter configured", slog.Int("max_attempts_per_minute", 5), slog.String("mode", "production"))
	} else {
		slog.Info("Auth limiter configured", slog.Int("max_attempts_per_minute", 50), slog.String("mode", "development"))
	}
	authLimiter := limiter.New(authLimiterConfig(isProd))
	authGroup.Post("/register", authLimiter, h.Register)
	authGroup.Post("/login", authLimiter, h.Login)
	authGroup.Get("/me", middleware.RequireAuth(), h.Me)
	authGroup.Post(
		"/logout",
		middleware.RequireAuth(),
		h.Logout,
	)

	// Public event detail route (no auth middleware)
	app.Get("/api/public/events/:slug", h.GetPublicEvent)

	// Public registration route (no auth middleware)
	registrationLimiter := limiter.New(registrationLimiterConfig())
	app.Post("/api/events/:id/register", registrationLimiter, h.RegisterAttendee)
	paymentLimiter := limiter.New(paymentLimiterConfig())
	app.Post("/api/events/:id/pay", paymentLimiter, h.CreateCheckout)
	app.Post("/api/webhooks/:gateway", h.HandleWebhook)

	// Events routes group (protected)
	eventsGroup := app.Group("/api/events", middleware.RequireAuth())
	eventsGroup.Post("/", h.CreateEvent)
	eventsGroup.Get("/", h.ListEvents)
	eventsGroup.Get("/:id", h.GetEvent)
	eventsGroup.Patch("/:id", h.UpdateEvent)
	eventsGroup.Delete("/:id", h.DeleteEvent)
	eventsGroup.Post("/:id/forms", h.SetFormFields)
	eventsGroup.Get("/:id/forms", h.GetFormFields)
	eventsGroup.Get("/:id/analytics", h.GetAnalytics)
	eventsGroup.Get("/:id/export", h.ExportAttendees)

	// Check-in scan route (protected)
	app.Post(
		"/api/checkins/scan",
		middleware.RequireAuth(),
		h.ScanCheckIn,
	)

	go func() {
		<-appCtx.Done()
		slog.Info("Gracefully shutting down...")
		if err := app.Shutdown(); err != nil {
			slog.Error(
				"Fiber shutdown failed",
				slog.String("error", err.Error()),
			)
		}
	}()

	listenAddr := fmt.Sprintf(":%s", port)
	slog.Info("Starting Eventio Backend", slog.String("port", port))
	if err := app.Listen(listenAddr); err != nil {
		if appCtx.Err() == nil {
			slog.Error(
				"Server encountered an error",
				slog.String("error", err.Error()),
			)
		}
	}

	stop()
	backgroundWorkers.Wait()
	slog.Info("Server was successfully shutdown.")
}

func parseTrustedProxies(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		proxy := strings.TrimSpace(part)
		if proxy != "" {
			out = append(out, proxy)
		}
	}

	return out
}

func resolveCSRFCookieDomain(isProd bool) string {
	if !isProd {
		return ""
	}

	domain := strings.TrimSpace(os.Getenv("CSRF_COOKIE_DOMAIN"))
	if domain == "" {
		return ""
	}

	normalized := strings.TrimPrefix(strings.ToLower(domain), ".")
	if normalized == "localhost" || normalized == "127.0.0.1" || normalized == "::1" {
		return ""
	}

	return domain
}

func authLimiterConfig(isProd bool) limiter.Config {
	maxAttempts := 5
	if !isProd {
		maxAttempts = 50
	}

	return limiter.Config{
		Max:        maxAttempts,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		SkipSuccessfulRequests: true,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many authentication attempts. Please retry in a minute.",
			})
		},
	}
}

func registrationLimiterConfig() limiter.Config {
	return limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many registration attempts. Please retry in a minute.",
			})
		},
	}
}

func paymentLimiterConfig() limiter.Config {
	return limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many checkout attempts. Please retry in a minute.",
			})
		},
	}
}

func configuredPaymentGateway() (string, bool) {
	gateway := strings.ToLower(
		strings.TrimSpace(os.Getenv("ACTIVE_PAYMENT_GATEWAY")),
	)
	if gateway == "" {
		return "", false
	}

	return gateway, true
}
