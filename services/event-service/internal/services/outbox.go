package services

import (
	"encoding/json"
	"event-service/internal/models"
	"event-service/internal/rabbitmq"
	"event-service/internal/repository"
	"time"

	"go.uber.org/zap"
)

type OutboxService interface {
	PublishPendingEvents() error
	StartOutboxProcessor(interval time.Duration, stopCh <-chan struct{})
}

type outboxService struct {
	outboxRepo    repository.OutboxRepository
	eventProducer *rabbitmq.EventProducer
	logger        *zap.Logger
}

func NewOutboxService(
	outboxRepo repository.OutboxRepository,
	eventProducer *rabbitmq.EventProducer,
	logger *zap.Logger,
) OutboxService {
	return &outboxService{
		outboxRepo:    outboxRepo,
		eventProducer: eventProducer,
		logger:        logger,
	}
}

func (s *outboxService) PublishPendingEvents() error {
	events, err := s.outboxRepo.GetUnpublishedEvents()
	if err != nil {
		s.logger.Error("Failed to get unpublished events", zap.Error(err))
		return err
	}

	for _, event := range events {
		err := s.publishEvent(event)
		if err != nil {
			s.logger.Error("Failed to publish event",
				zap.String("event_id", event.ID.String()),
				zap.Error(err))
			continue
		}

		err = s.outboxRepo.MarkAsPublished(event.ID)
		if err != nil {
			s.logger.Error("Failed to mark event as published",
				zap.String("event_id", event.ID.String()),
				zap.Error(err))
		}
	}

	return nil
}

func (s *outboxService) publishEvent(event models.OutboxEvent) error {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(event.EventData), &eventData); err != nil {
		return err
	}

	switch event.EventType {
	case "event.created":
		return s.eventProducer.PublishEventCreated(
			event.AggregateID,
			getString(eventData, "organizer_id"),
			getString(eventData, "event_title"),
			getSlice(eventData, "ticket_types"),
		)
	case "event.updated":
		return s.eventProducer.PublishEventUpdated(
			event.AggregateID,
			getMap(eventData, "updates"),
		)
	case "event.deleted":
		return s.eventProducer.PublishEventDeleted(event.AggregateID)
	default:
		s.logger.Warn("Unknown event type", zap.String("event_type", event.EventType))
		return nil
	}
}

func (s *outboxService) StartOutboxProcessor(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Starting outbox processor", zap.Duration("interval", interval))

	for {
		select {
		case <-ticker.C:
			if err := s.PublishPendingEvents(); err != nil {
				s.logger.Error("Error processing outbox events", zap.Error(err))
			}
		case <-stopCh:
			s.logger.Info("Stopping outbox processor")
			return
		}
	}
}

// Helper functions to extract values from map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getSlice(data map[string]interface{}, key string) []interface{} {
	if val, ok := data[key].([]interface{}); ok {
		return val
	}
	return nil
}

func getMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key].(map[string]interface{}); ok {
		return val
	}
	return nil
}
