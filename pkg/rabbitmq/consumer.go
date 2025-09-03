package rabbitmq

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer handles message consumption from RabbitMQ
type Consumer struct {
	Client *Client
}

// NewConsumer creates a new consumer from a client
func NewConsumer(client *Client) *Consumer {
	return &Consumer{Client: client}
}

// MessageHandler defines the function signature for processing messages
type MessageHandler func(ctx context.Context, delivery amqp.Delivery) error

// ConsumeOptions provides configuration for message consumption
type ConsumeOptions struct {
	QueueName     string
	ConsumerTag   string
	AutoAck       bool
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
	Args          amqp.Table
	PrefetchCount int
	PrefetchSize  int
}

// Consume starts consuming messages from a queue
func (c *Consumer) Consume(ctx context.Context, opts ConsumeOptions, handler MessageHandler) error {
	// Set QoS (Quality of Service) to control message prefetching
	err := c.Client.Channel.Qos(
		opts.PrefetchCount, // prefetch count
		opts.PrefetchSize,  // prefetch size
		false,              // global
	)
	if err != nil {
		log.Printf("Failed to set QoS: %v", err)
		return ErrConsumeFailed
	}

	// Start consuming messages
	deliveries, err := c.Client.Channel.Consume(
		opts.QueueName,   // queue
		opts.ConsumerTag, // consumer tag
		opts.AutoAck,     // auto-ack
		opts.Exclusive,   // exclusive
		opts.NoLocal,     // no-local
		opts.NoWait,      // no-wait
		opts.Args,        // args
	)
	if err != nil {
		log.Printf("Failed to start consumer: %v", err)
		return ErrConsumeFailed
	}

	log.Printf("Started consumer %s on queue %s", opts.ConsumerTag, opts.QueueName)

	// Process messages in a goroutine
	go c.processMessages(ctx, deliveries, handler, opts)

	return nil
}

// processMessages handles incoming messages and routes them to the handler
func (c *Consumer) processMessages(ctx context.Context, deliveries <-chan amqp.Delivery, handler MessageHandler, opts ConsumeOptions) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Consumer %s stopped by context", opts.ConsumerTag)
			return

		case delivery, ok := <-deliveries:
			if !ok {
				log.Printf("Delivery channel closed for consumer %s", opts.ConsumerTag)
				return
			}

			// Process the message in a goroutine to handle concurrent processing
			go c.handleMessage(ctx, delivery, handler, opts)
		}
	}
}

// handleMessage processes a single message with error handling and retry logic
func (c *Consumer) handleMessage(ctx context.Context, delivery amqp.Delivery, handler MessageHandler, opts ConsumeOptions) {
	start := time.Now()

	// Create a context with timeout for this message processing
	msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := handler(msgCtx, delivery)
	if err != nil {
		log.Printf("Message processing failed: %v (MessageID: %s)", err, delivery.MessageId)

		// Implement retry logic or dead letter queue handling here
		if !opts.AutoAck {
			// Reject and requeue the message for retry
			delivery.Reject(true)
		}
		return
	}

	// Acknowledge the message if not auto-ack
	if !opts.AutoAck {
		delivery.Ack(false)
	}

	log.Printf("Message processed successfully in %v (MessageID: %s)",
		time.Since(start), delivery.MessageId)
}

// CreateConsumerWithRetry creates a consumer with retry logic for resilience
func (c *Consumer) CreateConsumerWithRetry(ctx context.Context, opts ConsumeOptions, handler MessageHandler, maxRetries int) error {
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = c.Consume(ctx, opts, handler)
		if err == nil {
			return nil
		}

		log.Printf("Failed to create consumer (attempt %d/%d): %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	return err
}

// BindConsumer binds a consumer to an exchange with routing key
func (c *Consumer) BindConsumer(queueName, routingKey, exchangeName string) error {
	err := c.Client.BindQueue(queueName, routingKey, exchangeName)
	if err != nil {
		log.Printf("Failed to bind queue %s to exchange %s: %v", queueName, exchangeName, err)
		return err
	}

	log.Printf("Bound queue %s to exchange %s with routing key %s",
		queueName, exchangeName, routingKey)
	return nil
}
