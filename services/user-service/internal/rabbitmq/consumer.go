package rabbitmq

import (
	"context"
	"encoding/json"
	"time"
	"user-service/internal/services"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/whotterre/entritts/pkg/rabbitmq"
	"go.uber.org/zap"
)

type UserEventCustomer struct {
	userService services.UserService
	logger      *zap.Logger
}

// EventMessage represents the structure of event messages
type EventMessage struct {
	EventID     string    `json:"event_id"`
	Action      string    `json:"action"`
	OrganizerID string    `json:"organizer_id"`
	EventTitle  string    `json:"event_title"`
	Timestamp   time.Time `json:"timestamp"`
}

func NewUserEventConsumer(userService services.UserService, logger *zap.Logger) *UserEventCustomer {
	return &UserEventCustomer{
		userService: userService,
		logger:      logger,
	}
}

// HandleEventMessage processes incoming event messages
func (c *UserEventCustomer) HandleEventMessage(ctx context.Context, msg amqp091.Delivery) error {
	startTime := time.Now()

	// Log the incoming message
	c.logger.Info("Processing event message",
		zap.String("event_id", ""),
		zap.String("action", ""),
		zap.String("routing_key", msg.RoutingKey),
	)

	// Parse the message body
	var eventMsg EventMessage
	if err := json.Unmarshal(msg.Body, &eventMsg); err != nil {
		c.logger.Error("Failed to unmarshal event message",
			zap.Error(err),
			zap.String("message_body", string(msg.Body)),
		)
		return msg.Nack(false, false) 
	}

	// Update log with parsed event details
	c.logger.Info("Processing event message",
		zap.String("event_id", eventMsg.EventID),
		zap.String("action", eventMsg.Action),
		zap.String("routing_key", msg.RoutingKey),
	)

	// Route message based on action
	switch eventMsg.Action {
	case "created":
		if err := c.handleEventCreated(ctx, eventMsg); err != nil {
			c.logger.Error("Failed to process event created",
				zap.String("event_id", eventMsg.EventID),
				zap.Error(err),
			)
			return msg.Nack(false, true) // Requeue for retry
		}
	case "updated":
		if err := c.handleEventUpdated(ctx, eventMsg); err != nil {
			c.logger.Error("Failed to process event updated",
				zap.String("event_id", eventMsg.EventID),
				zap.Error(err),
			)
			return msg.Nack(false, true)
		}
	case "deleted":
		if err := c.handleEventDeleted(ctx, eventMsg); err != nil {
			c.logger.Error("Failed to process event deleted",
				zap.String("event_id", eventMsg.EventID),
				zap.Error(err),
			)
			return msg.Nack(false, true) 
		}
	default:
		c.logger.Warn("Unknown event action",
			zap.String("action", eventMsg.Action),
			zap.String("event_id", eventMsg.EventID),
		)
		return msg.Ack(false) 
	}

	// Acknowledge successful processing
	processingTime := time.Since(startTime)
	c.logger.Info("Message processed successfully",
		zap.Duration("processing_time", processingTime),
		zap.String("message_id", msg.MessageId),
	)

	return msg.Ack(false)
}

