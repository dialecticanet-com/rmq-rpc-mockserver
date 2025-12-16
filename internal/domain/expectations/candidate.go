package expectations

import (
	"encoding/json"
	"strings"
)

// Candidate represents a candidate message to match against expectations.
type Candidate struct {
	Exchange   string
	RoutingKey string
	Body       json.RawMessage
}

// NewCandidate creates a new Candidate instance.
func NewCandidate(exchange, rk string, body json.RawMessage) (*Candidate, error) {
	if exchange == "" {
		return nil, ErrEmptyExchange
	}

	if rk == "" {
		return nil, ErrEmptyRoutingKey
	}

	return &Candidate{
		Exchange:   exchange,
		RoutingKey: rk,
		Body:       body,
	}, nil
}

func (c *Candidate) FormattedBody(offset int) string {
	raw, _ := json.MarshalIndent(c.Body, strings.Repeat(" ", offset), "  ")
	return string(raw)
}
