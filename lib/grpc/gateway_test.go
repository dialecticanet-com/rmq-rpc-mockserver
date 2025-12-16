package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/lib/grpc/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCGateway(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	// create a new grpc server
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()
	grpcServer, err := NewServer(listener)
	require.NoError(t, err)

	// create a gateway connected to the grpc server
	grpcGateway, err := NewGateway(grpcServer)
	require.NoError(t, err)

	// register echo service everywhere
	testSvc := &testEchoService{}
	fixtures.RegisterEchoServiceServer(grpcServer.Server, testSvc)
	err = grpcGateway.RegisterServiceHandlerFromEndpoint(ctx, fixtures.RegisterEchoServiceHandlerFromEndpoint)
	require.NoError(t, err)

	// create an HTTP server connected to the gateway
	httpServer := http.Server{Handler: grpcGateway}
	httpListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer func() { _ = httpListener.Close() }()
	httpHost := fmt.Sprintf("http://localhost:%d", httpListener.Addr().(*net.TCPAddr).Port)

	wg := sync.WaitGroup{}
	// start the grpc server
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Run(ctx)
		assert.NoError(t, err)
	}()

	// start the gateway
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := httpServer.Serve(httpListener)
		assert.ErrorIs(t, err, http.ErrServerClosed)
	}()

	// shut down the server when context is cancelled
	go func() {
		select {
		case <-ctx.Done():
			err := httpServer.Shutdown(ctx)
			assert.NoError(t, err)
		}
	}()

	// give time for servers to start
	time.Sleep(300 * time.Millisecond)

	// prepare request to call the echo service
	client := http.Client{}
	reqBody := []byte(`{"message": "hello"}`)
	req, err := http.NewRequest(http.MethodPost, httpHost+"/echo", bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	// call the echo service using HTTP protocol
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// assert that the response is correct echo message
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"message": "hello"}`, string(resBody))

	// cancel the context and wait for servers to shut down
	cnl()
	wg.Wait()
}

type testEchoService struct {
	fixtures.UnimplementedEchoServiceServer
}

func (s *testEchoService) Echo(_ context.Context, req *fixtures.EchoRequest) (*fixtures.EchoResponse, error) {
	return &fixtures.EchoResponse{Message: req.Message}, nil
}
