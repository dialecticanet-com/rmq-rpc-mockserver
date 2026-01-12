# Development Guide

This guide covers development workflows, testing, code generation, and contribution guidelines for the RMQ RPC MockServer.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Linting](#linting)
- [Proto Code Generation](#proto-code-generation)
- [Code Style](#code-style)
- [Contributing](#contributing)

## Prerequisites

### Required

- **Go 1.23+**: [Download Go](https://golang.org/dl/)
- **RabbitMQ**: Running instance for testing
  - Can be started locally with Docker: `make rmq`

### Optional (for code generation)

- **Protocol Buffers Compiler (`protoc`)**: Required if modifying `.proto` files
- **Proto generation tools**: Can be installed with `make proto-install`

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/dialecticanet-com/rmq-rpc-mockserver.git
cd rmq-rpc-mockserver
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Start RabbitMQ (if not already running)

```bash
make rmq
```

This starts a RabbitMQ container with:
- AMQP port: `5672`
- Management UI: `http://localhost:15672` (guest/guest)

### 4. Run the MockServer

```bash
make run
```

The server will start with:
- HTTP API on `http://localhost:8080`
- gRPC API on `localhost:8081`

## Development Workflow

### Typical Development Cycle

1. **Make code changes**
2. **Run linter**: `make lint`
3. **Run tests**: `make test`
4. **Run locally**: `make run`
5. **Test manually** using curl or gRPC client
6. **Commit changes**

## Testing

### Running All Tests

```bash
make test
```

This runs all unit and integration tests.

### Integration Tests

Integration tests are located in the `test/` directory. They require a running RabbitMQ instance:

```bash
# Start RabbitMQ
make rmq

# Run integration tests
go test ./test/...
```

## Linting

The project uses `golangci-lint` for code quality checks.

### Running the Linter

```bash
make lint
```

### Installing golangci-lint

If not already installed:

```bash
# macOS
brew install golangci-lint

# Linux / WSL
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or using Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Proto Code Generation

If you modify the `.proto` files, you need to regenerate the Go code.

### Installing Proto Tools (First Time Only)

```bash
make proto-install
```

This installs:
- `protoc-gen-go` - Generates Go protobuf messages
- `protoc-gen-go-grpc` - Generates Go gRPC client/server code
- `protoc-gen-grpc-gateway` - Generates gRPC-Gateway HTTP bindings
- `protoc-gen-openapiv2` - Generates OpenAPI/Swagger documentation

### Prerequisites: Install protoc

#### macOS

```bash
brew install protobuf
```

#### Linux (Debian/Ubuntu)

```bash
sudo apt-get install protobuf-compiler
```

#### Other Platforms

Download from: https://github.com/protocolbuffers/protobuf/releases

### Regenerating Code

After modifying `api/v1/mockserver.proto`:

```bash
make proto-gen
```

This generates:
- `api/v1/mockserver.pb.go` - Protobuf messages
- `api/v1/mockserver_grpc.pb.go` - gRPC client/server
- `api/v1/mockserver.pb.gw.go` - HTTP gateway

### Cleaning Generated Files

```bash
make proto-clean
```

## Code Style

### General Guidelines

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write descriptive variable and function names
- Add comments for exported functions and types
- Keep functions small and focused
- Handle errors explicitly

## Contributing

We welcome contributions! Please follow these guidelines:

### 1. Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/rmq-rpc-mockserver.git
cd rmq-rpc-mockserver
```

### 2. Create a Branch

```bash
git checkout -b feature/my-new-feature
# or
git checkout -b fix/my-bug-fix
```

### 3. Make Changes

- Write code following the style guidelines
- Add tests for new functionality
- Update documentation as needed

### 4. Run Tests and Linter

```bash
make test
make lint
```

### 5. Commit Changes

Write clear, descriptive commit messages. 
This project uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).


```bash
git commit -m "feat: add support for custom headers in expectations"
```

**Good commit messages**:
- `feat: add regex body matching support`
- `fix: race condition in assertion tracker`
- `docs: update API documentation for new endpoint`

**Bad commit messages**:
- `fix bug`
- `changes`
- `fix: update`

### 6. Push and Create Pull Request

```bash
git push origin feature/my-new-feature
```

Then create a pull request on GitHub with:
- Clear description of changes
- Link to related issues (if any)
- Test results

### Pull Request Checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated (if needed)
- [ ] Commit messages are clear
- [ ] Changes are backwards compatible (or documented)

### Code Review Process

1. Maintainer reviews the PR
2. Address feedback and requested changes
3. Once approved, maintainer merges the PR

## Getting Help

If you need help:

1. Check the [README](../README.md) for general documentation
2. Review the [API documentation](API.md)
3. Check [existing issues](https://github.com/dialecticanet-com/rmq-rpc-mockserver/issues)
4. Open a new issue with:
   - Clear description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
