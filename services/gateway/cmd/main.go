package main

import (
	"gateway/internal/config"
	"gateway/internal/middleware"
	"time"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"go.uber.org/zap"
)

// TODO: Add rate limiting as well

// ServiceRegistry holds the URLs for the different microservices.
type ServiceRegistry struct {
	services map[string]string
	logger   *zap.Logger
}

func NewServiceRegistry(config *config.Config, logger *zap.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		services: map[string]string{
			"users":         "http://" + config.UserServiceHost + ":8081",
			"events":        "http://" + config.EventServiceHost + ":8080",
			"orders":        "http://" + config.OrderServiceHost + ":3000",
			"tickets":       "http://" + config.TicketServiceHost + ":3000",
			"notifications": "http://" + config.NotificationServiceHost + ":3000",
			"payments":      "http://" + config.PaymentServiceHost + ":3000",
		},
		logger: logger,
	}
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	rateLimiter := limiter.New(limiter.Config{
		Max: 20,
		Expiration: 1 * time.Second,
	})
	cfg := config.LoadConfig()
	registry := NewServiceRegistry(cfg, logger)

	app := fiber.New()

	// Middleware
	app.Use(fiberzap.New(fiberzap.Config{Logger: logger}))
	app.Use(rateLimiter)
	app.Use(cors.New())

	// Public Routes Group
	publicGroup := app.Group("/api/v1")

	// Health check
	publicGroup.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "gateway",
		})
	})

	// Public auth routes
	publicGroup.All("/auth/*", func(c *fiber.Ctx) error {
		// Get the path after /api/v1/auth/
		path := c.Params("*")

		// Remove any leading slashes from path
		for len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		// Check if user service exists in registry
		if _, exists := registry.services["users"]; !exists {
			logger.Error("User service not found in registry")
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Service configuration error",
			})
		}

		// Build the target URL with the correct path
		// Convert /user/ to /users/ for proper user service routing
		if len(path) >= 5 && path[:5] == "user/" {
			path = "users/" + path[5:]
		}
		targetURL := registry.services["users"] + "/" + path

		logger.Info("Proxying auth request",
			zap.String("targetURL", targetURL),
			zap.String("originalURL", c.OriginalURL()),
			zap.String("path", path))

		// Attempt the proxy request with simple retry logic
		var err error
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			if err = proxy.Do(c, targetURL); err == nil {
				break
			}
			if i < maxRetries-1 {
				logger.Info("Retrying proxy request",
					zap.String("targetURL", targetURL),
					zap.Int("attempt", i+2),
					zap.Error(err))
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			}
		}

		if err != nil {
			logger.Error("Auth proxy error",
				zap.Error(err),
				zap.String("targetURL", targetURL))

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   "Authentication service unavailable",
				"details": err.Error(),
			})
		}
		return nil
	})

	// Public user routes (for registration)
	publicGroup.All("/users/*", func(c *fiber.Ctx) error {
		// Get the path after /api/v1/users/
		path := c.Params("*")

		// Remove any leading slashes from path
		for len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		// Check if user service exists in registry
		if _, exists := registry.services["users"]; !exists {
			logger.Error("User service not found in registry")
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Service configuration error",
			})
		}

		// Build the target URL
		targetURL := registry.services["users"] + "/users/" + path

		logger.Info("Proxying user request",
			zap.String("targetURL", targetURL),
			zap.String("originalURL", c.OriginalURL()),
			zap.String("path", path))

		// Attempt the proxy request with simple retry logic
		var err error
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			if err = proxy.Do(c, targetURL); err == nil {
				break
			}
			if i < maxRetries-1 {
				logger.Info("Retrying proxy request",
					zap.String("targetURL", targetURL),
					zap.Int("attempt", i+2),
					zap.Error(err))
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			}
		}

		if err != nil {
			logger.Error("User proxy error",
				zap.Error(err),
				zap.String("targetURL", targetURL))

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   "User service unavailable",
				"details": err.Error(),
			})
		}
		return nil
	})

	// Public event routes
	publicGroup.Get("/events", func(c *fiber.Ctx) error {
		targetURL := registry.services["events"] + "/public" + c.Params("*")
		if err := proxy.Do(c, targetURL); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Events service unavailable",
			})
		}
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	})

	// Protected Routes Group
	privateGroup := app.Group("/api/v1")
	privateGroup.Use(middleware.RequireAuth(cfg, logger))

	// Setup proxy routes for protected services
	for serviceName, serviceURL := range registry.services {
		// Skip auth routes as they're handled separately
		if serviceName == "auth" {
			continue
		}

		// Create a closure to capture the serviceName and serviceURL
		func(name, url string) {
			privateGroup.All("/"+name+"/*", func(c *fiber.Ctx) error {
				targetURL := url + c.Params("*")
				if err := proxy.Do(c, targetURL); err != nil {
					return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
						"error": "Service unavailable",
					})
				}
				c.Response().Header.Del(fiber.HeaderServer)
				return nil
			})
		}(serviceName, serviceURL)
	}

	logger.Fatal("Gateway server crashed", zap.Error(app.Listen(cfg.GatewayPort)))
}
