package grpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/app"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/comparators"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestNewExpectationsRequest tests the newExpectationsRequest function
func TestNewExpectationsRequest(t *testing.T) {
	// Create a proto request with JSON body
	protoReq := &grpcApi.Request{
		Exchange:   "test-exchange",
		RoutingKey: "test-routing-key",
		Body: &grpcApi.Request_JsonBody{
			JsonBody: &grpcApi.JSONBodyAssertion{
				MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
				Body:      createJSONStruct(t, `{"foo":"bar"}`),
			},
		},
	}

	// Convert to domain request
	domainReq, err := newExpectationsRequest(protoReq)

	// Verify the conversion
	require.NoError(t, err)
	assert.Equal(t, "test-exchange", domainReq.Exchange)
	assert.Equal(t, "test-routing-key", domainReq.RoutingKey)

	// Verify the body comparator
	jsonBody, ok := domainReq.BodyComparator.(*comparators.JSONBody)
	require.True(t, ok, "Expected JSONBody comparator")
	assert.Equal(t, comparators.MatchTypePartial, jsonBody.MatchType)

	// Create a proto request with Regex body
	protoReq = &grpcApi.Request{
		Exchange:   "test-exchange",
		RoutingKey: "test-routing-key",
		Body: &grpcApi.Request_RegexBody{
			RegexBody: &grpcApi.RegexBodyAssertion{
				Regex: "foo.*bar",
			},
		},
	}

	// Convert to domain request
	domainReq, err = newExpectationsRequest(protoReq)

	// Verify the conversion
	require.NoError(t, err)
	assert.Equal(t, "test-exchange", domainReq.Exchange)
	assert.Equal(t, "test-routing-key", domainReq.RoutingKey)

	// Verify the body comparator
	regexBody, ok := domainReq.BodyComparator.(*comparators.Regex)
	require.True(t, ok, "Expected Regex comparator")
	assert.Equal(t, "foo.*bar", regexBody.Regex.String())
}

// TestNewExpectationsResponse tests the newExpectationsResponse function
func TestNewExpectationsResponse(t *testing.T) {
	// Create a proto response
	protoRes := &grpcApi.Response{
		Body: createJSONValue(t, `{"result":"success"}`),
	}

	// Convert to domain response
	domainRes, err := newExpectationsResponse(protoRes)

	// Verify the conversion
	require.NoError(t, err)
	assert.JSONEq(t, `{"result":"success"}`, string(domainRes.Body))
}

// TestNewExpectationOptions tests the newExpectationOptions function
func TestNewExpectationOptions(t *testing.T) {
	t.Run("with limited times", func(t *testing.T) {
		// Create a proto request with limited times
		times := uint32(5)
		protoReq := &grpcApi.CreateExpectationRequest{
			Times: &grpcApi.Times{
				Times: &grpcApi.Times_RemainingTimes{
					RemainingTimes: times,
				},
			},
		}

		// Get the options
		options := newExpectationOptions(protoReq)

		// Create an expectation with these options
		exp, err := createTestExpectation("exchange", "rk", options...)
		require.NoError(t, err)

		// Verify the times
		assert.NotNil(t, exp.Times)
		assert.False(t, exp.Times.Unlimited)
		assert.Equal(t, times, exp.Times.RemainingTimes)
	})

	t.Run("with unlimited times", func(t *testing.T) {
		// Create a proto request with unlimited times
		protoReq := &grpcApi.CreateExpectationRequest{
			Times: &grpcApi.Times{
				Times: &grpcApi.Times_Unlimited{
					Unlimited: true,
				},
			},
		}

		// Get the options
		options := newExpectationOptions(protoReq)

		// Create an expectation with these options
		exp, err := createTestExpectation("exchange", "rk", options...)
		require.NoError(t, err)

		// Verify the times
		assert.NotNil(t, exp.Times)
		assert.True(t, exp.Times.Unlimited)
	})

	t.Run("with TTL", func(t *testing.T) {
		// Create a proto request with TTL
		ttlSeconds := float32(60)
		protoReq := &grpcApi.CreateExpectationRequest{
			TimeToLiveSeconds: &ttlSeconds,
		}

		// Get the options
		options := newExpectationOptions(protoReq)

		// Create an expectation with these options
		exp, err := createTestExpectation("exchange", "rk", options...)
		require.NoError(t, err)

		// Verify the TTL
		assert.NotNil(t, exp.TimeToLive)
		assert.Equal(t, time.Duration(float64(ttlSeconds)*float64(time.Second)), exp.TimeToLive.TTL)
	})
}

