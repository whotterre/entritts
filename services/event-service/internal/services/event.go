package services

import (
	"event-service/internal/dto"
	"event-service/internal/models"
	"event-service/internal/pkg/utils"
	"event-service/internal/rabbitmq"
	"event-service/internal/repository"

	"go.uber.org/zap"
)

type EventService interface {
	CreateNewEvent(eventData dto.CreateNewEventDto) (*models.Event, error)
}

type eventService struct {
	eventRepository repository.EventRepository
	eventProducer   rabbitmq.EventProducer
	logger          *zap.Logger
}

func NewEventService(eventRepository repository.EventRepository, eventProducer rabbitmq.EventProducer, logger *zap.Logger) EventService {
	return &eventService{
		eventRepository: eventRepository,
		eventProducer:   eventProducer,
		logger:          logger,
	}
}

func (s *eventService) CreateNewEvent(eventData dto.CreateNewEventDto) (*models.Event, error) {
	newEvent := models.Event{
		OrganizerId: utils.StringToUUIDFormat(eventData.OrganizerId),
		Title:       eventData.Title,
		Description: eventData.Description,
		CategoryId:  eventData.CategoryId,
		StartDate:   eventData.StartDate,
		EndDate:     eventData.EndDate,
		Status:      models.EventStatus(eventData.Status),
	}

	event, err := s.eventRepository.Create(&newEvent)
	if err != nil {
		return nil, err
	}

	return event, nil
}
