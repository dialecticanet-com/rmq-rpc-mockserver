package app

import (
	"net/http"
	"strconv"

	gocoregrpc "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/grpc"
	gocorehttp "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/http"
)

type infraServer struct {
	grpcServer  *gocoregrpc.Server
	httpServer  *gocorehttp.Server
	grpcGateway *gocoregrpc.Gateway
}

func newInfraServer(httpPort, grpcPort int) (*infraServer, error) {
	grpcListener, err := gocoregrpc.Listen(gocoregrpc.ListenerWithPort(strconv.Itoa(grpcPort)))
	if err != nil {
		return nil, err
	}

	grpcServer, err := gocoregrpc.NewServer(grpcListener)
	if err != nil {
		return nil, err
	}

	grpcGateway, err := gocoregrpc.NewGateway(grpcServer)
	if err != nil {
		return nil, err
	}

	httpServer, err := gocorehttp.NewServer(gocorehttp.ServerWithPort(httpPort))
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcGateway.ServeMux)
	httpServer.Handler = mux

	return &infraServer{
		grpcServer:  grpcServer,
		httpServer:  httpServer,
		grpcGateway: grpcGateway,
	}, nil
}
