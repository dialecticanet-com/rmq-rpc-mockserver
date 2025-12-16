package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/app"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/config"
	gocoreamqp "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/amqp"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rmqURL          = "amqp://guest:guest@localhost:5672/"
	rmqCon          *gocoreamqp.Connection
	rmqChannel      *amqp.Channel
	testExchange    = fmt.Sprintf("te-%s", uuid.NewString())
	serviceHostHTTP = ""
	serviceHostGRPC = ""
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

	wg := runService(ctx)

	// run tests
	status := m.Run()

	// cleanup
	cnl()
	wg.Wait()

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

func runService(ctx context.Context) *sync.WaitGroup {
	ports, err := GetFreePorts(2)
	checkErr(err)

	// initiate start of mockserver
	cfg := &config.Config{
		LogLevelStr:                      "debug",
		HTTPPortStr:                      strconv.Itoa(ports[0]),
		GRPCPortStr:                      strconv.Itoa(ports[1]),
		RabbitMQURL:                      rmqURL,
		AMQPQueuesStr:                    "",
		RabbitMQConnectionTimeoutSeconds: 10,
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := app.Run(ctx, cfg)
		if err != nil {
			fmt.Println("error running testing mockserver instance:" + err.Error())
		}
	}()

	serviceHostHTTP = "http://localhost:" + cfg.HTTPPortStr
	serviceHostGRPC = "localhost:" + cfg.GRPCPortStr

	// wait for service to start
	attempts := 0
	for {
		attempts++
		if attempts > 10 {
			fmt.Println("failed to start service")
			os.Exit(1)
		}
		res, err := http.Get(serviceHostHTTP + "/api/v1/version")
		if res != nil {
			_ = res.Body.Close()
		}
		if err == nil && res != nil && res.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return &wg
}