// TestNewComparator tests the newComparator function
func TestNewComparator(t *testing.T) {
	t.Run("with JSON body", func(t *testing.T) {
		// Create a proto request with JSON body
		protoReq := &grpcApi.Request{
			Body: &grpcApi.Request_JsonBody{
				JsonBody: &grpcApi.JSONBodyAssertion{
					MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
					Body:      createJSONStruct(t, `{"foo":"bar"}`),
				},
			},
		}

		// Create a comparator
		comparator, err := newComparator(protoReq)

		// Verify the comparator
		require.NoError(t, err)
		jsonBody, ok := comparator.(*comparators.JSONBody)
		require.True(t, ok, "Expected JSONBody comparator")
		assert.Equal(t, comparators.MatchTypePartial, jsonBody.MatchType)
	})

	t.Run("with Regex body", func(t *testing.T) {
		// Create a proto request with Regex body
		protoReq := &grpcApi.Request{
			Body: &grpcApi.Request_RegexBody{
				RegexBody: &grpcApi.RegexBodyAssertion{
					Regex: "foo.*bar",
				},
			},
		}

		// Create a comparator
		comparator, err := newComparator(protoReq)

		// Verify the comparator
		require.NoError(t, err)
		regexBody, ok := comparator.(*comparators.Regex)
		require.True(t, ok, "Expected Regex comparator")
		assert.Equal(t, "foo.*bar", regexBody.Regex.String())
	})
}

// Helper function to create a test expectation
func createTestExpectation(exchange, routingKey string, opts ...expectations.ExpectationOption) (*expectations.Expectation, error) {
	// Create a body comparator
	bodyComparator, err := comparators.NewJSONBody([]byte(`{"foo":"bar"}`), comparators.MatchTypePartial)
	if err != nil {
		return nil, err
	}

	// Create a request
	request, err := expectations.NewRequest(exchange, routingKey, bodyComparator)
	if err != nil {
		return nil, err
	}

	// Create a response
	response, err := expectations.NewResponse([]byte(`{"result":"success"}`))
	if err != nil {
		return nil, err
	}

	// Create an expectation
	return expectations.NewExpectation(request, response, opts...)
}

// Helper function to create a JSON struct for testing
func createJSONStruct(t *testing.T, jsonStr string) *structpb.Struct {
	t.Helper()

	var v map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &v)
	require.NoError(t, err)

	pbValue, err := structpb.NewStruct(v)
	require.NoError(t, err)

	return pbValue
}

// Helper function to create a JSON value for testing
func createJSONValue(t *testing.T, jsonStr string) *structpb.Value {
	t.Helper()

	var v interface{}
	err := json.Unmarshal([]byte(jsonStr), &v)
	require.NoError(t, err)

	pbValue, err := structpb.NewValue(v)
	require.NoError(t, err)

	return pbValue
}

// MockExpectationsService is a simple implementation of the ExpectationsService interface for testing
type MockExpectationsService struct {
	expectations []*expectations.Expectation
	resetCalled  bool
}

func (s *MockExpectationsService) Create(exp *expectations.Expectation) error {
	s.expectations = append(s.expectations, exp)
	return nil
}

func (s *MockExpectationsService) Reset() {
	s.resetCalled = true
	s.expectations = nil
}

func (s *MockExpectationsService) Match(_ *expectations.Candidate) *expectations.Response {
	return nil
}

func (s *MockExpectationsService) GetExpectations(_ app.GetExpectationsRequest) []*expectations.Expectation {
	return s.expectations
}

func (s *MockExpectationsService) GetExpectation(id uuid.UUID) *expectations.Expectation {
	for _, exp := range s.expectations {
		if exp.ID == id {
			return exp
		}
	}
	return nil
}

