package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPanicRecoveryInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		res := "test"
		handler := func(ctx context.Context, req any) (any, error) {
			return res, nil
		}

		got, err := unaryServerPanicRecoveryInterceptor()(context.Background(), nil, nil, handler)
		assert.NoError(t, err)
		assert.Equal(t, res, got)
	})

	t.Run("panic", func(t *testing.T) {
		t.Parallel()

		handler := func(ctx context.Context, req any) (any, error) {
			panic("test")
		}

		got, err := unaryServerPanicRecoveryInterceptor()(context.Background(), nil, nil, handler)
		assert.Nil(t, got)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "panic caught", st.Message())
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		handler := func(ctx context.Context, req any) (any, error) {
			return nil, status.Errorf(codes.PermissionDenied, "test")
		}

		got, err := unaryServerPanicRecoveryInterceptor()(context.Background(), nil, nil, handler)
		assert.Nil(t, got)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
		assert.Equal(t, "test", st.Message())
	})
}
