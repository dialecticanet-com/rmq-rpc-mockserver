package expectations

import (
	"encoding/json"
	"strings"
)

type BodyComparator interface {
	Match(payload []byte) bool
}

type Request struct {
	Exchange       string
	RoutingKey     string
	BodyComparator BodyComparator
}

func NewRequest(exchange, routingKey string, bodyCmp BodyComparator) (*Request, error) {
	if exchange == "" {
		return nil, ErrEmptyExchange
	}

	if routingKey == "" {
		return nil, ErrEmptyRoutingKey
	}

	return &Request{
		Exchange:       exchange,
		RoutingKey:     routingKey,
		BodyComparator: bodyCmp,
	}, nil
}

func (r *Request) Matches(cnd *Candidate) bool {
	return r.Exchange == cnd.Exchange && r.RoutingKey == cnd.RoutingKey && r.BodyComparator.Match(cnd.Body)
}

func (r *Request) FormattedBody(offset int) string {
	raw, _ := json.MarshalIndent(r.BodyComparator, strings.Repeat(" ", offset), "  ")
	return string(raw)
}
