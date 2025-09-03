package rabbitmq

import (
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// Initialize RabbitMQ client
func NewClient(cfg Config, logger *log.Logger) (*Client, error) {
	conn, err := connectWithRetry(cfg.URL, 5, 5*time.Second, logger)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	logger.Println("Successfully connected to RabbitMQ")
	return &Client{Conn: conn, Channel: ch}, nil
}

// EnsureExchange declares an exchange if it doesn't exist
func (c *Client) EnsureExchange(name, kind string) error {
	return c.Channel.ExchangeDeclare(
		name,  // exchange name
		kind,  // exchange type: direct, fanout, topic, headers
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// Attempts to connect with retries
func connectWithRetry(url string,
	maxRetries int,
	delay time.Duration,
	logger *log.Logger,
) (*amqp091.Connection, error) {
	var conn *amqp091.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp091.Dial(url)
		if err == nil {
			return conn, nil
		}
		logger.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return nil, err
}

func (c *Client) Close(logger *log.Logger) {
	if c.Channel != nil {
		c.Channel.Close()
		logger.Println("RabbitMQ channel closed")
	}

	if c.Conn != nil {
		c.Conn.Close()
		logger.Println("RabbitMQ connection closed")
	}
}

// Creates a new message queue if one doesn't exists
func (c *Client) EnsureQueue(name string) (amqp091.Queue, error) {
	return c.Channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
}

// Binds a queue to an exchange with a routing key
func (c *Client) BindQueue(queueName, routingKey, exchangeName string) error {
	return c.Channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
}
