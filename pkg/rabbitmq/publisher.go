package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewPublisher(client *Client) *Publisher {
	return &Publisher{Client: client}
}

func (p *Publisher) Publish(event Event) error {
	ctx := context.Background()
	err := p.Client.EnsureExchange(event.Exchange, "topic")
	if err != nil {
		return err
	}

	message := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         event.Body,
		Timestamp:    time.Now(),
		Headers:      event.Headers,
		DeliveryMode: amqp091.Persistent,
	}

	err = p.Client.Channel.PublishWithContext(
		ctx,
		event.Exchange,
		event.RoutingKey,
		false,
		false,
		message,
	)

	if err != nil {
		return err
	}

	log.Printf("Published message to %s:%s", event.Exchange, event.RoutingKey)
	return nil
}

func (p *Publisher) PublishJSON(exchange, routingKey string, data any, headers amqp.Table) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	event := Event{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Body:       jsonData,
		Headers:    headers,
	}
	return p.Publish(event)
}

func (p *Publisher) PublishWithRetry(event Event, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = p.Publish(event)
		if err == nil {
			return nil
		}
		log.Printf("⚠️ Failed to publish (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries - 1 {
			time.Sleep(delay)
		}

	}
	return err
}

func (p *Publisher) PublishWithConfirmation(event Event) error {
	err := p.Client.Channel.Confirm(false)
	if err != nil {
		return err
	}

	confirms := p.Client.Channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	err = p.Publish(event)
	if err != nil {
		return err
	}

	if confirmed := <-confirms; confirmed.Ack {
		log.Printf("Message confirmed for %s:%s", event.Exchange, event.RoutingKey)
		return nil
	} else {
		log.Printf("Message not confirmed for %s:%s", event.Exchange, event.RoutingKey)
		return err
	}

}
