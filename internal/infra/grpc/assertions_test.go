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
)

// TestNewProtoAssertionInAssertions tests the newProtoAssertion function in the assertions context
func TestNewProtoAssertionInAssertions(t *testing.T) {
	// Create a test candidate
	exchange := "test-exchange"
	routingKey := "test-routing-key"
	body := []byte(`{"foo":"bar"}`)

	candidate, err := expectations.NewCandidate(exchange, routingKey, body)
	require.NoError(t, err)

	// Create an assertion without a matched expectation
	assertion := &expectations.Assertion{
		Candidate: candidate,
		CreatedAt: time.Now(),
	}

	// Convert to proto without including expectation
	protoAssertion := newProtoAssertion(assertion, nil)
	require.NotNil(t, protoAssertion)

	// Verify the conversion
	assert.NotEmpty(t, protoAssertion.Id)
	assert.Equal(t, exchange, protoAssertion.Candidate.Exchange)
	assert.Equal(t, routingKey, protoAssertion.Candidate.RoutingKey)
	assert.NotNil(t, protoAssertion.Candidate.Body)
	assert.Equal(t, assertion.CreatedAt.Format(time.RFC3339), protoAssertion.CreatedAt)
	assert.Nil(t, protoAssertion.Expectation)

	// Verify the body
	bodyJSON, err := protoAssertion.Candidate.Body.MarshalJSON()
	require.NoError(t, err)

	var originalMap, protoMap map[string]interface{}
	err = json.Unmarshal(body, &originalMap)
	require.NoError(t, err)
	err = json.Unmarshal(bodyJSON, &protoMap)
	require.NoError(t, err)

	assert.Equal(t, originalMap, protoMap)

	// Test with a matched expectation
	// Create a test expectation
	jsonBody := []byte(`{"foo":"bar"}`)
	bodyComparator, err := comparators.NewJSONBody(jsonBody, comparators.MatchTypePartial)
	require.NoError(t, err)

	request, err := expectations.NewRequest(exchange, routingKey, bodyComparator)
	require.NoError(t, err)

	response, err := expectations.NewResponse([]byte(`{"result":"success"}`))
	require.NoError(t, err)

	exp, err := expectations.NewExpectation(request, response)
	require.NoError(t, err)

	// Create an assertion with a matched expectation
	assertionWithMatch := &expectations.Assertion{
		Candidate:   candidate,
		CreatedAt:   time.Now(),
		Expectation: exp,
	}

	// Test with include=expectation
	protoAssertionWithMatch := newProtoAssertion(assertionWithMatch, []string{"expectation"})
	require.NotNil(t, protoAssertionWithMatch)

	// Verify the conversion
	assert.NotEmpty(t, protoAssertionWithMatch.Id)
	assert.Equal(t, exchange, protoAssertionWithMatch.Candidate.Exchange)
	assert.Equal(t, routingKey, protoAssertionWithMatch.Candidate.RoutingKey)
	assert.NotNil(t, protoAssertionWithMatch.Candidate.Body)
	assert.Equal(t, assertionWithMatch.CreatedAt.Format(time.RFC3339), protoAssertionWithMatch.CreatedAt)
	assert.NotNil(t, protoAssertionWithMatch.Expectation)
	assert.Equal(t, exp.ID.String(), protoAssertionWithMatch.Expectation.Id)

	// Test without include=expectation
	protoAssertionWithoutInclude := newProtoAssertion(assertionWithMatch, []string{})
	require.NotNil(t, protoAssertionWithoutInclude)

	// Verify the conversion
	assert.NotEmpty(t, protoAssertionWithoutInclude.Id)
	assert.Equal(t, exchange, protoAssertionWithoutInclude.Candidate.Exchange)
	assert.Equal(t, routingKey, protoAssertionWithoutInclude.Candidate.RoutingKey)
	assert.NotNil(t, protoAssertionWithoutInclude.Candidate.Body)
	assert.Equal(t, assertionWithMatch.CreatedAt.Format(time.RFC3339), protoAssertionWithoutInclude.CreatedAt)
	assert.Nil(t, protoAssertionWithoutInclude.Expectation)
}

// TestExpectationsService is a simple implementation of the ExpectationsService interface for testing
type TestExpectationsService struct {
	assertions []*expectations.Assertion
}

func (s *TestExpectationsService) Create(_ *expectations.Expectation) error {
	return nil
}

func (s *TestExpectationsService) Reset() {
}

func (s *TestExpectationsService) Match(_ *expectations.Candidate) *expectations.Response {
	return nil
}

func (s *TestExpectationsService) GetExpectations(_ app.GetExpectationsRequest) []*expectations.Expectation {
	return nil
}

func (s *TestExpectationsService) GetExpectation(_ uuid.UUID) *expectations.Expectation {
	return nil
}

