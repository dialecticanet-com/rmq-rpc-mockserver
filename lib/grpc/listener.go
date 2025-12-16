package grpc

import "net"

const (
	defaultGrpcPort = "50051"
)

// Listener is a wrapper around net.Listener that allows for easy creation of gRPC listeners.
// it supports setting the port number with default value.
type Listener struct {
	net.Listener
	port string
}

// Listen creates a new TPC listener with provided options.
func Listen(opts ...ListenerOption) (*Listener, error) {
	l := &Listener{
		port: defaultGrpcPort,
	}

	for _, opt := range opts {
		opt(l)
	}

	var err error
	l.Listener, err = net.Listen("tcp", ":"+l.port)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// ListenerOption is a function that sets options for the listener.
type ListenerOption func(*Listener)

// ListenerWithPort sets the port number for the listener.
func ListenerWithPort(port string) ListenerOption {
	return func(l *Listener) {
		l.port = port
	}
}
