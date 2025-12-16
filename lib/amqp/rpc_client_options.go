package amqp

import (
	"strings"
	"time"
)

type rpcClientOptions struct {
	timeout time.Duration
}

func defaultRPCClientOptions() *rpcClientOptions {
	return &rpcClientOptions{
		timeout: 5 * time.Second,
	}
}

// RPCClientOptionFunc is a function that configures an RPC client.
type RPCClientOptionFunc func(*rpcClientOptions) error

// RPCClientWithTimeout sets the timeout for the RPC client.
func RPCClientWithTimeout(timeout time.Duration) RPCClientOptionFunc {
	return func(o *rpcClientOptions) error {
		o.timeout = timeout
		return nil
	}
}

type rpcCallOptions struct {
	contentType string
	mandatory   bool
	immediate   bool
	bearerToken string
}

func defaultRPCCallOptions() *rpcCallOptions {
	return &rpcCallOptions{
		contentType: "application/json",
		mandatory:   false,
		immediate:   false,
		bearerToken: "",
	}
}

// RPCCallOptionFunc is a function that configures an RPC call.
type RPCCallOptionFunc func(*rpcCallOptions) error

// RPCCallWithMandatory sets the mandatory flag for the RPC call.
func RPCCallWithMandatory(mandatory bool) RPCCallOptionFunc {
	return func(o *rpcCallOptions) error {
		o.mandatory = mandatory
		return nil
	}
}

// RPCCallWithImmediate sets the immediate flag for the RPC call.
func RPCCallWithImmediate(immediate bool) RPCCallOptionFunc {
	return func(o *rpcCallOptions) error {
		o.immediate = immediate
		return nil
	}
}

// RPCCallWithBearerToken sets the bearer token for the RPC call.
func RPCCallWithBearerToken(token string) RPCCallOptionFunc {
	return func(o *rpcCallOptions) error {
		// Ensure the token is prefixed with "Bearer ".
		if !strings.HasPrefix(token, "Bearer ") {
			token = "Bearer " + token
		}

		o.bearerToken = token
		return nil
	}
}

// RPCCallWithContentType sets the content type for the RPC call.
func RPCCallWithContentType(contentType string) RPCCallOptionFunc {
	return func(o *rpcCallOptions) error {
		o.contentType = contentType
		return nil
	}
}
