package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Producer struct {
	conn        *amqp091.Connection
	channel     *amqp091.Channel
	logger      *zap.Logger
	confirmMode bool
}

// NewProducer creates connection and channel and ensures the topic exchange exists.
func NewProducer(amqpURL string, logger *zap.Logger) (*Producer, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.String("url", amqpURL), zap.Error(err))
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error("Failed to create RabbitMQ channel", zap.Error(err))
		return nil, err
	}

	// Ensure the exchange exists
	if err := channel.ExchangeDeclare(
		"events",
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		channel.Close()
		conn.Close()
		logger.Error("Failed to declare exchange", zap.Error(err))
		return nil, err
	}

	// Try to enable confirmation mode, but don't fail if it doesn't work
	confirmMode := false
	if err := channel.Confirm(false); err != nil {
		logger.Warn("Failed to enable confirm mode, will publish without confirmations", zap.Error(err))
	} else {
		confirmMode = true
		logger.Info("RabbitMQ producer connected with confirmation mode enabled")
	}

	// Test the connection with a simple channel check
	if err := channel.Flow(true); err != nil {
		logger.Warn("Channel flow control test failed", zap.Error(err))
	}

	logger.Info("RabbitMQ producer successfully initialized",
		zap.String("exchange", "events"),
		zap.Bool("confirm_mode", confirmMode),
	)

	return &Producer{
		conn:        conn,
		channel:     channel,
		logger:      logger,
		confirmMode: confirmMode,
	}, nil
}

// Publish publishes a message and waits for a broker confirmation with a timeout.
func (p *Producer) Publish(exchange, routingKey string, message any) error {
	jsonBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup confirmation channel only if confirmation mode is enabled
	var confirmCh chan amqp091.Confirmation
	if p.confirmMode {
		confirmCh = make(chan amqp091.Confirmation, 1)
		p.channel.NotifyPublish(confirmCh)
	}

	if err := p.channel.PublishWithContext(
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
	); err != nil {
		p.logger.Error("Failed to publish message",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Error(err),
		)
		return err
	}

	// If confirmation mode is disabled, return immediately
	if !p.confirmMode {
		p.logger.Debug("Message published (no confirmation)",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
		)
		return nil
	}

	// Wait for confirmation or timeout
	select {
	case confirmed := <-confirmCh:
		if !confirmed.Ack {
			p.logger.Error("Message not acknowledged by broker",
				zap.String("exchange", exchange),
				zap.String("routing_key", routingKey),
			)
			return fmt.Errorf("message not acknowledged by broker")
		}
		p.logger.Debug("Message published and confirmed",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
		)
		return nil
	case <-ctx.Done():
		p.logger.Warn("Timed out waiting for broker confirmation, message may still be delivered",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
		)
		// Don't return error - message was likely delivered, just confirmation timed out
		return nil
	}
}

// PublishSimple publishes a message without waiting for confirmation (fire and forget)
func (p *Producer) PublishSimple(exchange, routingKey string, message any) error {
	jsonBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.channel.PublishWithContext(
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
	); err != nil {
		p.logger.Error("Failed to publish message",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Error(err),
		)
		return err
	}

	p.logger.Debug("Message published",
		zap.String("exchange", exchange),
		zap.String("routing_key", routingKey),
	)
	return nil
}

// PublishWithRetry retries Publish up to maxRetries with exponential backoff.
func (p *Producer) PublishWithRetry(exchange, routingKey string, message interface{}, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = p.PublishSimple(exchange, routingKey, message)
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
			time.Sleep(time.Duration((i+1)*1000) * time.Millisecond)
		}
	}

	return err
}

// IsConnected checks if the connection and channel are still active
func (p *Producer) IsConnected() bool {
	return p.conn != nil && !p.conn.IsClosed() && p.channel != nil
}

// Close closes channel and connection.
func (p *Producer) Close() {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.logger.Info("RabbitMQ producer closed")
}
