package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

// Client holds RabbitMQ connection and channel
type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// Publisher wraps the client with publishing capabilities
type Publisher struct {
	Client *Client
}

// Event represents a message to be published
type Event struct {
	Exchange   string
	RoutingKey string
	Body       []byte
	Headers    amqp.Table
}

// Config holds RabbitMQ connection configuration
type Config struct {
	URL string
}

// ConsumerConfig holds configuration for message consumers
type ConsumerConfig struct {
	QueueName    string
	ConsumerName string
	AutoAck      bool
	Exclusive    bool
	NoLocal      bool
	NoWait       bool
	Args         amqp.Table
}