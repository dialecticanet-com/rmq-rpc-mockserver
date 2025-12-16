package grpc

import (
	"encoding/json"
	"testing"
	"time"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/comparators"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubscription(t *testing.T) {
	// Create a subscription
	queue := "test-queue"
	sub := subscriptions.NewSubscription(queue)

	// Convert to proto
	protoSub := newSubscription(sub)

	// Verify the conversion
	assert.Equal(t, sub.ID().String(), protoSub.Id)
	assert.Equal(t, sub.Queue(), protoSub.Queue)
}

func TestNewProtoExpectation(t *testing.T) {
	// Create request with JSONBody comparator
	jsonBody := []byte(`{"foo":"bar"}`)
	bodyComparator, err := comparators.NewJSONBody(jsonBody, comparators.MatchTypePartial)
	require.NoError(t, err)

	request, err := expectations.NewRequest("test-exchange", "test-routing-key", bodyComparator)
	require.NoError(t, err)

	// Create response
	responseBody := []byte(`{"result":"success"}`)
	response, err := expectations.NewResponse(responseBody)
	require.NoError(t, err)

	t.Run("expectation with unlimited times", func(t *testing.T) {
		// Create expectation with unlimited times
		exp, err := expectations.NewExpectation(request, response, expectations.WithUnlimitedTimes())
		require.NoError(t, err)

		// Convert to proto
		protoExp := newProtoExpectation(exp)

		// Verify the conversion
		assert.Equal(t, exp.ID.String(), protoExp.Id)
		assert.Equal(t, exp.CreatedAt.Format(time.RFC3339), protoExp.CreatedAt)
		assert.NotNil(t, protoExp.Request)
		assert.NotNil(t, protoExp.Response)
		assert.NotNil(t, protoExp.Times)
		assert.True(t, protoExp.GetTimes().GetUnlimited())
		assert.Nil(t, protoExp.ExpiresAt)
	})

	t.Run("expectation with limited times", func(t *testing.T) {
		// Create expectation with limited times
		exp, err := expectations.NewExpectation(request, response, expectations.WithLimitedTimes(5))
		require.NoError(t, err)

		// Convert to proto
		protoExp := newProtoExpectation(exp)

		// Verify the conversion
		assert.Equal(t, exp.ID.String(), protoExp.Id)
		assert.NotNil(t, protoExp.Times)
		assert.Equal(t, uint32(5), protoExp.GetTimes().GetRemainingTimes())
	})

	t.Run("expectation with TTL", func(t *testing.T) {
		// Create expectation with TTL
		ttl := 10 * time.Minute
		exp, err := expectations.NewExpectation(request, response, expectations.WithTimeToLive(ttl))
		require.NoError(t, err)

		// Convert to proto
		protoExp := newProtoExpectation(exp)

		// Verify the conversion
		assert.Equal(t, exp.ID.String(), protoExp.Id)
		assert.NotNil(t, protoExp.ExpiresAt)

		// Calculate expected expiration time
		expectedExpiresAt := exp.CreatedAt.Add(ttl).Format(time.RFC3339)
		assert.Equal(t, expectedExpiresAt, *protoExp.ExpiresAt)
	})
}

func TestNewProtoRequest(t *testing.T) {
	exchange := "test-exchange"
	routingKey := "test-routing-key"

	t.Run("request with JSON body", func(t *testing.T) {
		// Create request with JSONBody comparator
		jsonBody := []byte(`{"foo":"bar"}`)
		bodyComparator, err := comparators.NewJSONBody(jsonBody, comparators.MatchTypePartial)
		require.NoError(t, err)

		request, err := expectations.NewRequest(exchange, routingKey, bodyComparator)
		require.NoError(t, err)

		// Convert to proto
		protoReq := newProtoRequest(request)

		// Verify the conversion
		assert.Equal(t, exchange, protoReq.Exchange)
		assert.Equal(t, routingKey, protoReq.RoutingKey)
		assert.NotNil(t, protoReq.GetJsonBody())
		assert.Equal(t, grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL, protoReq.GetJsonBody().MatchType)
	})

	t.Run("request with Regex body", func(t *testing.T) {
		// Create request with Regex comparator
		regexPattern := "foo.*bar"
		bodyComparator, err := comparators.NewRegex(regexPattern)
		require.NoError(t, err)

		request, err := expectations.NewRequest(exchange, routingKey, bodyComparator)
		require.NoError(t, err)

		// Convert to proto
		protoReq := newProtoRequest(request)

		// Verify the conversion
		assert.Equal(t, exchange, protoReq.Exchange)
		assert.Equal(t, routingKey, protoReq.RoutingKey)
		assert.NotNil(t, protoReq.GetRegexBody())
		assert.Equal(t, regexPattern, protoReq.GetRegexBody().Regex)
	})
}

