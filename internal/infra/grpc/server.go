package grpc

import (
	"context"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/app"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/config"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	"github.com/google/uuid"
)

// ExpectationsService is the interface that wraps the basic expectations service methods.
type ExpectationsService interface {
	Create(exp *expectations.Expectation) error
	Reset()
	Match(cnd *expectations.Candidate) *expectations.Response
	GetExpectations(req app.GetExpectationsRequest) []*expectations.Expectation
	GetExpectation(id uuid.UUID) *expectations.Expectation
	GetAssertions(req app.GetAssertionsRequest) []*expectations.Assertion
}

// SubscriptionsService is the interface that wraps the basic subscriptions service methods.
type SubscriptionsService interface {
	Subscribe(queue string, idempotent bool) (*subscriptions.Subscription, error)
	UnsubscribeByID(id uuid.UUID) error
	UnsubscribeByQueue(queue string) error
	GetAllSubscriptions() []*subscriptions.Subscription
	UnsubscribeAll() error
}

// AmqpMockServerServiceServer is the gRPC server implementation for the AmqpMockServerService service.
type AmqpMockServerServiceServer struct {
	grpcApi.UnimplementedAmqpMockServerServiceServer
	expectationsService  ExpectationsService
	subscriptionsService SubscriptionsService
	serviceInfo          *config.ServiceInfo
}

// NewAmqpMockServerServiceServer creates a new AmqpMockServerServiceServer instance.
func NewAmqpMockServerServiceServer(expSvc ExpectationsService, subSvc SubscriptionsService, si *config.ServiceInfo) *AmqpMockServerServiceServer {
	return &AmqpMockServerServiceServer{
		expectationsService:  expSvc,
		subscriptionsService: subSvc,
		serviceInfo:          si,
	}
}

func (s *AmqpMockServerServiceServer) GetVersion(_ context.Context, _ *grpcApi.GetVersionRequest) (*grpcApi.GetVersionResponse, error) {
	return &grpcApi.GetVersionResponse{
		Version:    s.serviceInfo.Version,
		CommitHash: s.serviceInfo.CommitHash,
		BuildDate:  s.serviceInfo.BuildDate,
	}, nil
}
