package expectations

import (
	"testing"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/comparators"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExpectation(t *testing.T) {
	t.Parallel()

	req, err := NewRequest("exchange", "rk", nil)
	require.NoError(t, err)

	resp, err := NewResponse([]byte("body"))
	require.NoError(t, err)

	t.Run("simple setup", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp)
		require.NoError(t, err)
		require.NotNil(t, exp)

		assert.Equal(t, req, exp.Request)
		assert.Equal(t, resp, exp.Response)
		assert.InDelta(t, time.Now().UnixMilli(), exp.CreatedAt.UnixMilli(), 100)
		assert.NoError(t, uuid.Validate(exp.ID.String()))
		assert.Equal(t, 0, exp.Priority)
		require.NotNil(t, exp.Times)
		assert.Equal(t, uint32(1), exp.Times.RemainingTimes)
		assert.False(t, exp.Times.Unlimited)
	})

	t.Run("with times", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithLimitedTimes(2))
		require.NoError(t, err)
		require.NotNil(t, exp)

		assert.Equal(t, uint32(2), exp.Times.RemainingTimes)
		assert.False(t, exp.Times.Unlimited)
	})

	t.Run("with priority", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithPriority(1))
		require.NoError(t, err)
		require.NotNil(t, exp)

		assert.Equal(t, 1, exp.Priority)
	})

	t.Run("with ttl", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithTimeToLive(time.Minute))
		require.NoError(t, err)
		require.NotNil(t, exp)

		assert.Equal(t, time.Minute, exp.TimeToLive.TTL)
	})

	t.Run("with indefinite times", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithUnlimitedTimes())
		require.NoError(t, err)
		require.NotNil(t, exp)

		assert.True(t, exp.Times.Unlimited)
	})

	t.Run("with invalid times", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithLimitedTimes(0))
		assert.ErrorIs(t, err, ErrBadRequestTimes)
		assert.Nil(t, exp)
	})

	t.Run("with invalid ttl", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithTimeToLive(0))
		assert.ErrorIs(t, err, ErrTTLMustBeGreaterThanZero)
		assert.Nil(t, exp)
	})
}

func TestExpectation_IsActive(t *testing.T) {
	t.Parallel()

	req, err := NewRequest("exchange", "rk", nil)
	require.NoError(t, err)

	resp, err := NewResponse([]byte("body"))
	require.NoError(t, err)

	t.Run("active expectation", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp)
		require.NoError(t, err)

		assert.True(t, exp.IsActive())
	})

	t.Run("used expectation", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp)
		require.NoError(t, err)

		exp.Use()
		assert.False(t, exp.IsActive())
	})

	t.Run("expired expectation", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithTimeToLive(time.Millisecond*50))
		require.NoError(t, err)

		assert.True(t, exp.IsActive())
		time.Sleep(time.Millisecond * 51)
		assert.False(t, exp.IsActive())
	})

	t.Run("unlimited expectation", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp, WithUnlimitedTimes())
		require.NoError(t, err)

		assert.True(t, exp.IsActive())
		exp.Use()
		exp.Use()
		exp.Use()
		assert.True(t, exp.IsActive())
	})
}

func TestExpectation_Matches(t *testing.T) {
	t.Parallel()

	cmp, err := comparators.NewRegex("foo")
	require.NoError(t, err)

	req, err := NewRequest("exchange", "rk", cmp)
	require.NoError(t, err)

	resp, err := NewResponse([]byte("body"))
	require.NoError(t, err)

	t.Run("matches", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp)
		require.NoError(t, err)

		cnd, err := NewCandidate("exchange", "rk", []byte("something_foo_something"))
		require.NoError(t, err)

		assert.True(t, exp.Matches(cnd))
	})

	t.Run("does not match", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp)
		require.NoError(t, err)

		cnd, err := NewCandidate("exchange", "rk", []byte("something_something"))
		require.NoError(t, err)

		assert.False(t, exp.Matches(cnd))
	})

	t.Run("does not match inactive", func(t *testing.T) {
		t.Parallel()

		exp, err := NewExpectation(req, resp)
		require.NoError(t, err)

		cnd, err := NewCandidate("exchange", "rk", []byte("something_foo_something"))
		require.NoError(t, err)

		assert.True(t, exp.Matches(cnd))
		exp.Use()
		assert.False(t, exp.Matches(cnd))
	})
}

func TestExpectation_Copy(t *testing.T) {
	t.Parallel()

	cmp, err := comparators.NewRegex("foo")
	require.NoError(t, err)

	req, err := NewRequest("exchange", "rk", cmp)
	require.NoError(t, err)

	resp, err := NewResponse([]byte("body"))
	require.NoError(t, err)

	exp, err := NewExpectation(req, resp, WithLimitedTimes(2), WithTimeToLive(time.Second))
	require.NoError(t, err)

	cpy := exp.Copy()

	assert.Equal(t, exp, cpy)
}
