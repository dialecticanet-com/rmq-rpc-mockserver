package amqp

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	gocoreamqp "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/amqp"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	rmqURL       = "amqp://guest:guest@localhost:5672/"
	rmqCon       *gocoreamqp.Connection
	rmqChannel   *amqp.Channel
	testExchange = fmt.Sprintf("te-%s", uuid.NewString())
)

func TestMain(m *testing.M) {
	ctx, cnl := context.WithTimeout(context.Background(), 10*time.Second)

	host := "localhost"
	if os.Getenv("RABBITMQ_HOST") != "" {
		host = os.Getenv("RABBITMQ_HOST")
	}
	rmqURL = fmt.Sprintf("amqp://guest:guest@%s:5672/", host)

	var err error
	rmqCon, err = gocoreamqp.EstablishConnection(ctx, rmqURL)
	checkErr(err)

	rmqChannel, err = rmqCon.Connection().Channel()
	checkErr(err)

	// tests will use one single exchange and create queues and bindings as needed
	err = rmqChannel.ExchangeDeclare(testExchange, "direct", true, false, false, false, nil)
	checkErr(err)

	status := m.Run()

	cnl()
	if err := rmqChannel.Close(); err != nil {
		fmt.Println("Failed to close rabbitMQ channel:", err.Error())
	}
	if err := rmqCon.Connection().Close(); err != nil {
		fmt.Println("Failed to close rabbitMQ connection:", err.Error())
	}

	os.Exit(status)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func TestConsumer(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	cns, err := NewConsumer(rmqCon, &testMatcher{t: t})
	require.NoError(t, err)

	go func() {
		err := cns.Run(ctx)
		assert.NoError(t, err)
	}()

	q1, rk1 := createRandomQueue(t)
	q2, rk2 := createRandomQueue(t)

	sub1 := subscriptions.NewSubscription(q1)
	err = cns.Subscribe(sub1)
	require.NoError(t, err)

	sub2 := subscriptions.NewSubscription(q2)
	err = cns.Subscribe(sub2)
	require.NoError(t, err)

	// create another subscription for the same queue
	sub3 := subscriptions.NewSubscription(q2)
	err = cns.Subscribe(sub3)
	require.NoError(t, err)

	subs := cns.GetAllSubscriptions()
	assert.Len(t, subs, 3)

	q2Subs := cns.GetQueueSubscriptions(q2)
	assert.Len(t, q2Subs, 2)

	rpcClient, err := gocoreamqp.NewRPCClient(rmqCon)
	require.NoError(t, err)

	resp1, err := rpcClient.Call(ctx, testExchange, rk1, []byte("foo"))
	require.NoError(t, err)
	assert.Equal(t, "foo_bar", string(resp1))

	resp2, err := rpcClient.Call(ctx, testExchange, rk2, []byte("baz"))
	require.NoError(t, err)
	assert.Equal(t, "baz_bar", string(resp2))

	err = cns.Unsubscribe(sub1.ID())
	require.NoError(t, err)

	err = cns.UnsubscribeFromQueue(q2)
	require.NoError(t, err)

	// check that there are no more listeners
	// we can't send a message and check that it is not arrived, since the call just times out
	assert.Empty(t, cns.listeners)

	// graceful shutdown
	cnl()
	time.Sleep(100 * time.Millisecond)
}

type testMatcher struct {
	t      *testing.T
	called atomic.Uint32
}

func (m *testMatcher) Match(candidate *expectations.Candidate) *expectations.Response {
	m.t.Log("matching candidate", "body", string(candidate.Body), "rk", candidate.RoutingKey, "exchange", candidate.Exchange)
	m.called.Add(1)
	return &expectations.Response{Body: []byte(string(candidate.Body) + "_bar")}
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
