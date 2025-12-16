package amqp

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rmqURL string

func TestMain(m *testing.M) {
	host := "localhost"
	if os.Getenv("RABBITMQ_HOST") != "" {
		host = os.Getenv("RABBITMQ_HOST")
	}
	rmqURL = fmt.Sprintf("amqp://guest:guest@%s:5672/", host)

	con, err := amqp.Dial(rmqURL)
	checkErr(err)

	ch, err := con.Channel()
	checkErr(err)
	_, err = ch.QueueDeclare("test", true, false, false, false, nil)
	checkErr(err)
	_, err = ch.QueueDeclare("test-rpc", true, false, false, false, nil)
	checkErr(err)
	err = ch.ExchangeDeclare("test", "direct", true, false, false, false, nil)
	checkErr(err)
	err = ch.QueueBind("test", "trk", "test", false, nil)
	checkErr(err)
	err = ch.QueueBind("test-rpc", "rpc", "test", false, nil)
	checkErr(err)

	err = con.Close()
	checkErr(err)

	os.Exit(m.Run())
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func TestEstablishConnection(t *testing.T) {
	t.Run("simple start", func(t *testing.T) {
		ctx, cnl := context.WithCancel(context.Background())
		defer cnl()

		con, err := EstablishConnection(ctx, rmqURL, ConnectionWithDialConfig(&amqp.Config{Properties: map[string]interface{}{"test": "val"}}))
		require.NoError(t, err)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := con.Run(ctx)
			assert.NoError(t, err)
			wg.Done()
		}()

		cnl()
		wg.Wait()
	})

	t.Run("connection shutdown", func(t *testing.T) {
		ctx, cnl := context.WithCancel(context.Background())
		defer cnl()

		con, err := EstablishConnection(ctx, rmqURL, ConnectionWithDialConfig(&amqp.Config{Properties: map[string]interface{}{"test": "val"}}))
		require.NoError(t, err)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := con.Run(ctx)
			assert.NoError(t, err)
			wg.Done()
		}()

		err = con.Connection().Close()
		require.NoError(t, err)
		wg.Wait()
	})

	t.Run("connection error", func(t *testing.T) {
		ctx, cnl := context.WithCancel(context.Background())
		defer cnl()

		con, err := EstablishConnection(ctx, "amqp://guest:guest@localhost:1234/")
		require.ErrorContains(t, err, "connection refused")
		require.Nil(t, con)
	})
}

func TestPublishConsume(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	con, err := EstablishConnection(ctx, rmqURL)
	require.NoError(t, err)

	sub, err := NewConsumer(con, "test")
	require.NoError(t, err)

	consumed := atomic.Bool{}
	consumed.Store(false)
	sub.Subscribe("test", "trk", func(_ context.Context, delivery *amqp.Delivery) *HandlerError {
		assert.Equal(t, "application/json", delivery.ContentType)
		assert.JSONEq(t, `{"test":"val"}`, string(delivery.Body))
		consumed.Store(true)
		return nil
	})

	// run the consumer
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		err := sub.Run(ctx)
		assert.NoError(t, err)
		wg.Done()
	}()
	// run the graceful connection
	go func() {
		err := con.Run(ctx)
		assert.NoError(t, err)
		wg.Done()
	}()

	pub, err := NewPublisher(con)
	require.NoError(t, err)
	err = pub.Publish(ctx, "test", "trk", amqp.Publishing{
		Headers:     nil,
		ContentType: "application/json",
		Body:        []byte(`{"test":"val"}`),
	})
	require.NoError(t, err)

	// wait for message to be consumed
	assert.Eventually(t, consumed.Load, time.Second, 10*time.Millisecond, "message not consumed")

	// wait for the consumer to stop
	cnl()
	wg.Wait()
}

func TestConsumer_Error(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	con, err := EstablishConnection(ctx, rmqURL)
	require.NoError(t, err)

	sub, err := NewConsumer(con, "test")
	require.NoError(t, err)

	consumed := atomic.Bool{}
	consumed.Store(false)
	sub.Subscribe("test", "trk", func(_ context.Context, _ *amqp.Delivery) *HandlerError {
		consumed.Store(true)
		return NewHandlerError(fmt.Errorf("test error"), false)
	})

	// run the consumer
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		err := sub.Run(ctx)
		assert.NoError(t, err)
		wg.Done()
	}()
	// run the graceful connection
	go func() {
		err := con.Run(ctx)
		assert.NoError(t, err)
		wg.Done()
	}()

	pub, err := NewPublisher(con)
	require.NoError(t, err)
	err = pub.Publish(ctx, "test", "trk", amqp.Publishing{
		Headers:     nil,
		ContentType: "application/json",
		Body:        []byte(`{"test":"val"}`),
	})
	require.NoError(t, err)

	// wait for message to be consumed
	assert.Eventually(t, consumed.Load, time.Second, 10*time.Millisecond, "message not consumed")

	// wait for the consumer to stop
	cnl()
	wg.Wait()
}

func TestRPC(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	con, err := EstablishConnection(ctx, rmqURL)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(2)
	// start the echo RPC server
	go func() {
		err := startEchoTestRPCServer(ctx, con.Connection(), "test-rpc")
		assert.NoError(t, err)
		wg.Done()
	}()
	// run the graceful connection
	go func() {
		err := con.Run(ctx)
		assert.NoError(t, err)
		wg.Done()
	}()

	// wait for the server to start
	time.Sleep(300 * time.Millisecond)

	client, err := NewRPCClient(con)
	require.NoError(t, err)

	resp, err := client.Call(ctx, "test", "rpc", []byte(`{"test":"val"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"test":"val"}`, string(resp))

	cnl()
	wg.Wait()
}

// startEchoTestRPCServer starts a server that listens for RPC requests on the given queue.
// It is a fake server for testing purposes which starts a span for each request and stops on context cancellation.
// It will reply to each request with the same message.
func startEchoTestRPCServer(ctx context.Context, conn *amqp.Connection, queue string) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery := <-deliveries:
			msg := amqp.Publishing{ContentType: "application/json", CorrelationId: delivery.CorrelationId, Body: delivery.Body}
			if err := ch.Publish("", delivery.ReplyTo, false, false, msg); err != nil {
				return err
			}

			if err := delivery.Ack(false); err != nil {
				return err
			}
		}
	}
}
