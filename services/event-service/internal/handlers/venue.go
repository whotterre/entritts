package handlers

import (
	"event-service/internal/dto"
	"event-service/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EventVenueHandler struct {
	eventVenueService services.EventVenueService
	logger            *zap.Logger
}

func NewEventVenueHandler(eventVenueService services.EventVenueService, logger *zap.Logger) *EventVenueHandler {
	return &EventVenueHandler{
		eventVenueService: eventVenueService,
		logger:            logger,
	}
}

func (h *EventVenueHandler) CreateVenue(c *fiber.Ctx) error {
	var request dto.CreateVenueRequest

	if err := c.BodyParser(&request); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	venue, err := h.eventVenueService.CreateVenue(request)
	if err != nil {
		h.logger.Error("Failed to create venue", zap.Error(err))
		status := http.StatusInternalServerError
		if err.Error() == "venue name is required" ||
			err.Error() == "venue with this name already exists" ||
			err.Error() == "venue address is required" ||
			err.Error() == "venue capacity must be greater than 0" {
			status = http.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to create venue",
			"details": err.Error(),
		})
	}

	response := dto.VenueResponse{
		VenueID:      venue.VenueID.String(),
		VenueName:    venue.VenueName,
		VenueAddress: venue.VenueAddress,
		City:         venue.City,
		State:        venue.State,
		Country:      venue.Country,
		Capacity:     venue.Capacity,
		Latitude:     venue.Latitude,
		Longitude:    venue.Longitude,
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Venue created successfully",
		"venue":   response,
	})
}

func (h *EventVenueHandler) GetVenues(c *fiber.Ctx) error {
	venues, err := h.eventVenueService.GetAllVenues()
	if err != nil {
		h.logger.Error("Failed to retrieve venues", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve venues",
			"details": err.Error(),
		})
	}

	venueResponses := make([]dto.VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = dto.VenueResponse{
			VenueID:      venue.VenueID.String(),
			VenueName:    venue.VenueName,
			VenueAddress: venue.VenueAddress,
			City:         venue.City,
			State:        venue.State,
			Country:      venue.Country,
			Capacity:     venue.Capacity,
			Latitude:     venue.Latitude,
			Longitude:    venue.Longitude,
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"venues": venueResponses,
		"total":  len(venueResponses),
	})
}

func (h *EventVenueHandler) GetVenueByID(c *fiber.Ctx) error {
	venueIDStr := c.Params("id")
	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		h.logger.Error("Invalid venue ID format", zap.String("venue_id", venueIDStr))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid venue ID",
			"details": "Venue ID must be a valid UUID",
		})
	}

	venue, err := h.eventVenueService.GetVenueByID(venueID)
	if err != nil {
		h.logger.Error("Failed to retrieve venue", zap.Error(err))
		status := http.StatusInternalServerError
		if err.Error() == "venue not found" {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to retrieve venue",
			"details": err.Error(),
		})
	}

	response := dto.VenueResponse{
		VenueID:      venue.VenueID.String(),
		VenueName:    venue.VenueName,
		VenueAddress: venue.VenueAddress,
		City:         venue.City,
		State:        venue.State,
		Country:      venue.Country,
		Capacity:     venue.Capacity,
		Latitude:     venue.Latitude,
		Longitude:    venue.Longitude,
	}

	return c.Status(http.StatusOK).JSON(response)
}

func (h *EventVenueHandler) UpdateVenue(c *fiber.Ctx) error {
	venueIDStr := c.Params("id")
	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		h.logger.Error("Invalid venue ID format", zap.String("venue_id", venueIDStr))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid venue ID",
			"details": "Venue ID must be a valid UUID",
		})
	}

	var request dto.UpdateVenueRequest
	if err := c.BodyParser(&request); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	venue, err := h.eventVenueService.UpdateVenue(venueID, request)
	if err != nil {
		h.logger.Error("Failed to update venue",
			zap.String("venue_id", venueIDStr),
			zap.Error(err))
		status := http.StatusInternalServerError
		if err.Error() == "venue not found" {
			status = http.StatusNotFound
		} else if err.Error() == "venue name is required" ||
			err.Error() == "venue with this name already exists" ||
			err.Error() == "venue address is required" ||
			err.Error() == "venue capacity must be greater than 0" {
			status = http.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to update venue",
			"details": err.Error(),
		})
	}

	response := dto.VenueResponse{
		VenueID:      venue.VenueID.String(),
		VenueName:    venue.VenueName,
		VenueAddress: venue.VenueAddress,
		City:         venue.City,
		State:        venue.State,
		Country:      venue.Country,
		Capacity:     venue.Capacity,
		Latitude:     venue.Latitude,
		Longitude:    venue.Longitude,
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Venue updated successfully",
		"venue":   response,
	})
}


func (h *EventVenueHandler) DeleteVenue(c *fiber.Ctx) error {
	venueIDStr := c.Params("id")
	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		h.logger.Error("Invalid venue ID format", zap.String("venue_id", venueIDStr))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid venue ID",
			"details": "Venue ID must be a valid UUID",
		})
	}

	err = h.eventVenueService.DeleteVenue(venueID)
	if err != nil {
		h.logger.Error("Failed to delete venue",
			zap.String("venue_id", venueIDStr),
			zap.Error(err))
		status := http.StatusInternalServerError
		if err.Error() == "venue not found" {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{
			"error":   "Failed to delete venue",
			"details": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Venue deleted successfully",
	})
}
