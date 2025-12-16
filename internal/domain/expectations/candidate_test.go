package expectations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCandidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		exchange string
		rk       string
		body     json.RawMessage
		expError error
	}{
		"success": {
			exchange: "exchange",
			rk:       "rk",
			body:     []byte("body"),
		},
		"empty exchange": {
			rk:       "rk",
			body:     []byte("body"),
			expError: ErrEmptyExchange,
		},
		"empty routing key": {
			exchange: "exchange",
			body:     []byte("body"),
			expError: ErrEmptyRoutingKey,
		},
		"empty body": {
			exchange: "exchange",
			rk:       "rk",
			body:     nil,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cnd, err := NewCandidate(tt.exchange, tt.rk, tt.body)

			if tt.expError != nil {
				assert.ErrorIs(t, err, tt.expError)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, cnd)
				assert.Equal(t, tt.exchange, cnd.Exchange)
				assert.Equal(t, tt.rk, cnd.RoutingKey)
				assert.Equal(t, tt.body, cnd.Body)
			}
		})
	}
}
