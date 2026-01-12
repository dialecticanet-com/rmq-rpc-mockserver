# RabbitMQ RPC MockServer

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23%2B-blue.svg)](https://golang.org/dl/)

A mock server for testing RabbitMQ RPC applications. 
Define expectations for incoming RMQ RPC requests and specify the responses that should be returned, 
making it easy to test and develop messaging applications without depending on actual third-party services.

## Features

- **Dual API Interface**: Manage via gRPC or HTTP JSON APIs
- **Flexible Message Matching**: Exact JSON, partial JSON, and regex-based matching
- **Priority-based Expectations**: Control match order with configurable priorities
- **Lifetime Management**: Set TTL and usage count limits for expectations
- **Assertions Tracking**: Monitor all requests and their matching status
- **Dynamic Queue Management**: Subscribe/unsubscribe from queues at runtime
- **Real-time Logging**: Detailed logs for debugging and monitoring

## Installation

### Using Docker

```bash
docker pull ghcr.io/dialecticanet-com/rmq-rpc-mockserver:latest

docker run -p 8080:8080 -p 8081:8081 \
  -e RABBITMQ_URL=amqp://guest:guest@localhost:5672 \
  ghcr.io/dialecticanet-com/rmq-rpc-mockserver:latest
```

**Note**: If RabbitMQ is running on the host machine, you might need to use `--network host` or `host.docker.internal`.

### Building from Source

```bash
git clone https://github.com/dialecticanet-com/rmq-rpc-mockserver.git
cd rmq-rpc-mockserver
go mod download
make run
```

### Downloading Pre-built Binaries

Download the latest release from the [Releases Page](https://github.com/dialecticanet-com/rmq-rpc-mockserver/releases).

## Quick Start

### 1. Start the MockServer

```bash
docker run -p 8080:8080 -p 8081:8081 \
  -e RABBITMQ_URL=amqp://guest:guest@localhost:5672 \
  ghcr.io/dialecticanet-com/rmq-rpc-mockserver:latest
```

### 2. Subscribe to a Queue

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"queue": "my-service-queue", "idempotent": true}'
```

Keep in mind that the queue must already exist in RabbitMQ. Mockserver will not create it.

### 3. Create an Expectation

```bash
curl -X POST http://localhost:8080/api/v1/expectations \
  -H "Content-Type: application/json" \
  -d '{
    "request": {
      "exchange": "my_exchange",
      "routing_key": "my.routing.key",
      "json_body": {
        "body": {"action": "create", "userId": 123},
        "match_type": "MATCH_TYPE_PARTIAL"
      }
    },
    "response": {
      "body": {"status": "success", "userId": 123}
    },
    "times": {"remaining_times": 5},
    "priority": 10
  }'
```

### 4. Send a Request

Use your RabbitMQ RPC client to send a request to the queue.
The MockServer will match it against expectations and return the configured response.

### 5. Check Assertions

```bash
curl http://localhost:8080/api/v1/assertions?status=matched
```

## Configuration

Configure the MockServer using environment variables:

| Variable                              | Required | Default | Description                                                          |
|---------------------------------------|----------|---------|----------------------------------------------------------------------|
| `RABBITMQ_URL`                        | **Yes**  | N/A     | RabbitMQ connection string (e.g., `amqp://user:pass@localhost:5672`) |
| `RABBITMQ_CONNECTION_TIMEOUT_SECONDS` | No       | `300`   | Initial connection timeout in seconds                                |
| `AMQP_QUEUES`                         | No       | N/A     | Comma-separated list of queues to subscribe to at startup            |
| `LOG_LEVEL`                           | No       | `info`  | Logging level: `debug`, `info`, `warn`, or `error`                   |
| `HTTP_PORT`                           | No       | `8080`  | HTTP API port                                                        |
| `GRPC_PORT`                           | No       | `8081`  | gRPC API port                                                        |

**Example `.env` file**:

```bash
RABBITMQ_URL=amqp://guest:guest@localhost:5672
AMQP_QUEUES=service-queue-1,service-queue-2
LOG_LEVEL=debug
HTTP_PORT=8080
GRPC_PORT=8081
```

## Documentation

- **[API Documentation](docs/API.md)** - Complete API reference for gRPC and HTTP endpoints
- **[Architecture](docs/ARCHITECTURE.md)** - System design, components, and request flow
- **[Development Guide](docs/DEVELOPMENT.md)** - Setup, testing, and contribution guidelines

## API Overview

The MockServer exposes two equivalent APIs:

- **HTTP JSON API**: `http://localhost:8080/api/v1`
- **gRPC API**: `localhost:8081`

### Key Endpoints

| Method | Endpoint              | Description                |
|--------|-----------------------|----------------------------|
| POST   | `/expectations`       | Create a new expectation   |
| GET    | `/expectations`       | List all expectations      |
| GET    | `/expectations/{id}`  | Get a specific expectation |
| DELETE | `/expectations`       | Delete all expectations    |
| POST   | `/subscriptions`      | Subscribe to a queue       |
| GET    | `/subscriptions`      | List all subscriptions     |
| DELETE | `/subscriptions/{id}` | Delete a subscription      |
| GET    | `/assertions`         | Get assertion history      |
| DELETE | `/reset`              | Reset all state            |
| GET    | `/version`            | Get version information    |

For detailed API documentation with examples, see [API Documentation](docs/API.md).

## Contributing

We welcome contributions! Please see our [Development Guide](docs/DEVELOPMENT.md) for details on:

- Setting up your development environment
- Running tests and linters
- Code style guidelines
- Submitting pull requests

**Quick contribution checklist**:

1. Fork the repository and create your branch from `main`
2. Write tests for any new functionality
3. Ensure all tests pass: `make test`
4. Lint your code: `make lint`
5. Update documentation as needed
6. Submit a pull request with a clear description

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
