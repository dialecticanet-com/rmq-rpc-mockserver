// Package http provides an HTTP server, routing and request handling implementation.
package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// Server is an HTTP server.
type Server struct {
	options *serverOptions
	*http.Server
}

// NewServer creates a new HTTP server with the provided options.
func NewServer(opts ...ServerOption) (*Server, error) {
	o := defaultServerOptions()
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	srv := &Server{
		options: o,
		Server: &http.Server{
			Addr:         fmt.Sprintf(":%d", o.port),
			ReadTimeout:  o.readTimeout,
			WriteTimeout: o.writeTimeout,
			IdleTimeout:  o.serverIdleTimeout,
		},
	}

	return srv, nil
}

// Run starts the HTTP server and listens for incoming requests.
// It blocks until the context is canceled and performs a graceful shutdown.
func (s *Server) Run(ctx context.Context) error {
	go s.waitForShutdown(ctx)
	slog.InfoContext(ctx, "HTTP server is listening", "port", s.options.port)

	err := s.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) waitForShutdown(ctx context.Context) {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.options.shutdownGracePeriod)
	defer cancel()

	// nolint:contextcheck
	if err := s.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(ctx, "HTTP server shutdown failed", "error", err.Error())
	}

	slog.InfoContext(ctx, "HTTP server stopped")
}
