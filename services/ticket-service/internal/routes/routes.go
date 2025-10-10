package routes
import (
	"ticket-service/internal/handlers"
	"ticket-service/internal/rabbitmq"
	"ticket-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, 
	db *gorm.DB, ticketService services.TicketService,
	consumer *rabbitmq.TicketConsumer, 
	producer *rabbitmq.TicketProducer,
	logger *zap.Logger) {

	api := app.Group("/api/v1")
	ticketRoutes := api.Group("/tickets")

	if consumer != nil {
		ticketService.SetConsumer(consumer)
	}
	ticketHandler := handlers.NewTicketHandler(ticketService, logger)

	// Health check endpoint
	ticketRoutes.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"service": "ticket-service",
		})
	})
	
	ticketRoutes.Post("/", ticketHandler.CreateNewTicket)

}
