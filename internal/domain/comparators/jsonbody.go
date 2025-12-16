package comparators

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/wI2L/jsondiff"
)

// MatchType represents what JSON body match type to use.
type MatchType string

const (
	// MatchTypeExact represents an exact match.
	MatchTypeExact MatchType = "EXACT"
	// MatchTypePartial represents a partial match - only fields that exist in the expectations will be matched.
	MatchTypePartial MatchType = "PARTIAL"
)

// JSONBody represents a JSON body to compare.
type JSONBody struct {
	Body      json.RawMessage
	MatchType MatchType
}

// NewJSONBody creates a new JSONBody instance.
func NewJSONBody(body json.RawMessage, matchType MatchType) (*JSONBody, error) {
	ok := json.Valid(body)
	if !ok {
		return nil, errors.New("invalid json body")
	}

	return &JSONBody{Body: body, MatchType: matchType}, nil
}

// Match matches the JSON body against a payload.
func (b *JSONBody) Match(payload []byte) bool {
	if !json.Valid(payload) {
		return false
	}

	if b.MatchType == MatchTypeExact {
		return b.matchStrict(payload)
	}

	if b.MatchType == MatchTypePartial {
		return b.matchSubset(payload)
	}

	return false
}

func (b *JSONBody) matchStrict(payload []byte) bool {
	var actualReq any
	_ = json.Unmarshal(payload, &actualReq)

	var expectedReq any
	_ = json.Unmarshal(b.Body, &expectedReq)

	return reflect.DeepEqual(expectedReq, actualReq)
}

func (b *JSONBody) matchSubset(payload []byte) bool {
	result, err := jsondiff.CompareJSON(b.Body, payload)
	if err != nil {
		return false
	}

	for _, change := range result {
		if change.Type != jsondiff.OperationAdd {
			return false
		}
	}

	return true
}
