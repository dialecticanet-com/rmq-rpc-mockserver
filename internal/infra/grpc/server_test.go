package grpc

import (
	"context"
	"testing"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/config"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAmqpMockServerServiceServer tests the NewAmqpMockServerServiceServer function
func TestNewAmqpMockServerServiceServer(t *testing.T) {
	// Create mock dependencies
	expSvc := &TestExpectationsService{}
	subSvc := &TestSubscriptionsService{}
	si := &config.ServiceInfo{
		Service:    "test-service",
		Version:    "1.0.0",
		CommitHash: "abc123",
		BuildDate:  "2023-01-01",
	}

	// Create the server
	server := NewAmqpMockServerServiceServer(expSvc, subSvc, si)

	// Verify the server was created correctly
	assert.NotNil(t, server)
	assert.Equal(t, expSvc, server.expectationsService)
	assert.Equal(t, subSvc, server.subscriptionsService)
	assert.Equal(t, si, server.serviceInfo)
}

// TestGetVersion tests the GetVersion method
func TestGetVersion(t *testing.T) {
	// Create mock dependencies
	expSvc := &TestExpectationsService{}
	subSvc := &TestSubscriptionsService{}
	si := &config.ServiceInfo{
		Service:    "test-service",
		Version:    "1.0.0",
		CommitHash: "abc123",
		BuildDate:  "2023-01-01",
	}

	// Create the server
	server := NewAmqpMockServerServiceServer(expSvc, subSvc, si)

	// Call the GetVersion method
	resp, err := server.GetVersion(context.Background(), &grpcApi.GetVersionRequest{})

	// Verify the response
	require.NoError(t, err)
	assert.Equal(t, si.Version, resp.Version)
	assert.Equal(t, si.CommitHash, resp.CommitHash)
	assert.Equal(t, si.BuildDate, resp.BuildDate)
}

// TestSubscriptionsService is a simple implementation of the SubscriptionsService interface for testing
type TestSubscriptionsService struct {
	subscriptions []*subscriptions.Subscription
}

func (s *TestSubscriptionsService) Subscribe(queue string, _ bool) (*subscriptions.Subscription, error) {
	sub := subscriptions.NewSubscription(queue)
	s.subscriptions = append(s.subscriptions, sub)
	return sub, nil
}

func (s *TestSubscriptionsService) UnsubscribeByID(_ uuid.UUID) error {
	return nil
}

func (s *TestSubscriptionsService) UnsubscribeByQueue(_ string) error {
	return nil
}

func (s *TestSubscriptionsService) GetAllSubscriptions() []*subscriptions.Subscription {
	return s.subscriptions
}

func (s *TestSubscriptionsService) UnsubscribeAll() error {
	s.subscriptions = nil
	return nil
}
