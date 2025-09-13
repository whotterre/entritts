package routes

import (
	"user-service/internal/handlers"
	"user-service/internal/repositories"
	"user-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, logger *zap.Logger) {

	// Health check endpoint
	authRepo := repositories.NewUserRepository(db)
	authService := services.NewUserService(authRepo)
	authHandler := handlers.NewUserHandler(authService, logger)
	app.Post("/users/register", authHandler.CreateNewUser)
	app.Get("/health", authHandler.GetHealthStatus)
}

