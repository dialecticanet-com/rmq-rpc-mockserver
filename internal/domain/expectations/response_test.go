package expectations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		body json.RawMessage
	}{
		"success": {
			body: json.RawMessage("body"),
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res, err := NewResponse(tt.body)
			assert.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tt.body, res.Body)
		})
	}
}
