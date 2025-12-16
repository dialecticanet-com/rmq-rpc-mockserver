package test

import (
	"fmt"
	"net"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewHTTPExpect(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  serviceHostHTTP,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

}

// createRandomQueue creates a random queue and binds it to the test exchange
// and returns the queue name and routing key.
func createRandomQueue(t *testing.T) (queue string, routingKey string) {
	t.Helper()

	queue = fmt.Sprintf("tq-%s", uuid.NewString())
	routingKey = fmt.Sprintf("rk-%s", uuid.NewString())

	_, err := rmqChannel.QueueDeclare(queue, true, false, false, false, nil)
	require.NoError(t, err)

	err = rmqChannel.QueueBind(queue, routingKey, testExchange, false, nil)
	require.NoError(t, err)

	return
}

func newJSONBodyAsStruct(t *testing.T, s string) *structpb.Struct {
	t.Helper()
	st := &structpb.Struct{}
	err := st.UnmarshalJSON([]byte(s))
	require.NoError(t, err)
	return st
}

func newJSONBodyAsValue(t *testing.T, s string) *structpb.Value {
	t.Helper()
	st := newJSONBodyAsStruct(t, s)
	return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: st}}
}

// GetFreePorts asks the kernel for a set of free ports that is ready to use.
func GetFreePorts(num int) ([]int, error) {
	var ports []int
	closers := make([]func() error, 0, num)

	for range num {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		if lst, err := net.ListenTCP("tcp", addr); err == nil {
			// we need to close the listener to release the port
			// so a caller can use it
			closers = append(closers, lst.Close)
			ports = append(ports, lst.Addr().(*net.TCPAddr).Port)
		}
	}

	// close all listeners
	for _, closer := range closers {
		if err := closer(); err != nil {
			fmt.Println("error closing listener while obtaining free port:", err)
		}
	}

	return ports, nil
}
