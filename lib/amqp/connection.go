package amqp

import (
	"context"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connection represents a connection to RabbitMQ.
type Connection struct {
	url        string
	connection *amqp.Connection
	options    *connectionOptions
	m          sync.RWMutex
}

// EstablishConnection establishes a connection to RabbitMQ.
func EstablishConnection(ctx context.Context, url string, opts ...ConnectionOption) (*Connection, error) {
	o := defaultConnectionOptions()
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	wcon := &Connection{
		url:     url,
		options: o,
	}

	if err := wcon.connect(); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "successfully connected to RabbitMQ")

	return wcon, nil
}

// Run will block until the context is canceled or the connection is closed.
func (c *Connection) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		err := c.connection.Close()
		if err != nil {
			return err
		}
		slog.Info("rabbitMQ connection closed gracefully")
		return nil
	case err, ok := <-c.Connection().NotifyClose(make(chan *amqp.Error)):
		if !ok {
			slog.Info("rabbitMQ connection closed gracefully")
			return nil
		}
		return err
	}
}

func (c *Connection) connect() error {
	var conn *amqp.Connection
	var err error

	if c.options.dialConfig == nil {
		conn, err = amqp.Dial(c.url)
	} else {
		conn, err = amqp.DialConfig(c.url, *c.options.dialConfig)
	}

	if err != nil {
		return err
	}

	c.m.Lock()
	c.connection = conn
	c.m.Unlock()

	return nil
}

// Connection returns the underlying amqp.Connection.
func (c *Connection) Connection() *amqp.Connection {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.connection
}
