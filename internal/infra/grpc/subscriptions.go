package grpc

import (
	"context"
	"fmt"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/google/uuid"
)

// AddSubscription adds a new subscription to the server.
func (s *AmqpMockServerServiceServer) AddSubscription(_ context.Context, request *grpcApi.AddSubscriptionRequest) (*grpcApi.AddSubscriptionResponse, error) {
	sub, err := s.subscriptionsService.Subscribe(request.Queue, request.Idempotent)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	return &grpcApi.AddSubscriptionResponse{
		Subscription: newSubscription(sub),
	}, nil
}

// DeleteSubscription deletes a subscription from the server.
func (s *AmqpMockServerServiceServer) DeleteSubscription(_ context.Context, request *grpcApi.DeleteSubscriptionRequest) (*grpcApi.DeleteSubscriptionResponse, error) {

	if err := s.subscriptionsService.UnsubscribeByID(uuid.MustParse(request.SubscriptionId)); err != nil {
		return nil, fmt.Errorf("failed to unsubscribe: %w", err)
	}

	return &grpcApi.DeleteSubscriptionResponse{}, nil
}

// UnsubscribeFromQueue unsubscribes from a queue.
func (s *AmqpMockServerServiceServer) UnsubscribeFromQueue(_ context.Context, request *grpcApi.UnsubscribeFromQueueRequest) (*grpcApi.UnsubscribeFromQueueResponse, error) {
	if err := s.subscriptionsService.UnsubscribeByQueue(request.Queue); err != nil {
		return nil, fmt.Errorf("failed to unsubscribe from queue: %w", err)
	}

	return &grpcApi.UnsubscribeFromQueueResponse{}, nil
}

// GetAllSubscriptions returns all subscriptions.
func (s *AmqpMockServerServiceServer) GetAllSubscriptions(_ context.Context, _ *grpcApi.GetAllSubscriptionsRequest) (*grpcApi.GetAllSubscriptionsResponse, error) {
	subs := s.subscriptionsService.GetAllSubscriptions()
	subsDTO := make([]*grpcApi.Subscription, 0, len(subs))
	for _, sub := range subs {
		subsDTO = append(subsDTO, newSubscription(sub))
	}

	return &grpcApi.GetAllSubscriptionsResponse{
		Subscriptions: subsDTO,
	}, nil
}

func (s *AmqpMockServerServiceServer) ResetSubscriptions(_ context.Context, _ *grpcApi.ResetSubscriptionsRequest) (*grpcApi.ResetSubscriptionsResponse, error) {
	err := s.subscriptionsService.UnsubscribeAll()

	return &grpcApi.ResetSubscriptionsResponse{}, err
}

func (s *AmqpMockServerServiceServer) ResetAll(_ context.Context, _ *grpcApi.ResetAllRequest) (*grpcApi.ResetAllResponse, error) {
	s.expectationsService.Reset()
	err := s.subscriptionsService.UnsubscribeAll()

	return &grpcApi.ResetAllResponse{}, err
}
