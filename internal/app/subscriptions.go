package app

import (
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	"github.com/google/uuid"
)

// Consumer is the interface that wraps the basic AMQP subscriber/consumer methods.
type Consumer interface {
	Subscribe(sub *subscriptions.Subscription) error
	Unsubscribe(id uuid.UUID) error
	UnsubscribeFromQueue(queue string) error
	GetAllSubscriptions() []*subscriptions.Subscription
	GetQueueSubscriptions(queue string) []*subscriptions.Subscription
	UnsubscribeAll() error
}

// SubscriptionsService is the application level service to manage AMQP subscriptions.
type SubscriptionsService struct {
	consumer Consumer
}

// NewSubscriptionsService creates a new SubscriptionsService instance.
func NewSubscriptionsService(consumer Consumer) *SubscriptionsService {
	return &SubscriptionsService{
		consumer: consumer,
	}
}

// Subscribe subscribes to a queue.
func (s *SubscriptionsService) Subscribe(queue string, idempotent bool) (*subscriptions.Subscription, error) {
	if idempotent {
		subs := s.consumer.GetQueueSubscriptions(queue)
		if len(subs) > 0 {
			// return the first subscription found
			return subs[0], nil
		}
	}

	sub := subscriptions.NewSubscription(queue)
	if err := s.consumer.Subscribe(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// UnsubscribeByID unsubscribes from a queue by ID.
func (s *SubscriptionsService) UnsubscribeByID(id uuid.UUID) error {
	return s.consumer.Unsubscribe(id)
}

// UnsubscribeByQueue unsubscribes from a queue.
func (s *SubscriptionsService) UnsubscribeByQueue(queue string) error {
	return s.consumer.UnsubscribeFromQueue(queue)
}

// GetAllSubscriptions returns all subscriptions.
func (s *SubscriptionsService) GetAllSubscriptions() []*subscriptions.Subscription {
	return s.consumer.GetAllSubscriptions()
}

// UnsubscribeAll resets all subscriptions.
func (s *SubscriptionsService) UnsubscribeAll() error {
	return s.consumer.UnsubscribeAll()
}
