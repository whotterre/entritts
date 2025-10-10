package rabbitmq

import (
	"encoding/json"
	"event-service/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type TicketConsumer struct {
	conn      *amqp.Connection
	eventRepo repository.EventRepository
	logger    *zap.Logger
}

func NewTicketConsumer(conn *amqp.Connection, eventRepo repository.EventRepository, logger *zap.Logger) *TicketConsumer {
	return &TicketConsumer{conn: conn, eventRepo: eventRepo, logger: logger}
}

func (c *TicketConsumer) ConsumeTicketEvents() {
	ch, err := c.conn.Channel()
	if err != nil {
		c.logger.Error("Failed to open channel", zap.Error(err))
		return
	}
	defer ch.Close()
	
	if err := ch.ExchangeDeclare("tickets", "topic", true, false, false, false, nil); err != nil {
		c.logger.Error("Failed to declare exchange", zap.Error(err))
		return
	}

	// limit unacked messages to 1 while we process
	if err := ch.Qos(1, 0, false); err != nil {
		c.logger.Error("Failed to set QoS", zap.Error(err))
		return
	}

	queue, err := ch.QueueDeclare("ticket_events_queue", true, false, false, false, nil)
	if err != nil {
		c.logger.Error("Failed to declare queue", zap.Error(err))
		return
	}

	if err := ch.QueueBind(queue.Name, "ticket.created.pending", "tickets", false, nil); err != nil {
		c.logger.Error("Failed to bind queue", zap.Error(err))
		return
	}

	// use manual acks so we can Nack malformed messages and Ack processed ones
	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		c.logger.Error("Failed to start consuming", zap.Error(err))
		return
	}

	c.logger.Info("Listening for ticket events")
	for m := range msgs {
		var msg map[string]any
		if err := json.Unmarshal(m.Body, &msg); err != nil {
			c.logger.Error("invalid ticket message", zap.Error(err))
			// reject malformed payloads without requeueing
			if err := m.Nack(false, false); err != nil {
				c.logger.Error("failed to nack message", zap.Error(err))
			}
			continue
		}

		c.logger.Info("ticket event", zap.Any("payload", msg))

		if err := m.Ack(false); err != nil {
			c.logger.Error("failed to ack message", zap.Error(err))
		}
	}
}
