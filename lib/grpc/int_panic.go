package grpc

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// unaryServerPanicRecoveryInterceptor returns a new unary server interceptor that recovers from panics.
// it logs the stack trace and returns an internal error.
// NOTE: this interceptor is internal and added by default to the server.
func unaryServerPanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				slog.ErrorContext(ctx, "panic recovered", "stack", string(stack), "panic", r)
				err = status.Errorf(codes.Internal, "panic caught")
			}
		}()

		return handler(ctx, req)
	}
}
