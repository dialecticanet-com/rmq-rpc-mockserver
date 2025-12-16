# RMQ RPC MockServer API

This directory contains the gRPC API definition and generated Go code for the RMQ RPC MockServer.

## Overview

The MockServer provides a gRPC API for managing expectations and subscriptions for RabbitMQ message testing. Users can import the generated client code to interact with the MockServer programmatically.

## Directory Structure

```
api/
├── v1/
│   ├── mockserver.proto        # Proto definition file
│   ├── mockserver.pb.go         # Generated Go protobuf messages
│   ├── mockserver_grpc.pb.go    # Generated Go gRPC client/server code
│   └── mockserver.pb.gw.go      # Generated gRPC-Gateway HTTP/REST bindings
└── README.md                     # This file
```

## Using the Client in Go

### Installation

Import the generated client code in your Go project:

```go
import (
    v1 "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)
```

### Example: Creating an Expectation

```go
package main

import (
    "context"
    "log"
    "time"

    v1 "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/protobuf/types/known/structpb"
)

func main() {
    // Connect to the MockServer
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create a client
    client := v1.NewAmqpMockServerServiceClient(conn)

    // Create an expectation
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Prepare request body
    requestBody, _ := structpb.NewStruct(map[string]interface{}{
        "userId": 123,
        "action": "create",
    })

    // Prepare response body
    responseBody, _ := structpb.NewValue(map[string]interface{}{
        "status": "success",
        "userId": 123,
    })

    resp, err := client.CreateExpectation(ctx, &v1.CreateExpectationRequest{
        Request: &v1.Request{
            Exchange:   "my-exchange",
            RoutingKey: "my.routing.key",
            Body: &v1.Request_JsonBody{
                JsonBody: &v1.JSONBodyAssertion{
                    Body:      requestBody,
                    MatchType: v1.JSONBodyAssertion_MATCH_TYPE_EXACT,
                },
            },
        },
        Response: &v1.Response{
            Body: responseBody,
        },
        Times: &v1.Times{
            Times: &v1.Times_Unlimited{
                Unlimited: true,
            },
        },
    })

    if err != nil {
        log.Fatalf("Failed to create expectation: %v", err)
    }

    log.Printf("Created expectation with ID: %s", resp.ExpectationId)
}
```

### Example: Adding a Subscription

```go
// Add a subscription to a queue
resp, err := client.AddSubscription(ctx, &v1.AddSubscriptionRequest{
    Queue:      "test-queue",
    Idempotent: true,
})

if err != nil {
    log.Fatalf("Failed to add subscription: %v", err)
}

log.Printf("Subscription ID: %s", resp.Subscription.Id)
```

### Example: Getting Assertions

```go
// Get all assertions
resp, err := client.GetAssertions(ctx, &v1.GetAssertionsRequest{
    Status:  proto.String("matched"),
    Include: []string{"expectation"},
})

if err != nil {
    log.Fatalf("Failed to get assertions: %v", err)
}

for _, assertion := range resp.Assertions {
    log.Printf("Assertion ID: %s, Matched: %v", assertion.Id, assertion.Matched)
}
```

### Example: Resetting the MockServer

```go
// Reset all expectations and subscriptions
_, err := client.ResetAll(ctx, &v1.ResetAllRequest{})
if err != nil {
    log.Fatalf("Failed to reset: %v", err)
}

log.Println("MockServer reset successfully")
```

## Available Services

The API provides the following gRPC methods:

### Expectations
- `CreateExpectation` - Create a new expectation for incoming requests
- `GetExpectations` - List all expectations
- `GetExpectation` - Get a specific expectation by ID
- `ResetExpectations` - Remove all expectations

### Subscriptions
- `AddSubscription` - Subscribe to a RabbitMQ queue
- `DeleteSubscription` - Remove a subscription by ID
- `UnsubscribeFromQueue` - Remove all subscriptions for a queue
- `GetAllSubscriptions` - List all active subscriptions
- `ResetSubscriptions` - Remove all subscriptions

### Assertions
- `GetAssertions` - Get history of assertion matches/mismatches

### Utility
- `ResetAll` - Reset both expectations and subscriptions
- `GetVersion` - Get MockServer version information

## HTTP/REST API

The gRPC service also exposes an HTTP/REST API via gRPC-Gateway. All gRPC methods are accessible via HTTP:

- `POST /api/v1/expectations` - Create expectation
- `GET /api/v1/expectations` - List expectations
- `GET /api/v1/expectations/{id}` - Get expectation
- `DELETE /api/v1/expectations` - Reset expectations
- `POST /api/v1/subscriptions` - Add subscription
- `DELETE /api/v1/subscriptions/{id}` - Delete subscription
- `GET /api/v1/subscriptions` - List subscriptions
- `DELETE /api/v1/subscriptions` - Reset subscriptions
- `GET /api/v1/assertions` - Get assertions
- `DELETE /api/v1/reset` - Reset all
- `GET /api/v1/version` - Get version

## Regenerating Code

If you modify the proto file, regenerate the Go code using:

```bash
make proto-gen
```

This will regenerate all `*.pb.go`, `*.pb.gw.go`, and `*_grpc.pb.go` files.

### Prerequisites for Code Generation

1. **Install protoc** (Protocol Buffer Compiler):
   ```bash
   # macOS
   brew install protobuf

   # Linux (Debian/Ubuntu)
   apt-get install protobuf-compiler

   # Or download from: https://github.com/protocolbuffers/protobuf/releases
   ```

2. **Install Go proto plugins**:
   ```bash
   make proto-install
   ```

   This installs:
   - `protoc-gen-go` - Generates Go protobuf messages
   - `protoc-gen-go-grpc` - Generates Go gRPC client/server code
   - `protoc-gen-grpc-gateway` - Generates gRPC-Gateway HTTP bindings
   - `protoc-gen-openapiv2` - Generates OpenAPI/Swagger documentation

3. **Download googleapis protos** (automatically handled):
   The Makefile automatically references the `third_party/googleapis` directory for Google API proto dependencies.

### Makefile Targets

- `make proto-install` - Install all proto generation tools
- `make proto-gen` - Generate Go code from proto files
- `make proto-clean` - Remove all generated files

## Proto Package

The proto package name is `rmqrpc.mockserver.api.v1` and the Go package is:

```
github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1
```

This follows gRPC best practices with a descriptive, version-aligned package name that clearly indicates this is for RabbitMQ RPC mocking, avoiding any potential namespace collisions with other mockserver implementations.

## Dependencies

The generated code depends on:
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers
- `github.com/grpc-ecosystem/grpc-gateway/v2` - gRPC-Gateway for HTTP/REST support

These are already included in the project's `go.mod`.

