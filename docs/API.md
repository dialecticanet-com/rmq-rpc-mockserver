# API Documentation

This document describes the gRPC and HTTP/REST APIs for the RMQ RPC MockServer.

## Table of Contents

- [Overview](#overview)
- [gRPC API](#grpc-api)
- [HTTP/REST API](#httprest-api)
- [API Reference](#api-reference)

## Overview

The MockServer provides both gRPC and HTTP/REST APIs for managing expectations and subscriptions for RabbitMQ message testing. 
The HTTP API is automatically generated from the gRPC definitions using gRPC-Gateway, 
ensuring consistency between both interfaces.

**Choose your interface:**
- **gRPC**: High-performance, strongly-typed client for Go applications
- **HTTP/REST**: Simple JSON API accessible from any HTTP client (curl, Postman, etc.)

## gRPC API

The MockServer provides a gRPC API defined in Protocol Buffers. 
This allows you to generate strongly-typed clients in your preferred programming language.

### Getting Started

1. **Get the Proto Definition**
   
   The gRPC service is defined in [`api/v1/mockserver.proto`](../api/v1/mockserver.proto).

2. **Generate a Client for Your Language**
   
   Use the protoc compiler with the appropriate language plugin to generate client code:
   
   - **Go**: `protoc --go_out=. --go-grpc_out=. mockserver.proto`
   - **Python**: `python -m grpc_tools.protoc --python_out=. --grpc_python_out=. mockserver.proto`
   - **Java**: `protoc --java_out=. --grpc-java_out=. mockserver.proto`
   - **Node.js**: `grpc_tools_node_protoc --js_out=. --grpc_out=. mockserver.proto`
   - See [gRPC documentation](https://grpc.io/docs/languages/) for other languages

3. **Connect to the MockServer**
   
   Connect your generated client to the MockServer gRPC endpoint (default: `localhost:8081`).

4. **Use the API**
   
   All service methods, request/response types, and field documentation are defined in the proto file.

## HTTP/REST API

All gRPC methods are accessible via HTTP/REST endpoints. The API accepts and returns JSON.

### Base URL

```
http://localhost:8080/api/v1
```

### API Endpoints

| Method | Endpoint                        | Description                              |
|--------|---------------------------------|------------------------------------------|
| POST   | `/expectations`                 | Create a new expectation                 |
| GET    | `/expectations`                 | List all expectations                    |
| GET    | `/expectations/{id}`            | Get a specific expectation               |
| DELETE | `/expectations`                 | Delete all expectations                  |
| POST   | `/subscriptions`                | Add a queue subscription                 |
| GET    | `/subscriptions`                | List all subscriptions                   |
| DELETE | `/subscriptions/{id}`           | Delete a subscription by ID              |
| DELETE | `/subscriptions/queues/{queue}` | Unsubscribe from a queue                 |
| DELETE | `/subscriptions`                | Delete all subscriptions                 |
| GET    | `/assertions`                   | Get assertion history                    |
| DELETE | `/reset`                        | Reset all (expectations + subscriptions) |
| GET    | `/version`                      | Get version information                  |

## API Reference

### Expectations

#### Create Expectation

**POST** `/api/v1/expectations`

Creates a new expectation that defines how the MockServer should respond to matching requests.

**Request Body**:

```json
{
  "request": {
    "exchange": "string",
    "routing_key": "string",
    "json_body": {
      "body": {},
      "match_type": "MATCH_TYPE_EXACT|MATCH_TYPE_PARTIAL"
    }
  },
  "response": {
    "body": {}
  },
  "times": {
    "remaining_times": 1,
    "unlimited": false
  },
  "time_to_live_seconds": 3600,
  "priority": 0
}
```

**Request Fields**:
- `request.exchange` (string, required): The exchange name to match
- `request.routing_key` (string, required): The routing key to match
- `request.json_body` (object, optional): JSON body matching configuration
  - `body` (object): The JSON structure to match
  - `match_type` (string): `MATCH_TYPE_EXACT` or `MATCH_TYPE_PARTIAL`
- `request.regex_body` (object, optional): Alternative to json_body
  - `regex` (string): Regular expression to match against request body
- `response.body` (object, required): The response body to return
- `times` (object, optional): Lifetime based on match count
  - `remaining_times` (int): Number of times to match (default: 1)
  - `unlimited` (bool): Match unlimited times
- `time_to_live_seconds` (int, optional): Lifetime in seconds
- `priority` (int, optional): Priority for matching order (default: 0)

**Response**:

```json
{
  "expectation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Example (Exact JSON Match)**:

```bash
curl -X POST http://localhost:8080/api/v1/expectations \
  -H "Content-Type: application/json" \
  -d '{
    "request": {
      "exchange": "orders_exchange",
      "routing_key": "order.create",
      "json_body": {
        "body": {
          "orderId": "12345",
          "product": "widget"
        },
        "match_type": "MATCH_TYPE_EXACT"
      }
    },
    "response": {
      "body": {
        "status": "created",
        "orderId": "12345"
      }
    },
    "times": {
      "remaining_times": 1
    }
  }'
```

**Example (Partial JSON Match)**:

```bash
curl -X POST http://localhost:8080/api/v1/expectations \
  -H "Content-Type: application/json" \
  -d '{
    "request": {
      "exchange": "users_exchange",
      "routing_key": "user.update",
      "json_body": {
        "body": {
          "userId": 123
        },
        "match_type": "MATCH_TYPE_PARTIAL"
      }
    },
    "response": {
      "body": {
        "success": true
      }
    },
    "times": {
      "unlimited": true
    },
    "priority": 5
  }'
```

**Example (Regex Match)**:

```bash
curl -X POST http://localhost:8080/api/v1/expectations \
  -H "Content-Type: application/json" \
  -d '{
    "request": {
      "exchange": "events_exchange",
      "routing_key": "event.log",
      "regex_body": {
        "regex": ".*timestamp.*:\\s*\"[0-9]{4}-[0-9]{2}-[0-9]{2}.*"
      }
    },
    "response": {
      "body": {
        "logged": true
      }
    },
    "times": {
      "unlimited": true
    }
  }'
```

#### Get Expectations

**GET** `/api/v1/expectations`

Retrieves all expectations, optionally filtered by status.

**Query Parameters**:
- `status` (string, optional): Filter by status (`active` or `expired`)

**Example**:

```bash
# Get all expectations
curl http://localhost:8080/api/v1/expectations

# Get only active expectations
curl "http://localhost:8080/api/v1/expectations?status=active"
```

**Response**:

```json
{
  "expectations": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "request": {
        "exchange": "orders_exchange",
        "routing_key": "order.create",
        "json_body": {
          "body": {
            "orderId": "12345"
          },
          "match_type": "MATCH_TYPE_PARTIAL"
        }
      },
      "response": {
        "body": {
          "status": "created"
        }
      },
      "times": {
        "remaining_times": 3
      },
      "priority": 10,
      "time_to_live_seconds": 3600,
      "created_at": "2025-12-19T10:00:00Z"
    }
  ]
}
```

#### Get Expectation by ID

**GET** `/api/v1/expectations/{id}`

Retrieves a specific expectation by its ID.

**Example**:

```bash
curl http://localhost:8080/api/v1/expectations/550e8400-e29b-41d4-a716-446655440000
```

#### Delete All Expectations

**DELETE** `/api/v1/expectations`

Removes all expectations.

**Example**:

```bash
curl -X DELETE http://localhost:8080/api/v1/expectations
```

### Subscriptions

#### Add Subscription

**POST** `/api/v1/subscriptions`

Subscribes to a RabbitMQ queue to receive messages.

**Request Body**:

```json
{
  "queue": "my-queue-name",
  "idempotent": true
}
```

**Request Fields**:
- `queue` (string, required): Queue name to subscribe to
- `idempotent` (bool, optional): If true, prevents duplicate subscriptions to the same queue

**Response**:

```json
{
  "subscription": {
    "id": "sub-123",
    "queue": "my-queue-name"
  }
}
```

**Example**:

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "orders-queue",
    "idempotent": true
  }'
```

