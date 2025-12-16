package app_test

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/dialecticanet-com/rmq-rpc-mockserver/internal/app"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/comparators"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testComparator, _ = comparators.NewRegex("foo")

func TestExpectationsService_Match(t *testing.T) {
	t.Parallel()

	svc := newExpectationsService(t, []*expectations.Expectation{
		// just normal expectation
		newTestExpectation(t, "exchange", "rk", []byte("body1")),
		// expectation with another routing key
		newTestExpectation(t, "exchange", "rk2", []byte("body2")),
		// expectation equal to the first one but with higher priority
		newTestExpectation(t, "exchange", "rk", []byte("body3"), expectations.WithPriority(10)), // the same path as the first one
		// expectation with TTL
		newTestExpectation(t, "exchange", "rk", []byte("body4"), expectations.WithTimeToLive(10*time.Millisecond)),
	})

	// expectations added
	assert.Len(t, svc.GetExpectations(GetExpectationsRequest{}), 4)

	// this candidate does not have a match
	candidate := newTestCandidate(t, "exchange", "rk", []byte("bar"))
	resp := svc.Match(candidate)
	assert.Nil(t, resp)

	// this candidate has a match
	candidate = newTestCandidate(t, "exchange", "rk", []byte("foo"))

	// higher priority should be used
	resp = svc.Match(candidate)
	require.NotNil(t, resp)
	assert.Equal(t, json.RawMessage("body3"), resp.Body)

	// wait for TTL so the last expectation will expire
	time.Sleep(50 * time.Millisecond)

	// since higher priority was used, the first expectation should be used
	resp = svc.Match(candidate)
	require.NotNil(t, resp)
	assert.Equal(t, json.RawMessage("body1"), resp.Body)

	// this candidate has a match on rk2
	candidate = newTestCandidate(t, "exchange", "rk2", json.RawMessage("foo"))

	// expectation with TTL should not be used since it's expired
	resp = svc.Match(candidate)
	require.NotNil(t, resp)
	assert.Equal(t, json.RawMessage("body2"), resp.Body)

	// since there are no more expectations, nil should be returned
	resp = svc.Match(candidate)
	assert.Nil(t, resp)
}

func TestExpectationsService_Reset(t *testing.T) {
	t.Parallel()

	svc := newExpectationsService(t, []*expectations.Expectation{
		newTestExpectation(t, "exchange", "rk", []byte("body1")),
		newTestExpectation(t, "exchange", "rk2", []byte("body2")),
	})

	assert.Len(t, svc.GetExpectations(GetExpectationsRequest{}), 2)

	svc.Reset()

	assert.Len(t, svc.GetExpectations(GetExpectationsRequest{}), 0)
}

func TestExpectationsService_GetExpectations(t *testing.T) {
	t.Parallel()

	svc := newExpectationsService(t, []*expectations.Expectation{
		newTestExpectation(t, "exchange", "rk", []byte("body1")),
		newTestExpectation(t, "exchange", "rk2", []byte("body2"), expectations.WithTimeToLive(10*time.Millisecond)),
	})

	time.Sleep(20 * time.Millisecond)

	assert.Len(t, svc.GetExpectations(GetExpectationsRequest{}), 2)

	expiredExps := svc.GetExpectations(GetExpectationsRequest{Status: ptrOf("expired")})
	activeExps := svc.GetExpectations(GetExpectationsRequest{Status: ptrOf("active")})

	require.Len(t, expiredExps, 1)
	require.Len(t, activeExps, 1)
	assert.Equal(t, json.RawMessage("body1"), activeExps[0].Response.Body)
	assert.Equal(t, json.RawMessage("body2"), expiredExps[0].Response.Body)
}

func TestExpectationsService_GetAssertions(t *testing.T) {
	t.Parallel()
	svc := newExpectationsService(t, []*expectations.Expectation{
		newTestExpectation(t, "exchange", "rk", []byte("body1")),
		newTestExpectation(t, "exchange", "rk2", []byte("body2")),
	})

	candidate := newTestCandidate(t, "exchange", "rk", []byte("foo"))
	svc.Match(candidate)
	candidate = newTestCandidate(t, "exchange", "rk333", []byte("foo2")) // should not match
	svc.Match(candidate)

	assert.Len(t, svc.GetAssertions(GetAssertionsRequest{}), 2)
	assert.Len(t, svc.GetAssertions(GetAssertionsRequest{Status: ptrOf("matched")}), 1)
	assert.Len(t, svc.GetAssertions(GetAssertionsRequest{Status: ptrOf("unmatched")}), 1)
	assert.Len(t, svc.GetAssertions(GetAssertionsRequest{ExpectationID: ptrOf(svc.GetExpectations(GetExpectationsRequest{})[0].ID)}), 1)
}

func newTestCandidate(t *testing.T, exc, rk string, body []byte) *expectations.Candidate {
	t.Helper()

	candidate, err := expectations.NewCandidate(exc, rk, body)
	require.NoError(t, err)

	return candidate
}

func newExpectationsService(t *testing.T, exps []*expectations.Expectation) *ExpectationsService {
	t.Helper()

	s := NewExpectationsService()
	for _, exp := range exps {
		err := s.Create(exp)
		require.NoError(t, err)
	}

	return s
}

// nolint: unparam
func newTestExpectation(t *testing.T, exc, rk string, respBody []byte, opts ...expectations.ExpectationOption) *expectations.Expectation {
	t.Helper()

	req, err := expectations.NewRequest(exc, rk, testComparator)
	require.NoError(t, err)

	res, err := expectations.NewResponse(respBody)
	require.NoError(t, err)

	exp, err := expectations.NewExpectation(req, res, opts...)
	require.NoError(t, err)

	return exp
}

func ptrOf[T any](v T) *T {
	return &v
}
