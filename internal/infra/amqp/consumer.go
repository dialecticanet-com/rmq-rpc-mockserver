package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	gocoreamqp "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/amqp"
	"github.com/google/uuid"
)

// Matcher is an interface for matching expectations against candidates.
type Matcher interface {
	Match(candidate *expectations.Candidate) *expectations.Response
}

// Consumer is a message consumer for AMQP messages.
type Consumer struct {
	m            sync.RWMutex
	conn         *gocoreamqp.Connection
	matcher      Matcher
	waitingGroup *sync.WaitGroup
	listeners    map[uuid.UUID]*amqpListener
}

// NewConsumer creates a new AMQP consumer for a list of queues.
func NewConsumer(con *gocoreamqp.Connection, matcher Matcher) (*Consumer, error) {
	return &Consumer{
		conn:         con,
		matcher:      matcher,
		listeners:    make(map[uuid.UUID]*amqpListener),
		waitingGroup: &sync.WaitGroup{},
	}, nil
}

// Run starts the consumer and processes incoming messages.
func (c *Consumer) Run(ctx context.Context) error {
	<-ctx.Done()
	if err := c.stop(); err != nil {
		return fmt.Errorf("stopping amqp consumer: %w", err)
	}

	// wait for all subscribers to shut down
	c.waitingGroup.Wait()
	slog.Info("All AMQP consumers stopped")
	return nil
}

// Subscribe subscribes to a queue.
func (c *Consumer) Subscribe(sub *subscriptions.Subscription) error {
	lst, err := startListener(c.waitingGroup, c.conn.Connection(), sub, c.matcher)
	if err != nil {
		return fmt.Errorf("starting amqp listener for queue %s: %w", sub.Queue(), err)
	}

	c.m.Lock()
	c.listeners[sub.ID()] = lst
	c.m.Unlock()

	return nil
}

// Unsubscribe removes a specific subscription by its ID.
func (c *Consumer) Unsubscribe(id uuid.UUID) error {
	c.m.Lock()
	defer c.m.Unlock()

	if lst, ok := c.listeners[id]; ok {
		if err := lst.stop(); err != nil {
			return fmt.Errorf("stopping amqp listener for queue %s: %w", lst.subscription.Queue(), err)
		}
		delete(c.listeners, id)
	}

	return nil
}

// UnsubscribeFromQueue removes all subscriptions for a specific queue.
func (c *Consumer) UnsubscribeFromQueue(queue string) error {
	c.m.Lock()
	defer c.m.Unlock()

	for id, lst := range c.listeners {
		if lst.subscription.Queue() == queue {
			if err := lst.stop(); err != nil {
				return fmt.Errorf("stopping amqp listener for queue %s: %w", lst.subscription.Queue(), err)
			}
			delete(c.listeners, id)
		}
	}

	return nil
}

func (c *Consumer) UnsubscribeAll() error {
	c.m.Lock()
	defer c.m.Unlock()

	for id, lst := range c.listeners {
		if err := lst.stop(); err != nil {
			return fmt.Errorf("stopping amqp listener for queue %s: %w", lst.subscription.Queue(), err)
		}
		delete(c.listeners, id)
	}

	return nil
}

// GetQueueSubscriptions returns all subscriptions for a specific queue.
func (c *Consumer) GetQueueSubscriptions(queue string) []*subscriptions.Subscription {
	c.m.RLock()
	defer c.m.RUnlock()

	var subs []*subscriptions.Subscription
	for _, lst := range c.listeners {
		if lst.subscription.Queue() == queue {
			subs = append(subs, lst.subscription)
		}
	}

	return subs
}

// GetAllSubscriptions returns all active subscriptions.
func (c *Consumer) GetAllSubscriptions() []*subscriptions.Subscription {
	c.m.RLock()
	defer c.m.RUnlock()

	subs := make([]*subscriptions.Subscription, 0, len(c.listeners))
	for _, lst := range c.listeners {
		subs = append(subs, lst.subscription)
	}

	return subs
}

func (c *Consumer) stop() error {
	c.m.Lock()
	defer c.m.Unlock()

	for _, lst := range c.listeners {
		if err := lst.stop(); err != nil {
			return fmt.Errorf("stopping amqp listener: %w", err)
		}
	}

	return nil
}
