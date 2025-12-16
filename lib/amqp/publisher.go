package amqp

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher is a message publisher for AMQP messages.
type Publisher struct {
	conn *Connection
	ch   *amqp.Channel
}

// NewPublisher creates a new AMQP publisher.
func NewPublisher(con *Connection) (*Publisher, error) {
	ch, err := con.Connection().Channel()
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn: con,
		ch:   ch,
	}, nil
}

// Publish sends a message to the specified exchange and routing key.
func (p *Publisher) Publish(_ context.Context, exchange, routingKey string, msg amqp.Publishing) error {
	err := p.ch.Publish(exchange, routingKey, false, false, msg)
	if err != nil {
		return err
	}

	return nil
}
