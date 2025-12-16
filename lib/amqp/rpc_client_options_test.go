package amqp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRPCClientWithTimeout(t *testing.T) {
	timeout := 5 * time.Second
	o := &rpcClientOptions{}
	err := RPCClientWithTimeout(timeout)(o)
	assert.NoError(t, err)
	assert.Equal(t, timeout, o.timeout)
}

func TestRPCCallWithMandatory(t *testing.T) {
	mandatory := true
	o := &rpcCallOptions{}
	err := RPCCallWithMandatory(mandatory)(o)
	assert.NoError(t, err)
	assert.Equal(t, mandatory, o.mandatory)
}

func TestRPCCallWithImmediate(t *testing.T) {
	immediate := true
	o := &rpcCallOptions{}
	err := RPCCallWithImmediate(immediate)(o)
	assert.NoError(t, err)
	assert.Equal(t, immediate, o.immediate)
}

func TestRPCCallWithBearerToken(t *testing.T) {
	token := "Bearer foo"
	o := &rpcCallOptions{}
	err := RPCCallWithBearerToken(token)(o)
	assert.NoError(t, err)
	assert.Equal(t, token, o.bearerToken)
}

func TestRPCCallWithBearerToken_NoBearer(t *testing.T) {
	token := "foo"
	o := &rpcCallOptions{}
	err := RPCCallWithBearerToken(token)(o)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer "+token, o.bearerToken)
}

func TestRPCCallWithContentType(t *testing.T) {
	contentType := "application/xml"
	o := &rpcCallOptions{}
	err := RPCCallWithContentType(contentType)(o)
	assert.NoError(t, err)
	assert.Equal(t, contentType, o.contentType)
}
