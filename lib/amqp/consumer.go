package amqp

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageHandler is a function that processes an AMQP message.
type MessageHandler func(context.Context, *amqp.Delivery) *HandlerError

// HandlerError is a generic error type for message handlers.
// it has some additional information about the error, like if the message should be requeued.
type HandlerError struct {
	err     error
	requeue bool
}

// Error returns the error message.
func (h HandlerError) Error() string {
	return h.err.Error()
}

// NewHandlerError creates a new HandlerError instance.
func NewHandlerError(err error, requeue bool) *HandlerError {
	return &HandlerError{
		err:     err,
		requeue: requeue,
	}
}

// Consumer is a message consumer for AMQP messages.
type Consumer struct {
	conn     *Connection
	ch       *amqp.Channel
	queue    string
	handlers map[string]MessageHandler
}

// NewConsumer creates a new AMQP consumer for a specific queue.
func NewConsumer(con *Connection, queue string) (*Consumer, error) {
	ch, err := con.Connection().Channel()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:     con,
		ch:       ch,
		queue:    queue,
		handlers: make(map[string]MessageHandler),
	}, nil
}

// Run starts the consumer and processes incoming messages.
func (c *Consumer) Run(ctx context.Context) error {
	slog.Info("starting AMQP consumer", "queue", c.queue)

	deliveries, err := c.ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("AMQP consumer stopped", "queue", c.queue)
			return nil
		case d, ok := <-deliveries:
			// check if the channel is closed
			if !ok {
				return nil
			}

			handler, ok := c.handlers[c.buildRoutingPath(d.Exchange, d.RoutingKey)]
			if !ok {
				slog.Error("no AMQP message handler found", "exchange", d.Exchange, "routing_key", d.RoutingKey)
				continue
			}

			// nolint: contextcheck
			// global context should not be used here, because it will be cancelled when the service is about to stop,
			// and we don't want to cancel the context for the message handler immediately, it should finish its work.
			c.handle(&d, handler)
		}
	}
}

// Subscribe registers a new message handler for the specified exchange and routing key.
func (c *Consumer) Subscribe(exchange, routingKey string, handler MessageHandler) {
	c.handlers[c.buildRoutingPath(exchange, routingKey)] = handler
}

func (c *Consumer) handle(d *amqp.Delivery, handler MessageHandler) {
	log := slog.With("exchange", d.Exchange, "routing_key", d.RoutingKey, "message_id", d.MessageId, "queue", c.queue)

	if err := handler(context.Background(), d); err != nil {
		log.Error("failed to process AMQP message", "error", err.Error(), "requeue", err.requeue)
		if err := d.Nack(false, err.requeue); err != nil {
			log.Error("failed to NACK the message", "error", err.Error())
		}

		return
	}

	if err := d.Ack(false); err != nil {
		log.Error("failed to ACK the message", "error", err.Error())
	}
}

func (c *Consumer) buildRoutingPath(exchange, routingKey string) string {
	return exchange + "/" + routingKey
}
