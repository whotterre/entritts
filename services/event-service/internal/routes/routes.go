package routes

import (
	"event-service/internal/rabbitmq"
	"event-service/internal/repository"
	"event-service/internal/services"
	"event-service/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, logger *zap.Logger, queue *rabbitmq.EventProducer) {
	api := app.Group("/api/v1")

	eventRepo := repository.NewEventRepository(db)
	eventService := services.NewEventService(eventRepo, *queue, logger)
	eventHandler := handlers.NewEventHandler(eventService, logger)


	// Basic health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"service": "event-service",
		})
	})

	// Service specific routes
	api.Post("/", eventHandler.CreateNewEvent)
}
