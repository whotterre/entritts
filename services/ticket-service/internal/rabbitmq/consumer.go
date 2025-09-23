package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type TicketConsumer struct {
	conn   *amqp.Connection
	logger *zap.Logger
}

func NewTicketConsumer(conn *amqp.Connection, logger *zap.Logger) *TicketConsumer {
	return &TicketConsumer{conn: conn, logger: logger}
}

func (c *TicketConsumer) ConsumeEvents() []byte {
	ch, err := c.conn.Channel()
	if err != nil {
		c.logger.Error("Failed to open channel", zap.Error(err))
		return nil
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"ticket_events_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Error("Failed to declare queue", zap.Error(err))
		return nil
	}

	err = ch.QueueBind(
		queue.Name,
		"ticket.created.pending",
		"tickets",
		false,
		nil,
	)
	if err != nil {
		c.logger.Error("Failed to bind queue", zap.Error(err))
		return nil
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
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
		c.logger.Info("Received message", zap.ByteString("body", msg.Body))
		return msg.Body
	}
	return nil
}
