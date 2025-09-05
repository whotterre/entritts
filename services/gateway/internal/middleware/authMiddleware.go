package middleware

import (
	"gateway/internal/config"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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