package services

import (
	"encoding/json"
	"errors"
	"event-service/internal/dto"
	"event-service/internal/models"
	"event-service/internal/pkg/utils"
	"event-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EventService interface {
	CreateNewEvent(eventData dto.CreateNewEventDto) (*dto.CreateNewEventResponse, error)
}

type eventService struct {
	eventRepository repository.EventRepository
	eventCategoryRepository repository.EventCategoryRepository
	eventVenueRepository repository.EventVenueRepository
	outboxRepo      repository.OutboxRepository
	db              *gorm.DB
	logger          *zap.Logger
}

func NewEventService(eventRepository repository.EventRepository, 
	eventCategoryRepository repository.EventCategoryRepository,
	eventVenueRepository repository.EventVenueRepository, 
	outboxRepo repository.OutboxRepository, 
	db *gorm.DB, logger *zap.Logger) EventService {
	return &eventService{
		eventRepository: eventRepository,
		eventCategoryRepository: eventCategoryRepository,
		eventVenueRepository: eventVenueRepository,
		outboxRepo:      outboxRepo,
		db:              db,
		logger:          logger,
	}
}

func (s *eventService) CreateNewEvent(eventData dto.CreateNewEventDto) (*dto.CreateNewEventResponse, error) {
	// Start database transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Ensure the start and end dates specified is not in the past
	currentDate := time.Now()
	if eventData.StartDate.Before(currentDate) {
		return nil, errors.New("Can't use past date for event start date")
	}

	if eventData.EndDate.Before(currentDate) {
		return nil, errors.New("Can't use past date for event end date")
	}

	// Ensure the end date is greater than the start date
	if !eventData.StartDate.Before(eventData.EndDate) {
		return nil, errors.New("End date must be on or after the event start date")
	}

	// Validate that the category exists
	categoryData, err := s.eventCategoryRepository.GetCategoryByID(eventData.CategoryId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if categoryData == nil {
		tx.Rollback()
		return nil, errors.New("invalid category: category does not exist")
	}

	// Validate venue if provided
	if eventData.VenueId != uuid.Nil && eventData.VenueId != (uuid.UUID{}) {
		venueData, err := s.eventVenueRepository.GetVenueByID(eventData.VenueId)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if venueData == nil {
			tx.Rollback()
			return nil, errors.New("invalid venue: venue does not exist")
		}
	}

	// Create event in 'PUBLISHED' state
	newEvent := models.Event{
		OrganizerId: utils.StringToUUIDFormat(eventData.OrganizerId),
		Title:       eventData.Title,
		Description: eventData.Description,
		CategoryId:  eventData.CategoryId,
		StartDate:   eventData.StartDate,
		EndDate:     eventData.EndDate,
		Status:      models.EventStatus("PUBLISHED"),
	}

	// Create event in the same transaction
	if err := tx.Create(&newEvent).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create outbox event in the same transaction
	eventDataJson, err := json.Marshal(map[string]interface{}{
		"event_id":     newEvent.EventId.String(),
		"organizer_id": newEvent.OrganizerId.String(),
		"event_title":  newEvent.Title,
		"ticket_types": eventData.TicketTypes,
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = s.outboxRepo.CreateOutboxEvent(tx, newEvent.EventId.String(), "event.created", string(eventDataJson))
	if err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create outbox event", zap.Error(err))
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Load the complete event with relationships for the response
	var eventWithRelations models.Event
	err = s.db.Preload("Category").Preload("Venue").First(&eventWithRelations, "event_id = ?", newEvent.EventId).Error
	if err != nil {
		s.logger.Warn("Failed to load event relationships, using basic event data", zap.Error(err))
		eventWithRelations = newEvent
	}

	s.logger.Info("Event created and outbox event stored", zap.String("event_id", eventWithRelations.EventId.String()))

	// Build response DTO
	response := dto.CreateNewEventResponse{
		EventId:     eventWithRelations.EventId,
		Title:       eventWithRelations.Title,
		Description: eventWithRelations.Description,
		StartDate:   eventWithRelations.StartDate,
		EndDate:     eventWithRelations.EndDate,
		Status:      string(eventWithRelations.Status),
		CreatedAt:   eventWithRelations.CreatedAt,
	}

	// Only include venue if one is set and loaded
	if eventWithRelations.VenueId != nil && eventWithRelations.Venue != nil {
		response.Venue = &models.EventVenue{
			VenueID:      eventWithRelations.Venue.VenueID,
			VenueName:    eventWithRelations.Venue.VenueName,
			VenueAddress: eventWithRelations.Venue.VenueAddress,
			City:         eventWithRelations.Venue.City,
			State:        eventWithRelations.Venue.State,
			Country:      eventWithRelations.Venue.Country,
			Latitude:     eventWithRelations.Venue.Latitude,
			Longitude:    eventWithRelations.Venue.Longitude,
			Capacity:     eventWithRelations.Venue.Capacity,
			CreatedAt:    eventWithRelations.Venue.CreatedAt,
			UpdatedAt:    eventWithRelations.Venue.UpdatedAt,
		}
	}
	return &response, nil
}
