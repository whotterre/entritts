package routes

import (
	"user-service/internal/handlers"
	"user-service/internal/repositories"
	"user-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, jwtSecret string, logger *zap.Logger) {
	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	sessionsRepo := repositories.NewSessionRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo, sessionsRepo) 
	userHandler := handlers.NewUserHandler(userService, logger, jwtSecret)

	// Setup routes
	app.Post("/users/register", userHandler.CreateNewUser)
	app.Post("/users/login", userHandler.LoginUser)
	app.Get("/health", userHandler.GetHealthStatus)
}
