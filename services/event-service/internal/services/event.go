package services

import (
	"event-service/internal/dto"
	"event-service/internal/models"
	"event-service/internal/pkg/utils"
	"event-service/internal/rabbitmq"
	"event-service/internal/repository"
	"time"

	"go.uber.org/zap"
)

type EventService interface {
	CreateNewEvent(eventData dto.CreateNewEventDto) error 
}

type eventService struct {
	eventRepository repository.EventRepository
	eventProducer rabbitmq.EventProducer
	logger *zap.Logger
}

func NewEventService(eventRepository repository.EventRepository, eventProducer rabbitmq.EventProducer, logger *zap.Logger) EventService {
	return &eventService{
		eventRepository: eventRepository,
		eventProducer: eventProducer,
		logger: logger,
	}
}

func (s *eventService) CreateNewEvent(eventData dto.CreateNewEventDto) error {
	// Call repository method
	newEvent := models.Event{
		OrganizerId: utils.StringToUUIDFormat(eventData.OrganizerId),
		Title: eventData.Title,
		Description: eventData.Description,
		CategoryId: eventData.CategoryId,
		StartDate: eventData.StartDate,
		EndDate: eventData.EndDate,
		Status: models.EventStatus(eventData.Status),
	}

	event, err := s.eventRepository.Create(&newEvent)
	if err != nil {
		return err
	}
	// Send RabbitMQ message to ticket service
	go func(){
		ticketTypesForMessage := s.prepareTicketTypesForMessage(eventData.TicketTypes)

		err := s.eventProducer.PublishEventCreated(event.EventId.String(), eventData.OrganizerId, eventData.Title, ticketTypesForMessage)
		if err != nil {
			s.logger.Error("Failed to publish event created message after retries",
				zap.String("event_id", event.EventId.String()),
				zap.Error(err),
			)
			// Consider storing failed messages for later retry
		} else {
			s.logger.Info("Event created message published successfully",
				zap.String("event_id", event.EventId.String()),
			)
		}
	}()

	return nil
}	


func (s *eventService) prepareTicketTypesForMessage(ticketTypes []dto.TicketType) []interface{} {
	var result []any
	for _, tt := range ticketTypes {
		result = append(result, map[string]interface{}{
			"name":         tt.Name,
			"description":  tt.Description,
			"price":        tt.Price,
			"quantity":     tt.TotalQuantity,
			"sale_start":   tt.SaleStartDate.Format(time.RFC3339),
			"sale_end":     tt.SaleEndDate.Format(time.RFC3339),
		})
	}
	return result
}