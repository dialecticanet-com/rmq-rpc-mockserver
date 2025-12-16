package amqp

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpListener struct {
	subscription *subscriptions.Subscription
	channel      *amqp.Channel
	deliveries   <-chan amqp.Delivery
	matcher      Matcher
}

func startListener(wg *sync.WaitGroup, con *amqp.Connection, sub *subscriptions.Subscription, matcher Matcher) (*amqpListener, error) {
	ch, err := con.Channel()
	if err != nil {
		return nil, err
	}

	deliveries, err := ch.Consume(sub.Queue(), "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	lst := &amqpListener{
		subscription: sub,
		channel:      ch,
		deliveries:   deliveries,
		matcher:      matcher,
	}

	wg.Add(1)
	go lst.listen(wg)

	return lst, nil
}

func (c *amqpListener) stop() error {
	err := c.channel.Close()

	// ignore if the connection is already closed
	if errors.Is(err, amqp.ErrClosed) {
		return nil
	}

	return c.channel.Close()
}

func (c *amqpListener) listen(wg *sync.WaitGroup) {
	slog.Info("starting AMQP listener", "queue", c.subscription.Queue())
	defer wg.Done()

	for {
		delivery, consuming := <-c.deliveries
		if !consuming {
			slog.Info("AMQP listener stopped", "queue", c.subscription.Queue())
			return
		}
		c.handleMessage(delivery)
	}
}

func (c *amqpListener) handleMessage(delivery amqp.Delivery) {
	candidate, err := expectations.NewCandidate(delivery.Exchange, delivery.RoutingKey, delivery.Body)
	if err != nil {
		slog.Error("failed to create candidate", "error", err)
		return
	}

	response := c.matcher.Match(candidate)
	if response == nil {
		notFoundResponse := `{"errors":"no match found"}`
		response, _ = expectations.NewResponse([]byte(notFoundResponse))
	}

	msg := amqp.Publishing{ContentType: "application/json", CorrelationId: delivery.CorrelationId, Body: response.Body}
	if err := c.channel.Publish("", delivery.ReplyTo, false, false, msg); err != nil {
		slog.Error("failed to publish response", "error", err)
	}

	if err := delivery.Ack(false); err != nil {
		slog.Error("failed to acknowledge message", "error", err)
	}
}
