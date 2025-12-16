package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/app"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/config"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/infra/amqp"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/infra/grpc"
	gocoreamqp "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/amqp"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/lib/components"
)

func Run(ctx context.Context, cfg *config.Config) error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       cfg.LogLevel(),
		ReplaceAttr: nil,
	})))

	expectationsSvc := app.NewExpectationsService()

	amqpCon, err := connectToRabbitMQ(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to establish RabbitMQ connection: %w", err)
	}
	defer func() {
		err := amqpCon.Connection().Close()
		if err != nil {
			slog.ErrorContext(ctx, "failed to close RabbitMQ connection", "error", err.Error())
		} else {
			slog.InfoContext(ctx, "RabbitMQ connection closed")
		}
	}()

	amqpConsumer, err := amqp.NewConsumer(amqpCon, expectationsSvc)
	if err != nil {
		return fmt.Errorf("failed to create RabbitMQ consumer: %w", err)
	}

	subscriptionsSvc := app.NewSubscriptionsService(amqpConsumer)
	for _, queue := range cfg.AMQPQueues() {
		if _, err := subscriptionsSvc.Subscribe(queue, false); err != nil {
			return fmt.Errorf("failed to subscribe to queue %s: %w", queue, err)
		}
	}

	infraSrv, err := newInfraServer(cfg.HTTPPort(), cfg.GRPCPort())
	if err != nil {
		return fmt.Errorf("failed to create infrastructure server: %w", err)
	}

	amqpMockserverService := grpc.NewAmqpMockServerServiceServer(expectationsSvc, subscriptionsSvc, &cfg.ServiceInfo)
	grpcApi.RegisterAmqpMockServerServiceServer(infraSrv.grpcServer.Server, amqpMockserverService)
	err = infraSrv.grpcGateway.RegisterServiceHandlerFromEndpoint(ctx, grpcApi.RegisterAmqpMockServerServiceHandlerFromEndpoint)
	if err != nil {
		return fmt.Errorf("failed to register gRPC gateway: %w", err)
	}

	// run all components
	cmp, err := components.NewRunner()
	if err != nil {
		return fmt.Errorf("failed to create components : %w", err)
	}

	return cmp.Run(ctx, infraSrv.grpcServer, infraSrv.httpServer, amqpConsumer)
}

func connectToRabbitMQ(ctx context.Context, cfg *config.Config) (*gocoreamqp.Connection, error) {
	timeout := time.After(cfg.RabbitMQConnectionTimeout())

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("RabbitMQ connection timeout")
		default:
			amqpCon, err := gocoreamqp.EstablishConnection(ctx, cfg.RabbitMQURL)
			if err == nil {
				return amqpCon, nil
			}
			slog.WarnContext(ctx, "failed to establish RabbitMQ connection. retrying in one second...", "error", err.Error())
			time.Sleep(1 * time.Second)
		}
	}
}
