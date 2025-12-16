package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

// Gateway is a gRPC gateway.
// Since it embeds runtime.ServeMux, it can be used as a http.Handler.
type Gateway struct {
	*runtime.ServeMux
	server *Server
}

// NewGateway creates a new gRPC gateway with provided options.
func NewGateway(server *Server, opts ...runtime.ServeMuxOption) (*Gateway, error) {
	if server == nil {
		return nil, errors.New("server is required")
	}

	muxOpts := []runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithIncomingHeaderMatcher(func(input string) (string, bool) {
			s := strings.ToLower(input)
			// TODO: consider what headers should be allowed.
			if strings.HasPrefix(s, "dialectica") || strings.HasPrefix(s, "x-") {
				return input, true
			}
			return "", false
		}),
	}
	muxOpts = append(muxOpts, opts...)

	return &Gateway{
		ServeMux: runtime.NewServeMux(muxOpts...),
		server:   server,
	}, nil
}

// RegisterServiceHandlerFromEndpoint registers a service handler from an endpoint.
// It is a wrapper around the grpc-gateway functionality and streamlines the process.
//
// example:
//
//	grpcGateway, err := grpc.NewGateway(server)
//	grpcGateway.RegisterServiceHandlerFromEndpoint(context.Background(), v1.RegisterMathServiceHandlerFromEndpoint)
func (g *Gateway) RegisterServiceHandlerFromEndpoint(ctx context.Context,
	registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error,
	opts ...grpc.DialOption) error {
	grpcDialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryGatewayClientInterceptor()),
	}
	grpcDialOpts = append(grpcDialOpts, opts...)

	return registerFunc(ctx, g.ServeMux, g.server.ListeningAddress(), grpcDialOpts)
}

// unaryGatewayClientInterceptor returns a gRPC unary client interceptor
// that injects the current span context into the gRPC metadata.
// It is used to propagate the tracing context to the gRPC server.
func unaryGatewayClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, callOpts...)
	}
}