func (s *TestExpectationsService) GetAssertions(req app.GetAssertionsRequest) []*expectations.Assertion {
	// Filter assertions based on request parameters
	var result []*expectations.Assertion

	// If ExpectationID is provided, filter by specific expectation
	if req.ExpectationID != nil {
		for _, a := range s.assertions {
			if a.Expectation != nil && a.Expectation.ID == *req.ExpectationID {
				result = append(result, a)
			}
		}
		return result
	}

	// If Status is provided, filter by match status
	if req.Status != nil {
		switch *req.Status {
		case "matched":
			for _, a := range s.assertions {
				if a.Expectation != nil {
					result = append(result, a)
				}
			}
		case "unmatched":
			for _, a := range s.assertions {
				if a.Expectation == nil {
					result = append(result, a)
				}
			}
		}
		return result
	}

	// Return all assertions
	return s.assertions
}

// TestGetAssertions tests the GetAssertions handler
func TestGetAssertions(t *testing.T) {
	// Create test data
	exchange := "test-exchange"
	routingKey := "test-routing-key"
	body := []byte(`{"foo":"bar"}`)

	candidate, err := expectations.NewCandidate(exchange, routingKey, body)
	require.NoError(t, err)

	// Create an assertion without a matched expectation
	unmatchedAssertion := &expectations.Assertion{
		Candidate: candidate,
		CreatedAt: time.Now(),
	}

	// Create a test expectation
	jsonBody := []byte(`{"foo":"bar"}`)
	bodyComparator, err := comparators.NewJSONBody(jsonBody, comparators.MatchTypePartial)
	require.NoError(t, err)

	request, err := expectations.NewRequest(exchange, routingKey, bodyComparator)
	require.NoError(t, err)

	response, err := expectations.NewResponse([]byte(`{"result":"success"}`))
	require.NoError(t, err)

	exp, err := expectations.NewExpectation(request, response)
	require.NoError(t, err)

	// Create an assertion with a matched expectation
	matchedAssertion := &expectations.Assertion{
		Candidate:   candidate,
		CreatedAt:   time.Now(),
		Expectation: exp,
	}

	// Create the test service with our test assertions
	testSvc := &TestExpectationsService{
		assertions: []*expectations.Assertion{unmatchedAssertion, matchedAssertion},
	}

	// Create the server with the test service
	server := &AmqpMockServerServiceServer{
		expectationsService: testSvc,
	}

	// Test cases
	testCases := []struct {
		name                string
		request             *grpcApi.GetAssertionsRequest
		expectedCount       int
		expectedStatus      string
		includeExpectations bool
	}{
		{
			name:          "Get all assertions",
			request:       &grpcApi.GetAssertionsRequest{},
			expectedCount: 2,
		},
		{
			name: "Get matched assertions",
			request: &grpcApi.GetAssertionsRequest{
				Status: stringPtr("matched"),
			},
			expectedCount:  1,
			expectedStatus: "matched",
		},
		{
			name: "Get unmatched assertions",
			request: &grpcApi.GetAssertionsRequest{
				Status: stringPtr("unmatched"),
			},
			expectedCount:  1,
			expectedStatus: "unmatched",
		},
		{
			name: "Get assertions for specific expectation",
			request: &grpcApi.GetAssertionsRequest{
				ExpectationId: stringPtr(exp.ID.String()),
			},
			expectedCount: 1,
		},
		{
			name: "Get all assertions with include=expectation",
			request: &grpcApi.GetAssertionsRequest{
				Include: []string{"expectation"},
			},
			expectedCount:       2,
			includeExpectations: true,
		},
		{
			name: "Get matched assertions with include=expectation",
			request: &grpcApi.GetAssertionsRequest{
				Status:  stringPtr("matched"),
				Include: []string{"expectation"},
			},
			expectedCount:       1,
			expectedStatus:      "matched",
			includeExpectations: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the handler
			resp, err := server.GetAssertions(context.Background(), tc.request)

			// Verify the response
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Assertions, tc.expectedCount)

			// Additional verification based on the test case
			if tc.expectedStatus == "matched" {
				for _, a := range resp.Assertions {
					if tc.includeExpectations {
						assert.NotNil(t, a.Expectation, "Expectation should be included")
					} else {
						assert.Nil(t, a.Expectation, "Expectation should not be included")
					}
				}
			} else if tc.expectedStatus == "unmatched" {
				for _, a := range resp.Assertions {
					assert.Nil(t, a.Expectation, "Unmatched assertions should never have expectations")
				}
			} else if tc.includeExpectations {
				// For mixed or unspecified status with include=expectation
				matchedFound := false
				for _, a := range resp.Assertions {
					if a.Expectation != nil {
						matchedFound = true
						break
					}
				}
				assert.True(t, matchedFound, "At least one assertion should have an expectation when include=expectation")
			} else {
				// For mixed or unspecified status without include=expectation
				for _, a := range resp.Assertions {
					assert.Nil(t, a.Expectation, "Expectation should not be included")
				}
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
