package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadTimeout(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		rt          time.Duration
		expectedErr string
	}{
		"success":      {rt: time.Second},
		"missing cert": {rt: -1 * time.Second, expectedErr: "negative or zero read timeout provided"},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			opts := &serverOptions{}
			err := ServerWithReadTimeout(tt.rt)(opts)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.rt, opts.readTimeout)
			}
		})
	}
}

func TestWriteTimeout(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		wt          time.Duration
		expectedErr string
	}{
		"success":      {wt: time.Second},
		"missing cert": {wt: -1 * time.Second, expectedErr: "negative or zero write timeout provided"},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			opts := &serverOptions{}
			err := ServerWithWriteTimeout(tt.wt)(opts)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wt, opts.writeTimeout)
			}
		})
	}
}

func TestHandlerTimeout(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		wt          time.Duration
		expectedErr string
	}{
		"success":      {wt: time.Second},
		"missing cert": {wt: -1 * time.Second, expectedErr: "negative or zero handler timeout provided"},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			opts := &serverOptions{}
			err := ServerWithHandlerTimeout(tt.wt)(opts)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wt, opts.handlerTimeout)
			}
		})
	}
}

func TestShutdownGracePeriod(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		gp          time.Duration
		expectedErr string
	}{
		"success":      {gp: time.Second},
		"missing cert": {gp: -1 * time.Second, expectedErr: "negative or zero shutdown grace period timeout provided"},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			opts := &serverOptions{}
			err := ServerWithShutdownGracePeriod(tt.gp)(opts)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.gp, opts.shutdownGracePeriod)
			}
		})
	}
}

func TestPort(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		port        int
		expectedErr string
	}{
		"success":      {port: 50000},
		"missing cert": {port: 120000, expectedErr: "invalid HTTP ServerWithPort provided"},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			opts := &serverOptions{}
			err := ServerWithPort(tt.port)(opts)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.port, opts.port)
			}
		})
	}
}
