package comparators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegex(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		regex    string
		expError string
	}{
		"success": {
			regex: `starts_with_.*`,
		},
		"bad regex": {
			regex:    `starts_with_.*[`,
			expError: "error parsing regexp",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := NewRegex(tt.regex)

			if tt.expError != "" {
				assert.ErrorContains(t, err, tt.expError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegex_Match(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		regex    string
		payload  []byte
		expMatch bool
	}{
		"matches": {
			regex:    `^starts_with_.*`,
			payload:  []byte(`starts_with_123`),
			expMatch: true,
		},
		"no match": {
			regex:   `^starts_with_.*`,
			payload: []byte(`not_starts_with_123`),
		},
		"match in the middle": {
			regex:    `contains`,
			payload:  []byte(`something that contains the word`),
			expMatch: true,
		},
		"match complex regex": {
			regex:    `^starts_.*\d{3}$`,
			payload:  []byte(`starts_with_123`),
			expMatch: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, err := NewRegex(tt.regex)
			require.NoError(t, err)

			assert.Equal(t, tt.expMatch, r.Match(tt.payload))
		})
	}
}
