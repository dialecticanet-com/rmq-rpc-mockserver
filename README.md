# RMQ RPC MockServer

A MockServer for testing RabbitMQ RPC-based integrations. This service allows you to create expectations for incoming AMQP messages and return predefined responses, making it easy to test systems that interact with RabbitMQ without requiring the actual third-party services.

## Features

- **Expectations Management**: Define expected AMQP requests and their corresponding responses
- **Queue Subscriptions**: Dynamically subscribe to RabbitMQ queues for message interception
- **Assertions History**: Track all incoming messages and their matching status
- **gRPC API**: Full-featured gRPC API with generated Go client code
- **HTTP/REST API**: RESTful HTTP endpoints via gRPC-Gateway
- **Flexible Matching**: Support for exact JSON matching, partial JSON matching, and regex patterns
- **TTL Support**: Set time-to-live for expectations
- **Multiple Response Modes**: Respond once, N times, or unlimited times

## Quick Start

### Prerequisites

- Go 1.24+
- RabbitMQ instance running
- Protocol Buffers compiler (`protoc`) for code generation

### Installation

```bash
# Clone the repository
git clone https://github.com/dialecticanet-com/rmq-rpc-mockserver.git
cd rmq-rpc-mockserver

# Install dependencies
go mod download

# Run RabbitMQ (if needed)
make rmq

# Run the MockServer
make run
```

### Using the gRPC Client

Import the generated client code in your Go project:

```go
import v1 "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
```

See [api/README.md](api/README.md) for detailed client usage examples and API documentation.

## Development

### Running Tests

```bash
# Ensure RabbitMQ is running
make rmq

# In another terminal, run tests
make test
```

### Linting

```bash
make lint
```

### Proto Code Generation

If you modify the proto file, regenerate the Go code:

```bash
# Install proto generation tools (first time only)
make proto-install

# Generate Go code from proto files
make proto-gen

# Clean generated files
make proto-clean
```

## API Documentation

- **gRPC API**: See [api/README.md](api/README.md) for complete API documentation and usage examples
- **Proto Definition**: [api/v1/mockserver.proto](api/v1/mockserver.proto)

## Architecture

The MockServer consists of:

- **AMQP Consumer**: Listens to subscribed RabbitMQ queues
- **Expectations Engine**: Matches incoming requests against defined expectations
- **Assertions Tracker**: Records all requests and their matching results
- **gRPC Server**: Provides management API for expectations and subscriptions
- **HTTP Gateway**: Exposes REST endpoints for the gRPC API

## Project Structure

```
.
├── api/                    # gRPC API definitions and generated code
│   ├── v1/
│   │   ├── mockserver.proto          # Proto definition
│   │   ├── mockserver.pb.go          # Generated protobuf messages
│   │   ├── mockserver_grpc.pb.go     # Generated gRPC client/server
│   │   └── mockserver.pb.gw.go       # Generated HTTP gateway
│   └── README.md           # API documentation
├── app/                    # Application entry points
├── internal/               # Internal packages
│   ├── app/                # Application logic
│   ├── config/             # Configuration
│   ├── domain/             # Domain models and business logic
│   └── infra/              # Infrastructure (AMQP, gRPC)
├── lib/                    # Reusable libraries
├── Makefile                # Build and development tasks
└── go.mod                  # Go module definition
```

## Configuration

Configuration is done via environment variables. See `.env.local` for an example configuration file.

## Makefile Targets

```bash
make help              # Show all available targets
make test              # Run tests
make lint              # Run linter
make run               # Run the application
make rmq               # Run RabbitMQ container
make proto-install     # Install proto generation tools
make proto-gen         # Generate code from proto files
make proto-clean       # Remove generated proto files
```

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please ensure all tests pass and code is properly linted before submitting a pull request.

