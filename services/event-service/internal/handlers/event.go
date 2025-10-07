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

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body",
			zap.Error(err),
			zap.String("content_type", c.Get("Content-Type")),
		)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INVALID_REQUEST_BODY",
				"message": "Invalid request format. Please check your JSON payload.",
				"details": "The request body could not be parsed as valid JSON",
			},
			"timestamp": c.Context().Time(),
		})
	}

	h.logger.Info("Processing event creation request",
		zap.String("title", req.Title),
		zap.String("organizer_id", req.OrganizerId),
	)

	// Create event via service
	newEvent, err := h.eventService.CreateNewEvent(req)
	if err != nil {
		h.logger.Error("Failed to create event",
			zap.Error(err),
			zap.String("title", req.Title),
			zap.String("organizer_id", req.OrganizerId),
		)

		// Determine error type and response
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		message := "An unexpected error occurred while creating the event"

		// Handle specific error types
		switch {
		case err.Error() == "Can't use past date for event start date" ||
			err.Error() == "Can't use past date for event end date" ||
			err.Error() == "End date must be on or after the event start date":
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_DATE"
			message = "Invalid event dates provided"

		case err.Error() == "Invalid category: category does not exist":
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_CATEGORY"
			message = "The specified category does not exist"

		case err.Error() == "Invalid venue: venue does not exist":
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_VENUE"
			message = "The specified venue does not exist"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    errorCode,
				"message": message,
				"details": err.Error(),
			},
			"timestamp": c.Context().Time(),
		})
	}

	h.logger.Info("Event created successfully",
		zap.String("event_id", newEvent.EventId.String()),
		zap.String("title", newEvent.Title),
	)

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Event created successfully",
		"data": fiber.Map{
			"event": newEvent,
		},
		"timestamp": c.Context().Time(),
	})
}
