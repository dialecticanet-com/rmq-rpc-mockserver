package grpc

import (
	"context"
	"fmt"
	"time"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/app"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/comparators"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/google/uuid"
)

// CreateExpectation creates a new expectation.
func (s *AmqpMockServerServiceServer) CreateExpectation(_ context.Context, req *grpcApi.CreateExpectationRequest) (*grpcApi.CreateExpectationResponse, error) {
	request, err := newExpectationsRequest(req.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to create expectation request: %w", err)
	}

	response, err := newExpectationsResponse(req.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to create expectation response: %w", err)
	}

	exp, err := expectations.NewExpectation(request, response, newExpectationOptions(req)...)
	if err != nil {
		return nil, fmt.Errorf("failed to create expectation: %w", err)
	}

	err = s.expectationsService.Create(exp)
	if err != nil {
		return nil, fmt.Errorf("failed to create expectation: %w", err)
	}

	return &grpcApi.CreateExpectationResponse{
		ExpectationId: exp.ID.String(),
	}, nil
}

// ResetExpectations removes all expectations from the service.
func (s *AmqpMockServerServiceServer) ResetExpectations(_ context.Context, _ *grpcApi.ResetExpectationsRequest) (*grpcApi.ResetExpectationsResponse, error) {
	s.expectationsService.Reset()
	return &grpcApi.ResetExpectationsResponse{}, nil
}

// GetExpectations returns list of expectations.
func (s *AmqpMockServerServiceServer) GetExpectations(_ context.Context, req *grpcApi.GetExpectationsRequest) (*grpcApi.GetExpectationsResponse, error) {
	appReq := app.GetExpectationsRequest{
		Status: req.Status,
	}

	exps := s.expectationsService.GetExpectations(appReq)

	expDTOs := make([]*grpcApi.Expectation, 0, len(exps))
	for _, exp := range exps {
		expDTOs = append(expDTOs, newProtoExpectation(exp))
	}

	return &grpcApi.GetExpectationsResponse{
		Expectations: expDTOs,
	}, nil
}

// GetExpectation returns a single expectation.
func (s *AmqpMockServerServiceServer) GetExpectation(_ context.Context, req *grpcApi.GetExpectationRequest) (*grpcApi.GetExpectationResponse, error) {
	expUID, err := uuid.Parse(req.ExpectationId)
	if err != nil {
		return nil, fmt.Errorf("invalid expectation id: %w", err)
	}

	exp := s.expectationsService.GetExpectation(expUID)
	if exp == nil {
		return nil, fmt.Errorf("expectation not found")
	}

	return &grpcApi.GetExpectationResponse{
		Expectation: newProtoExpectation(exp),
	}, nil
}

func newExpectationsRequest(req *grpcApi.Request) (*expectations.Request, error) {
	comparator, err := newComparator(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create body comparator: %w", err)
	}

	request, err := expectations.NewRequest(req.Exchange, req.RoutingKey, comparator)
	if err != nil {
		return nil, fmt.Errorf("failed to create expectation request: %w", err)
	}

	return request, nil
}

func newExpectationsResponse(res *grpcApi.Response) (*expectations.Response, error) {
	resBodyJSON, err := res.Body.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("unable to read expectation response JSON body: %w", err)
	}

	response, err := expectations.NewResponse(resBodyJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create expectation response: %w", err)
	}

	return response, nil
}

func newExpectationOptions(req *grpcApi.CreateExpectationRequest) []expectations.ExpectationOption {
	expOpts := make([]expectations.ExpectationOption, 0)

	if req.GetTimes() != nil {
		if req.GetTimes().GetRemainingTimes() > 0 {
			expOpts = append(expOpts, expectations.WithLimitedTimes(req.GetTimes().GetRemainingTimes()))
		} else {
			expOpts = append(expOpts, expectations.WithUnlimitedTimes())
		}
	}

	if req.TimeToLiveSeconds != nil {
		expOpts = append(expOpts, expectations.WithTimeToLive(time.Duration(float64(*req.TimeToLiveSeconds)*float64(time.Second))))
	}

	return expOpts
}

// nolint: ireturn
func newComparator(req *grpcApi.Request) (expectations.BodyComparator, error) {
	switch body := req.GetBody().(type) {
	case *grpcApi.Request_JsonBody:
		matchType := comparators.MatchTypeExact
		switch body.JsonBody.GetMatchType() {
		case grpcApi.JSONBodyAssertion_MATCH_TYPE_EXACT:
			matchType = comparators.MatchTypeExact
		case grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL:
			matchType = comparators.MatchTypePartial
		}

		rawBody, err := body.JsonBody.Body.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("unable to read expectation request JSON body: %w", err)
		}

		return comparators.NewJSONBody(rawBody, matchType)
	case *grpcApi.Request_RegexBody:
		return comparators.NewRegex(body.RegexBody.GetRegex())
	default:
		return nil, fmt.Errorf("unsupported request body type: %T", body)
	}
}
