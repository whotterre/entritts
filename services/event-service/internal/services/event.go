package services

import (
	"encoding/json"
	"errors"
	"event-service/internal/dto"
	"event-service/internal/models"
	"event-service/internal/pkg/utils"
	"event-service/internal/repository"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EventService interface {
	CreateNewEvent(eventData dto.CreateNewEventDto) (*models.Event, error)
}

type eventService struct {
	eventRepository repository.EventRepository
	outboxRepo      repository.OutboxRepository
	db              *gorm.DB
	logger          *zap.Logger
}

func NewEventService(eventRepository repository.EventRepository, outboxRepo repository.OutboxRepository, db *gorm.DB, logger *zap.Logger) EventService {
	return &eventService{
		eventRepository: eventRepository,
		outboxRepo:      outboxRepo,
		db:              db,
		logger:          logger,
	}
}

func (s *eventService) CreateNewEvent(eventData dto.CreateNewEventDto) (*models.Event, error) {
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
	if eventData.EndDate.Before(eventData.StartDate) {
   	   return nil, errors.New("End date must be on or after the event start date")
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

	s.logger.Info("Event created and outbox event stored", zap.String("event_id", newEvent.EventId.String()))
	return &newEvent, nil
}
