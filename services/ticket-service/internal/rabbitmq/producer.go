package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type TicketProducer struct {
	conn *amqp.Connection
	logger *zap.Logger 
}

func NewTicketProducer(conn *amqp.Connection, logger *zap.Logger) *TicketProducer {
	return &TicketProducer{conn: conn, logger: logger}
}

func (p *TicketProducer) PublishEvent(exchange, routingKey string, body []byte) error {
	ch, err := p.conn.Channel()
	if err != nil {
		p.logger.Error("Failed to create RabbitMQ channel because%v", zap.Error(err))
		return err
	}
	defer ch.Close()

	err = ch.Publish(
		exchange,   
		routingKey, 
		false,      
		false, 
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		p.logger.Error("Failed to publish message: %v", zap.Error(err))
		return err
	}
	return nil
}
