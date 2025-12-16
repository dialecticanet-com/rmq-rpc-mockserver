package expectations

import (
	"testing"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/comparators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	t.Parallel()
	bodyCmp, err := comparators.NewRegex("foo")
	require.NoError(t, err)

	testCases := map[string]struct {
		exchange string
		rk       string
		expError error
	}{
		"success": {
			exchange: "exchange",
			rk:       "rk",
		},
		"empty exchange": {
			rk:       "rk",
			expError: ErrEmptyExchange,
		},
		"empty routing key": {
			exchange: "exchange",
			expError: ErrEmptyRoutingKey,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := NewRequest(tt.exchange, tt.rk, bodyCmp)

			if tt.expError != nil {
				assert.ErrorIs(t, err, tt.expError)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, req)
				assert.Equal(t, tt.exchange, req.Exchange)
				assert.Equal(t, tt.rk, req.RoutingKey)
				assert.Equal(t, bodyCmp, req.BodyComparator)
			}
		})
	}
}

func TestRequest_Matches(t *testing.T) {
	t.Parallel()
	exchange := "exchange"
	routingKey := "rk"
	bodyCmp, err := comparators.NewRegex("foo")
	require.NoError(t, err)

	req, err := NewRequest(exchange, routingKey, bodyCmp)
	require.NoError(t, err)

	testCases := map[string]struct {
		exchange string
		rk       string
		body     []byte
		matches  bool
	}{
		"success": {
			exchange: "exchange",
			rk:       "rk",
			body:     []byte("something_foo_something"),
			matches:  true,
		},
		"exchange mismatch": {
			exchange: "exchange2",
			rk:       "rk",
			body:     []byte("something_foo_something"),
		},
		"routing key mismatch": {
			exchange: "exchange",
			rk:       "rk2",
			body:     []byte("something_foo_something"),
		},
		"body mismatch": {
			exchange: "exchange",
			rk:       "rk",
			body:     []byte("something_bar_something"),
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cnd, err := NewCandidate(tt.exchange, tt.rk, tt.body)
			require.NoError(t, err)

			assert.Equal(t, tt.matches, req.Matches(cnd))
		})
	}
}
