package grpc

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/lib/grpc/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCServer(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	srv, err := NewServer(listener)
	require.NoError(t, err)

	echoSrv := &testEchoService{}
	fixtures.RegisterEchoServiceServer(srv.Server, echoSrv)

	ctx, cnl := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := srv.Run(ctx)
		assert.NoError(t, err)
	}()

	// give time for the server to start
	time.Sleep(300 * time.Millisecond)

	// listening address should be the same as the listener
	assert.Equal(t, listener.Addr().String(), srv.ListeningAddress())

	// connect to the server
	con, err := NewInsecureClient(srv.ListeningAddress())
	require.NoError(t, err)

	// call the echo service and check the response
	echoClient := fixtures.NewEchoServiceClient(con)
	res, err := echoClient.Echo(ctx, &fixtures.EchoRequest{Message: "hello"})
	require.NoError(t, err)
	assert.Equal(t, "hello", res.Message)

	// cancel the context and wait for the server to stop
	cnl()
	wg.Wait()
}
