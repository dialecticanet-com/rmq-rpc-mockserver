package http

import (
	"errors"
	"time"
)

// ServerOption is a functional option for configuring a Server.
type ServerOption func(*serverOptions) error

type serverOptions struct {
	port                int
	readTimeout         time.Duration
	writeTimeout        time.Duration
	shutdownGracePeriod time.Duration
	handlerTimeout      time.Duration
	serverIdleTimeout   time.Duration
}

func defaultServerOptions() *serverOptions {
	return &serverOptions{
		port:                50000,
		readTimeout:         30 * time.Second,
		writeTimeout:        60 * time.Second,
		shutdownGracePeriod: 5 * time.Second,
		handlerTimeout:      59 * time.Second,
		serverIdleTimeout:   240 * time.Second,
	}
}

// ServerWithReadTimeout functional option.
func ServerWithReadTimeout(rt time.Duration) ServerOption {
	return func(opts *serverOptions) error {
		if rt <= 0*time.Second {
			return errors.New("negative or zero read timeout provided")
		}
		opts.readTimeout = rt

		return nil
	}
}

// ServerWithWriteTimeout functional option.
func ServerWithWriteTimeout(wt time.Duration) ServerOption {
	return func(opts *serverOptions) error {
		if wt <= 0*time.Second {
			return errors.New("negative or zero write timeout provided")
		}
		opts.writeTimeout = wt

		return nil
	}
}

// ServerWithHandlerTimeout functional option.
func ServerWithHandlerTimeout(wt time.Duration) ServerOption {
	return func(opts *serverOptions) error {
		if wt <= 0*time.Second {
			return errors.New("negative or zero handler timeout provided")
		}
		opts.handlerTimeout = wt

		return nil
	}
}

// ServerWithShutdownGracePeriod functional option.
func ServerWithShutdownGracePeriod(gp time.Duration) ServerOption {
	return func(opts *serverOptions) error {
		if gp <= 0*time.Second {
			return errors.New("negative or zero shutdown grace period timeout provided")
		}
		opts.shutdownGracePeriod = gp

		return nil
	}
}

// ServerWithPort functional option.
func ServerWithPort(port int) ServerOption {
	return func(opts *serverOptions) error {
		if port <= 0 || port > 65535 {
			return errors.New("invalid HTTP ServerWithPort provided")
		}
		opts.port = port

		return nil
	}
}
