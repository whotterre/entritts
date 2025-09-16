package middleware

import (
	"gateway/internal/config"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"go.uber.org/zap"
)

// AuthMiddleware performs authentication with Bearer Auth
func RequireAuth(config *config.Config, logger *zap.Logger) fiber.Handler {
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

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		logger.Info("Received token string", zap.String("token", tokenString))
		if tokenString == "" {
			logger.Warn("Empty token string")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Empty token",
			})
		}

		var jsonToken map[string]interface{}
		err := paseto.NewV2().Decrypt(tokenString, []byte(config.PasetoSecret), &jsonToken, nil)
		if err != nil {
			logger.Warn("Token validation failed", zap.String("token", tokenString), zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token: " + err.Error(),
			})
		}

		// Check token expiration
		exp, ok := jsonToken["exp"].(float64)
		if !ok || int64(exp) < time.Now().Unix() {
			logger.Warn("Token has expired or missing exp")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token has expired or missing exp",
			})
		}

		userIDStr, ok := jsonToken["user_id"].(string)
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
