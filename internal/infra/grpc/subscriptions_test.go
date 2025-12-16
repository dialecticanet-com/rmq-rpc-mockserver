package grpc

import (
	"context"
	"testing"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddSubscription tests the AddSubscription handler
func TestAddSubscription(t *testing.T) {
	// Create a mock subscriptions service
	mockSvc := &TestSubscriptionsService{}

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		subscriptionsService: mockSvc,
	}

	// Create a test request
	req := &grpcApi.AddSubscriptionRequest{
		Queue:      "test-queue",
		Idempotent: true,
	}

	// Call the handler
	resp, err := server.AddSubscription(context.Background(), req)

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Subscription.Id)
	assert.Equal(t, req.Queue, resp.Subscription.Queue)

	// Verify the subscription was created in the service
	assert.Len(t, mockSvc.subscriptions, 1)
	assert.Equal(t, req.Queue, mockSvc.subscriptions[0].Queue())
}

// TestDeleteSubscription tests the DeleteSubscription handler
func TestDeleteSubscription(t *testing.T) {
	// Create a mock subscriptions service
	mockSvc := &TestSubscriptionsService{}

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		subscriptionsService: mockSvc,
	}

	// Create a test subscription
	sub := subscriptions.NewSubscription("test-queue")
	mockSvc.subscriptions = append(mockSvc.subscriptions, sub)

	// Create a test request
	req := &grpcApi.DeleteSubscriptionRequest{
		SubscriptionId: sub.ID().String(),
	}

	// Call the handler
	resp, err := server.DeleteSubscription(context.Background(), req)

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestUnsubscribeFromQueue tests the UnsubscribeFromQueue handler
func TestUnsubscribeFromQueue(t *testing.T) {
	// Create a mock subscriptions service
	mockSvc := &TestSubscriptionsService{}

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		subscriptionsService: mockSvc,
	}

	// Create a test subscription
	sub := subscriptions.NewSubscription("test-queue")
	mockSvc.subscriptions = append(mockSvc.subscriptions, sub)

	// Create a test request
	req := &grpcApi.UnsubscribeFromQueueRequest{
		Queue: "test-queue",
	}

	// Call the handler
	resp, err := server.UnsubscribeFromQueue(context.Background(), req)

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestGetAllSubscriptions tests the GetAllSubscriptions handler
func TestGetAllSubscriptions(t *testing.T) {
	// Create a mock subscriptions service
	mockSvc := &TestSubscriptionsService{}

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		subscriptionsService: mockSvc,
	}

	// Create some test subscriptions
	sub1 := subscriptions.NewSubscription("test-queue-1")
	mockSvc.subscriptions = append(mockSvc.subscriptions, sub1)
	sub2 := subscriptions.NewSubscription("test-queue-2")
	mockSvc.subscriptions = append(mockSvc.subscriptions, sub2)

	// Call the handler
	resp, err := server.GetAllSubscriptions(context.Background(), &grpcApi.GetAllSubscriptionsRequest{})

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Subscriptions, 2)
	assert.Equal(t, sub1.ID().String(), resp.Subscriptions[0].Id)
	assert.Equal(t, sub1.Queue(), resp.Subscriptions[0].Queue)
	assert.Equal(t, sub2.ID().String(), resp.Subscriptions[1].Id)
	assert.Equal(t, sub2.Queue(), resp.Subscriptions[1].Queue)
}

// TestResetSubscriptions tests the ResetSubscriptions handler
func TestResetSubscriptions(t *testing.T) {
	// Create a mock subscriptions service
	mockSvc := &TestSubscriptionsService{}

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		subscriptionsService: mockSvc,
	}

	// Create some test subscriptions
	sub1 := subscriptions.NewSubscription("test-queue-1")
	mockSvc.subscriptions = append(mockSvc.subscriptions, sub1)
	sub2 := subscriptions.NewSubscription("test-queue-2")
	mockSvc.subscriptions = append(mockSvc.subscriptions, sub2)

	// Call the handler
	resp, err := server.ResetSubscriptions(context.Background(), &grpcApi.ResetSubscriptionsRequest{})

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestResetAll tests the ResetAll handler
func TestResetAll(t *testing.T) {
	// Create mock services
	mockExpSvc := &MockExpectationsService{}
	mockSubSvc := &TestSubscriptionsService{}

	// Create the server with the mock services
	server := &AmqpMockServerServiceServer{
		expectationsService:  mockExpSvc,
		subscriptionsService: mockSubSvc,
	}

	// Create some test data
	exp, err := createTestExpectation("exchange", "rk")
	require.NoError(t, err)
	err = mockExpSvc.Create(exp)
	require.NoError(t, err)

	sub := subscriptions.NewSubscription("test-queue")
	mockSubSvc.subscriptions = append(mockSubSvc.subscriptions, sub)

	// Call the handler
	resp, err := server.ResetAll(context.Background(), &grpcApi.ResetAllRequest{})

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify the reset was called on both services
	assert.True(t, mockExpSvc.resetCalled)
}
