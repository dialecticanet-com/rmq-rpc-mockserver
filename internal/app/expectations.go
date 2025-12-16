package app

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/google/uuid"
)

// ExpectationsService is the application level service to manage expectations.
type ExpectationsService struct {
	m            sync.RWMutex
	expectations []*expectations.Expectation
	assertions   expectations.Assertions
}

// NewExpectationsService creates a new ExpectationsService instance.
func NewExpectationsService() *ExpectationsService {
	return &ExpectationsService{}
}

// Create creates a new expectation.
func (s *ExpectationsService) Create(exp *expectations.Expectation) error {
	s.m.Lock()
	defer s.m.Unlock()

	s.expectations = append(s.expectations, exp)
	s.log(
		fmt.Sprintf("Expectation created. ExpectationID=%s, Exchange=%s, RoutingKey=%s", exp.ID, exp.Request.Exchange, exp.Request.RoutingKey),
		fmt.Sprintf("REQUEST:\n   %s", exp.Request.FormattedBody(3)),
	)

	if exp.TimeToLive != nil && exp.TimeToLive.TTL > 0 {
		go s.informExpectationExpired(exp.ID, exp.TimeToLive.TTL)
	}

	return nil
}

// Match matches a candidate against the expectations.
func (s *ExpectationsService) Match(candidate *expectations.Candidate) *expectations.Response {
	s.m.Lock()
	defer s.m.Unlock()

	matches := make([]*expectations.Expectation, 0)
	for _, exp := range s.expectations {
		if exp.Matches(candidate) {
			matches = append(matches, exp)
		}
	}

	if len(matches) == 0 {
		s.assertions.Add(expectations.NewUnmatchedAssertion(candidate))
		s.log(
			fmt.Sprintf("NO MATCH FOUND. Exchange: %s, RoutingKey: %s", candidate.Exchange, candidate.RoutingKey),
			fmt.Sprintf("REQUEST:\n   %s", candidate.FormattedBody(3)),
		)
		return nil
	}

	// sort matches by priority
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Priority > matches[j].Priority
	})

	matches[0].Use()
	s.assertions.Add(expectations.NewMatchedAssertion(candidate, matches[0]))
	s.log(
		fmt.Sprintf("MATCH FOUND. ExpectationID=%s, Exchange: %s, RoutingKey: %s", matches[0].ID, candidate.Exchange, candidate.RoutingKey),
		fmt.Sprintf("REQUEST:\n   %s", candidate.FormattedBody(3)),
		fmt.Sprintf("RESPONSE:\n   %s", matches[0].Response.FormattedBody(3)),
	)

	if !matches[0].IsActive() {
		s.log(fmt.Sprintf("Expectation usage limit reached. ExpectationID=%s", matches[0].ID))
	}

	return matches[0].Response
}

func (s *ExpectationsService) informExpectationExpired(id uuid.UUID, ttl time.Duration) {
	time.Sleep(ttl)
	s.log(fmt.Sprintf("Expectation expired. ExpectationID=%s, TTL=%v", id, ttl.Seconds()))
}

// Reset removes all expectations from the service.
func (s *ExpectationsService) Reset() {
	s.m.Lock()
	defer s.m.Unlock()

	s.expectations = nil
}

func (s *ExpectationsService) log(lines ...string) {
	payload := strings.Builder{}
	for i, line := range lines {
		prefix := "   "
		if i == 0 {
			prefix = "-> "
		}
		payload.WriteString(prefix + line)
		if i < len(lines)-1 {
			payload.WriteString("\n")
		}
	}

	fmt.Println(payload.String())
}

// GetExpectationsRequest represents the parameters for filtering expectations.
type GetExpectationsRequest struct {
	Status *string // "active" or "expired"
}

// GetExpectations returns expectations filtered by the given request parameters.
func (s *ExpectationsService) GetExpectations(req GetExpectationsRequest) []*expectations.Expectation {
	s.m.RLock()
	defer s.m.RUnlock()

	if req.Status == nil {
		// Return all expectations if status is not specified
		result := make([]*expectations.Expectation, len(s.expectations))
		for i, exp := range s.expectations {
			result[i] = exp.Copy()
		}

		return result
	}

	// Filter expectations based on status
	var result []*expectations.Expectation
	for _, exp := range s.expectations {
		switch *req.Status {
		case "active":
			if exp.IsActive() {
				result = append(result, exp.Copy())
			}
		case "expired":
			if !exp.IsActive() {
				result = append(result, exp.Copy())
			}
		}
	}

	return result
}

func (s *ExpectationsService) GetExpectation(id uuid.UUID) *expectations.Expectation {
	s.m.Lock()
	defer s.m.Unlock()

	for _, exp := range s.expectations {
		if exp.ID == id {
			return exp
		}
	}

	return nil
}

type GetAssertionsRequest struct {
	ExpectationID *uuid.UUID
	Status        *string
	Include       []string
}

func (s *ExpectationsService) GetAssertions(req GetAssertionsRequest) []*expectations.Assertion {
	s.m.RLock()
	defer s.m.RUnlock()

	// If ExpectationID is provided, filter by specific expectation
	if req.ExpectationID != nil {
		return s.assertions.GetForExpectationByID(req.ExpectationID.String())
	}

	// If Status is provided, filter by match status
	if req.Status != nil {
		switch *req.Status {
		case "matched":
			return s.assertions.GetMatched()
		case "unmatched":
			return s.assertions.GetUnmatched()
		}
	}

	return s.assertions.GetAll()
}
