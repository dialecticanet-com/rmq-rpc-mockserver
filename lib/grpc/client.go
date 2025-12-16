package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewInsecureClient returns a new insecure gRPC client connection to the target.
func NewInsecureClient(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	options = append(options, opts...)

	return grpc.NewClient(target, options...)
}
