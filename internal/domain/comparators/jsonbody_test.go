package comparators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONBody(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		body     []byte
		expError string
	}{
		"success": {
			body: []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
		},
		"bad json": {
			body:     []byte(`foo`),
			expError: "invalid json body",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := NewJSONBody(tt.body, MatchTypeExact)
			if tt.expError != "" {
				assert.EqualError(t, err, tt.expError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSONBody_Equal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		jsonBody    []byte
		matchType   MatchType
		testPayload []byte
		equal       bool
	}{
		"equal strict": {
			jsonBody:    []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			testPayload: []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			matchType:   MatchTypeExact,
			equal:       true,
		},
		"equal subset": {
			jsonBody:    []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			testPayload: []byte(`{"a": 1, "b": 2, "c": {"d": 3, "e": 5}}`),
			matchType:   MatchTypePartial,
			equal:       true,
		},
		"not equal strict": {
			jsonBody:    []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			testPayload: []byte(`{"a": 2, "b": 2, "c": {"d": 3}}`),
			matchType:   MatchTypeExact,
			equal:       false,
		},
		"not equal subset": {
			jsonBody:    []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			testPayload: []byte(`{"a": 1, "b": 3}`),
			matchType:   MatchTypePartial,
			equal:       false,
		},
		"not equal subset additional field": {
			jsonBody:    []byte(`{"a": 1, "b": 2, "c": {"d": 3, "e": 5}}`),
			testPayload: []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			matchType:   MatchTypePartial,
			equal:       false,
		},
		"bad json payload internals - subset": {
			jsonBody:    []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			testPayload: []byte(`foo`),
			matchType:   MatchTypePartial,
			equal:       false,
		},
		"bad json payload internals - strict": {
			jsonBody:    []byte(`{"a": 1, "b": 2, "c": {"d": 3}}`),
			testPayload: []byte(`foo`),
			matchType:   MatchTypeExact,
			equal:       false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jsonBody, err := NewJSONBody(tt.jsonBody, tt.matchType)
			require.NoError(t, err)

			equal := jsonBody.Match(tt.testPayload)
			assert.Equal(t, tt.equal, equal)
		})
	}
}
