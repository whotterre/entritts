package handlers

import (
	"net/http"
	"ticket-service/internal/dto"
	"ticket-service/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type TicketHandler interface {

}

type ticketHandler struct {
	ticketService services.TicketService
	logger *zap.Logger
}

func NewTicketHandler(ticketService services.TicketService, logger *zap.Logger) *ticketHandler{
	return &ticketHandler{ticketService: ticketService, logger: logger}
}


func (h *ticketHandler) CreateNewTicket(c *fiber.Ctx) error {
	// Parse request body
	var input dto.CreateNewTicketDto
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Failed to parse request body for ticket service", zap.Error(err))
	}

	// Pass it down to the service
	response, err := h.ticketService.CreateNewTicket(input)
	if err != nil {
		h.logger.Error("Failed to create new ticket", zap.Error(err))
	}

	c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Successfully created new ticket",
		"ticket": response,
	})
	return nil
}




