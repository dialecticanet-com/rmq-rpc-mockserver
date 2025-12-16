package grpc

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListen(t *testing.T) {
	// we need any port that is not in use, so let's create a temporary listener
	tmpListener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	// get the port and close temporary listener to free the port
	port := strconv.Itoa(tmpListener.Addr().(*net.TCPAddr).Port)
	_ = tmpListener.Close()

	// create a listener with the given port
	listener, err := Listen(ListenerWithPort(port))
	defer func() { _ = listener.Close() }()
	require.NoError(t, err)

	// listener should listen on the given port
	assert.Equal(t, "[::]:"+port, listener.Addr().String())
}
