package config

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/dialecticanet-com/rmq-rpc-mockserver/lib/config"
)

// Config holds the configuration for the application.
type Config struct {
	ServiceInfo                      ServiceInfo
	LogLevelStr                      string `envconfig:"LOG_LEVEL" required:"false" default:"info"`
	HTTPPortStr                      string `envconfig:"HTTP_PORT" required:"false" default:"8080"`
	GRPCPortStr                      string `envconfig:"GRPC_PORT" required:"false" default:"8081"`
	RabbitMQURL                      string `envconfig:"RABBITMQ_URL" required:"true"`
	RabbitMQConnectionTimeoutSeconds int    `envconfig:"RABBITMQ_CONNECTION_TIMEOUT_SECONDS" required:"false" default:"300"`
	AMQPQueuesStr                    string `envconfig:"AMQP_QUEUES" required:"false"`
}

type ServiceInfo struct {
	Service    string
	Version    string
	CommitHash string
	BuildDate  string
}

// NewConfig creates a new Config instance.
func NewConfig(envFile string, si ServiceInfo) (*Config, error) {
	cfg := &Config{ServiceInfo: si}

	opts := []config.Option{config.WithEnvPrefix("EMAILS")}
	if envFile != "" {
		opts = append(opts, config.WithEnvFile(envFile))
	}

	if err := config.Load(cfg, opts...); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LogLevel returns the log level suitable for slog.
func (c *Config) LogLevel() slog.Level {
	switch c.LogLevelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// AMQPQueues returns the list of AMQP queues.
func (c *Config) AMQPQueues() []string {
	if c.AMQPQueuesStr == "" {
		return nil
	}

	queues := strings.Split(c.AMQPQueuesStr, ",")
	for i, queue := range queues {
		queues[i] = strings.TrimSpace(queue)
	}

	return queues
}

func (c *Config) HTTPPort() int {
	port, _ := strconv.Atoi(c.HTTPPortStr)
	return port
}

func (c *Config) GRPCPort() int {
	port, _ := strconv.Atoi(c.GRPCPortStr)
	return port
}

func (c *Config) RabbitMQConnectionTimeout() time.Duration {
	return time.Duration(c.RabbitMQConnectionTimeoutSeconds) * time.Second
}
