package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Producer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	logger  *zap.Logger
}

func NewProducer(amqpURL string, logger *zap.Logger) (*Producer, error) {
	// Create a connection to the broker 
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	// Create a channel from the connection to process the AMQP messages
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	// Create an exchange with custom options if it doesn't already exist
	err = channel.ExchangeDeclare(
		"events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}
	logger.Info("Rabbit MQ producer connected successfully")
	return &Producer{
		conn:    conn,
		channel: channel,
		logger:  logger,
	}, nil

}

func (p *Producer) Publish(exchange, routingKey string, message any) error {
	jsonBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         jsonBody,
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		p.logger.Error("Failed to publish message",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Error(err),
		)
		return err
	}

	confirmed := <-p.channel.NotifyPublish(make(chan amqp091.Confirmation, 1))
	if !confirmed.Ack {
		p.logger.Error("Message not confirmed by broker",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
		)
		return amqp091.ErrClosed
	}
	p.logger.Debug("Message published successfully",
		zap.String("exchange", exchange),
		zap.String("routing_key", routingKey),
	)
	return nil
}

func (p *Producer) PublishWithRetry(exchange, routingKey string, message interface{}, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = p.Publish(exchange, routingKey, message)
		if err == nil {
			return nil
		}

		p.logger.Warn("Publish failed, retrying",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Int("attempt", i+1),
			zap.Int("max_retries", maxRetries),
			zap.Error(err),
		)

		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
		}
	}

	return err
}

func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	p.logger.Info("RabbitMQ producer closed")
}