func (s *MockExpectationsService) GetAssertions(_ app.GetAssertionsRequest) []*expectations.Assertion {
	return nil
}

// TestCreateExpectation tests the CreateExpectation handler
func TestCreateExpectation(t *testing.T) {
	// Create a mock expectations service
	mockSvc := &MockExpectationsService{}

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		expectationsService: mockSvc,
	}

	// Create a test request
	req := &grpcApi.CreateExpectationRequest{
		Request: &grpcApi.Request{
			Exchange:   "test-exchange",
			RoutingKey: "test-routing-key",
			Body: &grpcApi.Request_JsonBody{
				JsonBody: &grpcApi.JSONBodyAssertion{
					MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
					Body:      createJSONStruct(t, `{"foo":"bar"}`),
				},
			},
		},
		Response: &grpcApi.Response{
			Body: createJSONValue(t, `{"result":"success"}`),
		},
	}

	// Call the handler
	resp, err := server.CreateExpectation(context.Background(), req)

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.ExpectationId)

	// Verify the expectation was created in the service
	assert.Len(t, mockSvc.expectations, 1)
	assert.Equal(t, "test-exchange", mockSvc.expectations[0].Request.Exchange)
	assert.Equal(t, "test-routing-key", mockSvc.expectations[0].Request.RoutingKey)
}

// TestResetExpectations tests the ResetExpectations handler
func TestResetExpectations(t *testing.T) {
	// Create a mock expectations service
	mockSvc := &MockExpectationsService{}

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		expectationsService: mockSvc,
	}

	// Call the handler
	resp, err := server.ResetExpectations(context.Background(), &grpcApi.ResetExpectationsRequest{})

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify the reset was called on the service
	assert.True(t, mockSvc.resetCalled)
}

// TestGetExpectations tests the GetExpectations handler
func TestGetExpectations(t *testing.T) {
	// Create a mock expectations service with some test expectations
	mockSvc := &MockExpectationsService{}

	// Add some test expectations
	exp1, err := createTestExpectation("exchange1", "rk1")
	require.NoError(t, err)
	err = mockSvc.Create(exp1)
	require.NoError(t, err)

	exp2, err := createTestExpectation("exchange2", "rk2")
	require.NoError(t, err)
	err = mockSvc.Create(exp2)
	require.NoError(t, err)

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		expectationsService: mockSvc,
	}

	// Call the handler
	resp, err := server.GetExpectations(context.Background(), &grpcApi.GetExpectationsRequest{})

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Expectations, 2)

	// Verify the expectations in the response
	assert.Equal(t, exp1.ID.String(), resp.Expectations[0].Id)
	assert.Equal(t, exp2.ID.String(), resp.Expectations[1].Id)
}

// TestGetExpectation tests the GetExpectation handler
func TestGetExpectation(t *testing.T) {
	// Create a mock expectations service with a test expectation
	mockSvc := &MockExpectationsService{}

	// Add a test expectation
	exp, err := createTestExpectation("exchange", "rk")
	require.NoError(t, err)
	err = mockSvc.Create(exp)
	require.NoError(t, err)

	// Create the server with the mock service
	server := &AmqpMockServerServiceServer{
		expectationsService: mockSvc,
	}

	// Call the handler with a valid expectation ID
	resp, err := server.GetExpectation(context.Background(), &grpcApi.GetExpectationRequest{
		ExpectationId: exp.ID.String(),
	})

	// Verify the response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, exp.ID.String(), resp.Expectation.Id)

	// Call the handler with an invalid expectation ID
	_, err = server.GetExpectation(context.Background(), &grpcApi.GetExpectationRequest{
		ExpectationId: "invalid-id",
	})

	// Verify the error
	require.Error(t, err)

	// Call the handler with a non-existent expectation ID
	nonExistentID := uuid.New()
	_, err = server.GetExpectation(context.Background(), &grpcApi.GetExpectationRequest{
		ExpectationId: nonExistentID.String(),
	})

	// Verify the error
	require.Error(t, err)
}