#### Get All Subscriptions

**GET** `/api/v1/subscriptions`

Retrieves all active subscriptions.

**Example**:

```bash
curl http://localhost:8080/api/v1/subscriptions
```

**Response**:

```json
[
  {
    "id": "sub-123",
    "queue": "orders-queue"
  },
  {
    "id": "sub-456",
    "queue": "users-queue"
  }
]
```

#### Delete Subscription by ID

**DELETE** `/api/v1/subscriptions/{id}`

Removes a specific subscription by its ID.

**Example**:

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/sub-123
```

#### Unsubscribe from Queue

**DELETE** `/api/v1/subscriptions/queues/{queue}`

Removes all subscriptions associated with a specific queue.

**Example**:

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/queues/orders-queue
```

#### Delete All Subscriptions

**DELETE** `/api/v1/subscriptions`

Removes all subscriptions.

**Example**:

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions
```

### Assertions

#### Get Assertions

**GET** `/api/v1/assertions`

Retrieves the history of assertion attempts (matched and unmatched requests).

**Query Parameters**:
- `status` (string, optional): Filter by status (`matched` or `unmatched`)
- `expectation_id` (string, optional): Filter by specific expectation ID
- `include` (string, optional): Include additional data (use `expectation` to include full expectation details)

**Example**:

```bash
# Get all assertions
curl http://localhost:8080/api/v1/assertions