func TestNewProtoMatchType(t *testing.T) {
	tests := []struct {
		name      string
		matchType comparators.MatchType
		expected  grpcApi.JSONBodyAssertion_MatchType
	}{
		{
			name:      "exact match",
			matchType: comparators.MatchTypeExact,
			expected:  grpcApi.JSONBodyAssertion_MATCH_TYPE_EXACT,
		},
		{
			name:      "partial match",
			matchType: comparators.MatchTypePartial,
			expected:  grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newProtoMatchType(tt.matchType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewProtoResponse(t *testing.T) {
	// Create response
	responseBody := []byte(`{"result":"success"}`)
	response, err := expectations.NewResponse(responseBody)
	require.NoError(t, err)

	// Convert to proto
	protoRes := newProtoResponse(response)
	require.NotNil(t, protoRes)

	// Verify the conversion
	assert.NotNil(t, protoRes)

	// Convert proto value back to JSON for comparison
	protoJSON, err := protoRes.Body.MarshalJSON()
	require.NoError(t, err)

	// Compare the JSON content
	var originalMap, protoMap map[string]interface{}
	err = json.Unmarshal(responseBody, &originalMap)
	require.NoError(t, err)
	err = json.Unmarshal(protoJSON, &protoMap)
	require.NoError(t, err)

	assert.Equal(t, originalMap, protoMap)
}

func TestNewProtoAssertion(t *testing.T) {
	// Create a candidate
	exchange := "test-exchange"
	routingKey := "test-routing-key"
	body := []byte(`{"foo":"bar"}`)

	candidate, err := expectations.NewCandidate(exchange, routingKey, body)
	require.NoError(t, err)

	// Create an assertion
	assertion := &expectations.Assertion{
		Candidate: candidate,
		CreatedAt: time.Now(),
	}

	t.Run("assertion without matched expectation", func(t *testing.T) {
		// Convert to proto without including expectation
		protoAssertion := newProtoAssertion(assertion, []string{})

		// Verify the conversion
		assert.NotEmpty(t, protoAssertion.Id)
		assert.Equal(t, exchange, protoAssertion.Candidate.Exchange)
		assert.Equal(t, routingKey, protoAssertion.Candidate.RoutingKey)
		assert.NotNil(t, protoAssertion.Candidate.Body)
		assert.Equal(t, assertion.CreatedAt.Format(time.RFC3339), protoAssertion.CreatedAt)
		assert.Nil(t, protoAssertion.Expectation)
	})

	t.Run("assertion with matched expectation", func(t *testing.T) {
		// Create a request with JSONBody comparator
		jsonBody := []byte(`{"foo":"bar"}`)
		bodyComparator, err := comparators.NewJSONBody(jsonBody, comparators.MatchTypePartial)
		require.NoError(t, err)

		request, err := expectations.NewRequest(exchange, routingKey, bodyComparator)
		require.NoError(t, err)

		// Create response
		responseBody := []byte(`{"result":"success"}`)
		response, err := expectations.NewResponse(responseBody)
		require.NoError(t, err)

		// Create expectation
		exp, err := expectations.NewExpectation(request, response)
		require.NoError(t, err)

		// Set matched expectation
		assertionWithMatch := &expectations.Assertion{
			Candidate:   candidate,
			CreatedAt:   time.Now(),
			Expectation: exp,
		}

		// Convert to proto with including expectation
		protoAssertion := newProtoAssertion(assertionWithMatch, []string{"expectation"})

		// Verify the conversion
		assert.NotEmpty(t, protoAssertion.Id)
		assert.NotNil(t, protoAssertion.Expectation)
		assert.Equal(t, exp.ID.String(), protoAssertion.Expectation.Id)
	})
}
