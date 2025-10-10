package routes

import (
	"event-service/internal/handlers"
	"event-service/internal/repository"
	"event-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, logger *zap.Logger) {
	api := app.Group("/api/v1")

	// Event repositories and services
	eventRepo := repository.NewEventRepository(db)
	outBoxRepo := repository.NewOutboxRepository(db)
	eventCategoryRepo := repository.NewEventCategoryRepository(db)
	eventVenueRepo := repository.NewEventVenueRepository(db)

	eventService := services.NewEventService(eventRepo, eventCategoryRepo, eventVenueRepo, outBoxRepo, db, logger)
	eventHandler := handlers.NewEventHandler(eventService, logger)

	// Category repositories and services
	categoryRepo := repository.NewEventCategoryRepository(db)
	categoryService := services.NewEventCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// Venue repositories and services
	venueRepo := repository.NewEventVenueRepository(db)
	venueService := services.NewEventVenueService(venueRepo)
	venueHandler := handlers.NewEventVenueHandler(venueService, logger)

	events := api.Group("/events")
	// Basic health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"service": "event-service",
		})
	})
	// Event routes
	events.Post("/", eventHandler.CreateNewEvent)
	events.Post("/:id/ticket-type", eventHandler.CreateTicketTypeForEvent)
	// Event category routes
	eventCategories := events.Group("/category")
	eventCategories.Post("/", categoryHandler.CreateCategory)
	eventCategories.Get("/", categoryHandler.GetCategories)
	eventCategories.Get("/:id", categoryHandler.GetCategoryByID)
	eventCategories.Put("/:id", categoryHandler.UpdateCategory)
	eventCategories.Delete("/:id", categoryHandler.DeleteCategory)

	// Event venue routes
	venueCategories := events.Group("/venues")
	venueCategories.Post("/", venueHandler.CreateVenue)
	venueCategories.Get("/", venueHandler.GetVenues)
	venueCategories.Get("/:id", venueHandler.GetVenueByID)
	venueCategories.Put("/:id", venueHandler.UpdateVenue)
	venueCategories.Delete("/:id", venueHandler.DeleteVenue)

	
}
