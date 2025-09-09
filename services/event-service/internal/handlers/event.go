package handlers

import (
	"event-service/internal/dto"
	"event-service/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type EventHandler struct {
	eventService services.EventService
	logger       *zap.Logger
}

func NewEventHandler(eventService services.EventService, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		logger:       logger,
	}
}

func (h *EventHandler) CreateNewEvent(c *fiber.Ctx) error {
	var req dto.CreateNewEventDto
	// Bind the req body to the dto struct
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Something went wrong while trying to parse the request body",
		})
	}

	h.logger.Info("Request body", zap.Any("request", req))
	// Pass down control to the event service
	err := h.eventService.CreateNewEvent(req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully created event",
	})
}
