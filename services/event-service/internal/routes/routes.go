package routes

import (
	"event-service/internal/handlers"
	"event-service/internal/rabbitmq"
	"event-service/internal/repository"
	"event-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, logger *zap.Logger, queue *rabbitmq.EventProducer) {
	api := app.Group("/api/v1")

	// Event repositories and services
	eventRepo := repository.NewEventRepository(db)
	eventService := services.NewEventService(eventRepo, *queue, logger)
	eventHandler := handlers.NewEventHandler(eventService, logger)

	// Category repositories and services
	categoryRepo := repository.NewEventCategoryRepository(db)
	categoryService := services.NewEventCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// Basic health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"service": "event-service",
		})
	})

	// Category routes under /api/v1/events/category
	eventCategories := api.Group("/events/category")
	eventCategories.Post("/", categoryHandler.CreateCategory)
	eventCategories.Get("/", categoryHandler.GetCategories)
	eventCategories.Get("/:id", categoryHandler.GetCategoryByID)
	eventCategories.Put("/:id", categoryHandler.UpdateCategory)
	eventCategories.Delete("/:id", categoryHandler.DeleteCategory)

	// Event routes
	events := api.Group("/events")
	events.Post("/", eventHandler.CreateNewEvent)
}
