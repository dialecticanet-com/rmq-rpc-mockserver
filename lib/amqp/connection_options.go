package amqp

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type connectionOptions struct {
	reconnectionDelay time.Duration
	dialConfig        *amqp.Config
}

func defaultConnectionOptions() *connectionOptions {
	return &connectionOptions{
		reconnectionDelay: 5 * time.Second,
		dialConfig:        nil,
	}
}

// ConnectionOption is a function that sets options for the connection.
type ConnectionOption func(*connectionOptions) error

// ConnectionWithReconnectionDelay sets the reconnection delay for the connection.
func ConnectionWithReconnectionDelay(delay time.Duration) ConnectionOption {
	return func(o *connectionOptions) error {
		o.reconnectionDelay = delay
		return nil
	}
}

// ConnectionWithDialConfig sets the dial config for the connection.
func ConnectionWithDialConfig(config *amqp.Config) ConnectionOption {
	return func(o *connectionOptions) error {
		o.dialConfig = config
		return nil
	}
}
