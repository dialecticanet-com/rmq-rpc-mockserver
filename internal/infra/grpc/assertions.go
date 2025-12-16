package grpc

import (
	"context"
	"fmt"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/app"
	"github.com/google/uuid"
)

func (s *AmqpMockServerServiceServer) GetAssertions(_ context.Context, req *grpcApi.GetAssertionsRequest) (*grpcApi.GetAssertionsResponse, error) {
	appReq := app.GetAssertionsRequest{
		Status:  req.Status,
		Include: req.Include,
	}

	if req.ExpectationId != nil {
		expUID, err := uuid.Parse(*req.ExpectationId)
		if err != nil {
			return nil, fmt.Errorf("invalid expectation id: %w", err)
		}
		appReq.ExpectationID = &expUID
	}

	assertions := s.expectationsService.GetAssertions(appReq)

	assertionsDTO := make([]*grpcApi.Assertion, 0, len(assertions))
	for _, assertion := range assertions {
		assertionsDTO = append(assertionsDTO, newProtoAssertion(assertion, req.Include))
	}

	return &grpcApi.GetAssertionsResponse{
		Assertions: assertionsDTO,
	}, nil
}
