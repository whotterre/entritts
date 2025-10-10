package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type TicketConsumer struct {
	conn          *amqp.Connection
	logger        *zap.Logger
	ticketService TicketService
}

type TicketService interface {
	ProcessTicketCreateMessage(ctx context.Context, payload []byte) error
}

func NewTicketConsumer(conn *amqp.Connection, logger *zap.Logger, svc TicketService) *TicketConsumer {
	return &TicketConsumer{conn: conn, logger: logger, ticketService: svc}
}

func (c *TicketConsumer) ConsumeEvents() []byte {
	ch, err := c.conn.Channel()
	if err != nil {
		c.logger.Error("Failed to open channel", zap.Error(err))
		return nil
	}
	defer ch.Close()

	// ensure main exchange exists
	if err := ch.ExchangeDeclare("tickets", "topic", true, false, false, false, nil); err != nil {
		c.logger.Error("Failed to declare/ensure exchange", zap.Error(err))
		return nil
	}

	// Declare a retry DLX and a retry queue with TTL that routes back to main queue
	dlxName := "tickets.dlx"
	retryQueue := "ticket_events_retry_queue"
	mainQueue := "ticket_events_queue"

	if err := ch.ExchangeDeclare(dlxName, "topic", true, false, false, false, nil); err != nil {
		c.logger.Error("Failed to declare DLX", zap.Error(err))
		return nil
	}

	// retry queue with short TTL then dead-letters to the main exchange (requeue)
	retryArgs := amqp.Table{
		"x-dead-letter-exchange":    "tickets",
		"x-dead-letter-routing-key": "ticket.created.pending",
		"x-message-ttl":             int32(5000),
	}

	if _, err := ch.QueueDeclare(retryQueue, true, false, false, false, retryArgs); err != nil {
		c.logger.Error("Failed to declare retry queue", zap.Error(err))
		return nil
	}

	queue, err := ch.QueueDeclare(mainQueue, true, false, false, false, nil)
	if err != nil {
		c.logger.Error("Failed to declare queue", zap.Error(err))
		return nil
	}

	if err := ch.QueueBind(
		queue.Name,
		"ticket.created.pending",
		"tickets",
		false,
		nil,
	); err != nil {
		c.logger.Error("Failed to bind queue", zap.Error(err))
		return nil
	}

	if err := ch.Qos(1, 0, false); err != nil {
		c.logger.Error("Failed to set QoS", zap.Error(err))
		return nil
	}

	msgs, err := ch.Consume(
		queue.Name,
		"ticket-service-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Error("Failed to start consuming", zap.Error(err))
		return nil
	}

	c.logger.Info("Started consuming ticket events")
	for msg := range msgs {
		// read message-id from AMQP property header or message header
		var messageID string
		if msg.MessageId != "" {
			messageID = msg.MessageId
		} else if v, ok := msg.Headers["message-id"]; ok {
			if s, ok := v.(string); ok {
				messageID = s
			}
		}

		// if payload doesn't contain message_id, inject it for service dedupe
		payload := msg.Body
		if messageID != "" {
			var p map[string]any
			if err := json.Unmarshal(msg.Body, &p); err == nil {
				if _, ok := p["message_id"]; !ok {
					p["message_id"] = messageID
					if b, err := json.Marshal(p); err == nil {
						payload = b
					}
				}
			}
		}

		if err := c.ticketService.ProcessTicketCreateMessage(context.Background(), payload); err != nil {
			if err.Error() == "permanent invalid payload" { // Use string comparison
				c.logger.Error("Permanent payload error, rejecting message", zap.Error(err))
				_ = msg.Nack(false, false)
				continue
			}
			c.logger.Error("Transient error processing ticket create message, sending to retry", zap.Error(err))
			// nack and send to retry DLX by rejecting with requeue=false; message will route to retry queue
			_ = msg.Nack(false, false)
			continue
		}

		if err := msg.Ack(false); err != nil {
			c.logger.Error("Failed to ack message after processing", zap.Error(err))
			_ = msg.Nack(false, true)
			continue
		}
	}
	return nil
}
