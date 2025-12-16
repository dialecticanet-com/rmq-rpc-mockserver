package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	// we need any port that is not in use, so let's create a temporary listener
	tmpListener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	// get the port and close temporary listener to free the port
	port := tmpListener.Addr().(*net.TCPAddr).Port
	_ = tmpListener.Close()

	srv, err := NewServer(ServerWithPort(port))
	require.NoError(t, err)

	srv.Handler = http.NewServeMux()

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

	// call the alive endpoint and check the response
	url := fmt.Sprintf("http://localhost:%d", port)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = res.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	// cancel the context and wait for the server to stop
	cnl()
	wg.Wait()
}
