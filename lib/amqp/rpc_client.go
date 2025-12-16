package amqp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RPCClient is an AMQP RPC client.
type RPCClient struct {
	con     *Connection
	options *rpcClientOptions
}

// NewRPCClient creates a new RPC client.
func NewRPCClient(con *Connection, opts ...RPCClientOptionFunc) (*RPCClient, error) {
	o := defaultRPCClientOptions()
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	return &RPCClient{
		con:     con,
		options: o,
	}, nil
}

// Call sends a request to the specified exchange and routing key and waits for a response.
// nolint cyclop
func (c *RPCClient) Call(ctx context.Context, exchange, routingKey string, body []byte, opts ...RPCCallOptionFunc) ([]byte, error) {
	o := defaultRPCCallOptions()
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	const replyQueueName = "amq.rabbitmq.reply-to"

	channel, err := c.con.Connection().Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}
	defer func() {
		if err := channel.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close channel", "error", err.Error())
		}
	}()

	// Start a goroutine to consume the response
	msgs, err := channel.Consume(
		replyQueueName, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return nil, fmt.Errorf("consume response: %w", err)
	}

	headers := amqp.Table{}
	if o.bearerToken != "" {
		headers["performer"] = amqp.Table{
			"token": o.bearerToken,
		}
	}

	corrID := uuid.NewString()
	pubMsg := amqp.Publishing{
		Headers:       headers,
		ContentType:   o.contentType,
		CorrelationId: corrID,
		ReplyTo:       replyQueueName,
		Body:          body,
	}

	if err := channel.Publish(exchange, routingKey, o.mandatory, o.immediate, pubMsg); err != nil {
		return nil, err
	}

	for {
		select {
		case response := <-msgs:
			if response.CorrelationId != corrID {
				continue
			}

			if response.Body == nil {
				err := errors.New("empty response received")
				return nil, err
			}

			return response.Body, nil
		case <-time.After(c.options.timeout):
			err := errors.New("rpc call timeout")
			return nil, err
		}
	}
}