# Get only matched assertions
curl "http://localhost:8080/api/v1/assertions?status=matched"

# Get assertions for a specific expectation
curl "http://localhost:8080/api/v1/assertions?expectation_id=550e8400-e29b-41d4-a716-446655440000"

# Include full expectation details
curl "http://localhost:8080/api/v1/assertions?include=expectation"
```

**Response**:

```json
{
  "assertions": [
    {
      "id": "93471bd2-7638-4e61-93b5-007a416eb5fe",
      "candidate": {
        "exchange": "my_exchange",
        "routing_key": "my.routing.key",
        "body": {
          "action": "create",
          "userId": 121
        }
      },
      "matched": false,
      "created_at": "2026-01-12T13:22:49Z"
    },
    {
      "id": "74471d0d-ceeb-46fe-a954-1e3c49bd7717",
      "candidate": {
        "exchange": "my_exchange",
        "routing_key": "my.routing.key",
        "body": {
          "action": "create",
          "userId": 123
        }
      },
      "matched": true,
      "expectation": {
        "id": "9356f568-bd20-4e7b-8c8c-513f6a6d26b6",
        "request": {
          "exchange": "my_exchange",
          "routing_key": "my.routing.key",
          "json_body": {
            "body": {
              "action": "create",
              "userId": 123
            },
            "match_type": "MATCH_TYPE_PARTIAL"
          }
        },
        "response": {
          "body": {
            "status": "success",
            "userId": 123
          }
        },
        "times": {
          "remaining_times": 4
        },
        "created_at": "2026-01-12T13:18:04Z"
      },
      "created_at": "2026-01-12T13:23:00Z"
    }
  ]
}
```

### Utility

#### Reset All

**DELETE** `/api/v1/reset`

Removes all expectations and subscriptions.

**Example**:

```bash
curl -X DELETE http://localhost:8080/api/v1/reset
```

#### Get Version

**GET** `/api/v1/version`

Retrieves version information about the MockServer.

**Example**:

```bash
curl http://localhost:8080/api/v1/version
```

**Response**:

```json
{
  "version": "v1.0.0",
  "commit_hash": "abc123def456",
  "build_date": "2025-12-19T10:00:00Z"
}
```
