package main

import (
	"gateway/internal/config"
	"strings"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TODO: Add rate limiting as well 

// ServiceRegistry holds the URLs for the different microservices.
type ServiceRegistry struct {
	services map[string]string
	logger   *zap.Logger
}

func NewServiceRegistry(logger *zap.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		services: map[string]string{
			"users":         "http://user-service:3000",
			"events":        "http://event-service:3000",
			"orders":        "http://order-service:3000",
			"tickets":       "http://ticket-service:3000",
			"notifications": "http://notification-service:3000",
			"payments":      "http://payment:3000",
		},
		logger: logger,
	}
}

// AuthMiddleware performs authentication with Bearer Auth
func AuthMiddleware(config *config.Config, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if isPublicRoute(c.Path()) {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}
		// Ensure it has "Bearer prefix"
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn("Invalid auth type")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid auth type",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		// Ensure JWT is valid
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return "", nil
		})

		if err != nil {
			logger.Warn("Invalid token", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Warn("Invalid token claims")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			logger.Warn("Invalid user ID in token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user ID in token",
			})
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			logger.Warn("Invalid user ID format", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user ID format",
			})
		}
		

		c.Locals("userID", userID)
		return c.Next()
	}
}

func isPublicRoute(path string) bool {
	publicRoutes := []string{"/health", "/api/v1/users/login", "/api/v1/users/register"}
	for _, route := range publicRoutes {
		if path == route {
			return true
		}
	}
	return false
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	registry := NewServiceRegistry(logger)

	app := fiber.New()

	// Middleware
	app.Use(fiberzap.New(fiberzap.Config{Logger: logger}))
	app.Use(cors.New())
	config := config.LoadConfig()
	
	app.Use(AuthMiddleware(config, logger)) 

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "gateway",
		})
	})

	// Setup proxy routes for all services
	for serviceName, serviceURL := range registry.services {
		// Create a closure to capture the serviceName and serviceURL
		func(name, url string) {
			// app.All registers the route for all HTTP methods
			app.All("/api/v1/" + name + "/*", func(c *fiber.Ctx) error {
				targetURL := url + c.Params("*")
				if err := proxy.Do(c, targetURL); err != nil {
					return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
						"error": "Service unavailable",
					})
				}
				// Remove the Gateway's server header if present
				c.Response().Header.Del(fiber.HeaderServer)
				return nil
			})
		}(serviceName, serviceURL)
	}

	logger.Fatal("Gateway server crashed", zap.Error(app.Listen(config.GatewayPort)))
}