// handleEventCreated processes event creation messages
func (c *UserEventCustomer) handleEventCreated(ctx context.Context, msg EventMessage) error {
	c.logger.Info("Received event created message",
		zap.String("event_id", msg.EventID),
		zap.String("organizer_id", msg.OrganizerID),
		zap.String("event_title", msg.EventTitle),
	)

	// Validate organizer ID format
	_, err := uuid.Parse(msg.OrganizerID)
	if err != nil {
		c.logger.Error("Invalid organizer ID format",
			zap.String("organizer_id", msg.OrganizerID),
			zap.Error(err),
		)
		return err
	}

	// Get user by ID
	user, err := c.userService.GetUserByID(ctx, msg.OrganizerID)
	if err != nil {
		c.logger.Warn("Organizer not found, skipping event processing",
			zap.String("organizer_id", msg.OrganizerID),
			zap.String("event_id", msg.EventID),
		)
		return nil
	}

	// Check if user is nil (additional safety check)
	if user == nil {
		c.logger.Warn("User is nil, skipping event processing",
			zap.String("organizer_id", msg.OrganizerID),
			zap.String("event_id", msg.EventID),
		)
		return nil
	}

	// Increment user's event count
	err = c.userService.IncrementUserEventCount(ctx, user.ID.String())
	if err != nil {
		c.logger.Error("Failed to increment user event count",
			zap.String("user_id", user.ID.String()),
			zap.String("event_id", msg.EventID),
			zap.Error(err),
		)
		return err
	}

	if err := c.publishEventProcessedConfirmation(msg.EventID, "event_count_incremented"); err != nil {
		c.logger.Warn("Failed to publish confirmation event",
			zap.String("event_id", msg.EventID),
			zap.Error(err),
		)
	}

	c.logger.Info("Successfully processed event created",
		zap.String("event_id", msg.EventID),
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

// handleEventUpdated processes event update messages
func (c *UserEventCustomer) handleEventUpdated(ctx context.Context, msg EventMessage) error {
	c.logger.Info("Processing event updated",
		zap.String("event_id", msg.EventID),
		zap.String("organizer_id", msg.OrganizerID),
	)

	// For now, just log the update - you can add specific business logic here
	c.logger.Info("Event updated processed successfully",
		zap.String("event_id", msg.EventID),
	)

	return nil
}

// handleEventDeleted processes event deletion messages
func (c *UserEventCustomer) handleEventDeleted(ctx context.Context, msg EventMessage) error {
	c.logger.Info("Processing event deleted",
		zap.String("event_id", msg.EventID),
		zap.String("organizer_id", msg.OrganizerID),
	)

	// Validate organizer ID format
	_, err := uuid.Parse(msg.OrganizerID)
	if err != nil {
		c.logger.Error("Invalid organizer ID format",
			zap.String("organizer_id", msg.OrganizerID),
			zap.Error(err),
		)
		return err
	}

	// Get user by ID
	user, err := c.userService.GetUserByID(ctx, msg.OrganizerID)
	if err != nil {
		c.logger.Warn("Organizer not found, skipping event processing",
			zap.String("organizer_id", msg.OrganizerID),
			zap.String("event_id", msg.EventID),
		)
		return nil
	}

	// Check if user is nil (additional safety check)
	if user == nil {
		c.logger.Warn("User is nil, skipping event processing",
			zap.String("organizer_id", msg.OrganizerID),
			zap.String("event_id", msg.EventID),
		)
		return nil
	}

	// You could decrement event count here if needed
	c.logger.Info("Event deleted processed successfully",
		zap.String("event_id", msg.EventID),
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

// publishEventProcessedConfirmation publishes a confirmation message back to the event service
func (c *UserEventCustomer) publishEventProcessedConfirmation(eventID, status string) error {
	c.logger.Info("Would publish event processed confirmation",
		zap.String("event_id", eventID),
		zap.String("status", status),
	)

	return nil
}

// ConsumeEvents is a higher-level method that could be used to start consuming
func (c *UserEventCustomer) ConsumeEvents(ctx context.Context, consumer *rabbitmq.Consumer, queueName string) error {
	c.logger.Info("Starting event consumption",
		zap.String("queue", queueName),
	)

	consumeOpts := rabbitmq.ConsumeOptions{
		QueueName:     queueName,
		ConsumerTag:   "user-service-consumer",
		AutoAck:       false, 
		Exclusive:     false,
		NoLocal:       false,
		NoWait:        false,
		PrefetchCount: 10,
		PrefetchSize:  0,
	}

	return consumer.Consume(ctx, consumeOpts, c.HandleEventMessage)
}
