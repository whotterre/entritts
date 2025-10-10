package rabbitmq

import (
	"event-service/internal/dto"
	"time"

	"go.uber.org/zap"
)

type EventMessage struct {
	EventID     string `json:"event_id"`
	OrganizerID string `json:"organizer_id"`
	EventTitle  string `json:"event_title"`
	TicketTypes any    `json:"ticket_types,omitempty"`
	Updates     any    `json:"updates,omitempty"`
	Action      string `json:"action"`
	Timestamp   string `json:"timestamp"`
	Service     string `json:"service"`
}

type EventProducer struct {
	producer *Producer
	logger   *zap.Logger
}

func NewEventProducer(amqpURL string, logger *zap.Logger) (*EventProducer, error) {
	producer, err := NewProducer(amqpURL, logger)
	if err != nil {
		return nil, err
	}

	return &EventProducer{
		producer: producer,
		logger:   logger,
	}, nil
}

func (ep *EventProducer) PublishEventCreated(eventID, organizerID, title string, ticketTypes []any) error {
	message := map[string]any{
		"event_id":     eventID,
		"organizer_id": organizerID,
		"event_title":  title,
		"ticket_types": ticketTypes,
		"action":       "created",
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	// Use PublishWithRetry for better reliability in outbox processing
	return ep.producer.PublishWithRetry("events", "event.created", message, 3)
}

func (ep *EventProducer) PublishEventCreatedPending(eventID, organizerID, title string, ticketTypes []dto.TicketType) error {
	message := map[string]any{
		"event_id":     eventID,
		"organizer_id": organizerID,
		"event_title":  title,
		"ticket_types": ticketTypes,
		"action":       "created.pending",
		"timestamp":    time.Now().Format(time.RFC3339),
	}
	return ep.producer.PublishWithRetry("events", "event.created.pending", message, 3)
}

func (ep *EventProducer) PublishEventUpdated(eventID string, updates map[string]interface{}) error {
	message := map[string]interface{}{
		"event_id":  eventID,
		"updates":   updates,
		"action":    "updated",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return ep.producer.PublishWithRetry("events", "event.updated", message, 3)
}

func (ep *EventProducer) PublishEventDeleted(eventID string) error {
	message := map[string]interface{}{
		"event_id":  eventID,
		"action":    "deleted",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return ep.producer.PublishWithRetry("events", "event.deleted", message, 3)
}

func (ep *EventProducer) IsConnected() bool {
	return ep.producer.IsConnected()
}

func (ep *EventProducer) Close() {
	ep.producer.Close()
}

func (ep *EventProducer) PublishTicketCreate(payload map[string]any) error {
	return ep.producer.PublishWithRetry("tickets", "ticket.created.pending", payload, 3)
}
