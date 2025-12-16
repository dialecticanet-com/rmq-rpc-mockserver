package amqp

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestConnectionWithReconnectionDelay(t *testing.T) {
	delay := time.Minute
	o := &connectionOptions{}
	err := ConnectionWithReconnectionDelay(delay)(o)
	assert.NoError(t, err)
	assert.Equal(t, delay, o.reconnectionDelay)
}

func TestConnectionWithDialConfig(t *testing.T) {
	config := &amqp.Config{}
	o := &connectionOptions{}
	err := ConnectionWithDialConfig(config)(o)
	assert.NoError(t, err)
	assert.Equal(t, config, o.dialConfig)
}
