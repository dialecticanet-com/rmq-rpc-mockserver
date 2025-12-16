package subscriptions

import "github.com/google/uuid"

type Subscription struct {
	id    uuid.UUID
	queue string
}

func NewSubscription(queue string) *Subscription {
	return &Subscription{
		id:    uuid.New(),
		queue: queue,
	}
}

func (s *Subscription) ID() uuid.UUID {
	return s.id
}

func (s *Subscription) Queue() string {
	return s.queue
}
