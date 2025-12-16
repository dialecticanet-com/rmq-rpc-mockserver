// Package grpc provides a wrapper around the gRPC server and gRPC Gateway that provides additional functionality.
package grpc

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

// Server is a wrapper around grpc.Server that provides additional functionality.
type Server struct {
	*grpc.Server
	listener net.Listener
}

// NewServer creates a new gRPC server with provided options.
func NewServer(listener net.Listener, opts ...grpc.ServerOption) (*Server, error) {
	grpcOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryServerPanicRecoveryInterceptor()),
	}
	grpcOptions = append(grpcOptions, opts...)

	return &Server{
		Server:   grpc.NewServer(grpcOptions...),
		listener: listener,
	}, nil
}

// Run starts the gRPC server and blocks until it is stopped.
// it is handy for use with the components library which handles the lifecycle of all long-running background processes.
func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.GracefulStop()
		slog.InfoContext(ctx, "gRPC server stopped")
	}()

	slog.InfoContext(ctx, "gRPC server is listening", "address", s.listener.Addr().String())
	return s.Serve(s.listener)
}

// ListeningAddress returns the address the server is listening on.
func (s *Server) ListeningAddress() string {
	return s.listener.Addr().String()
}
