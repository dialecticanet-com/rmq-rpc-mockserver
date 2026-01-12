# Architecture

This document describes the architecture and design of the RMQ RPC MockServer.

## Table of Contents

- [Overview](#overview)
- [Components](#components)
- [Request Flow](#request-flow)
- [Matching Logic](#matching-logic)

## Overview

The RMQ RPC MockServer is designed to mock third-party services that communicate via RabbitMQ RPC. 
It provides a flexible and powerful system for defining expectations, matching incoming requests, 
and returning configured responses.

## Package Organization

The codebase follows clean architecture principles:

- **`cmd/`**: Entry points for executables
- **`app/`**: Application-level setup and initialization
- **`internal/`**: Private application code
    - **`app/`**: Application use cases and orchestration
    - **`domain/`**: Business logic and domain models
    - **`infra/`**: Infrastructure adapters (AMQP, gRPC)
    - **`config/`**: Configuration management
- **`lib/`**: Reusable library code
- **`api/`**: API contracts and generated code
- **`test/`**: Integration and end-to-end tests

## Components

### Domain Layer

Contains core business logic with no external dependencies.

**Expectations**: Encapsulates the business logic for request matching and response generation. 
Defines how incoming requests are evaluated against expectations, manages expectation lifecycle (TTL and usage counts), 
and determines which response to return based on priority and matching rules.

**Comparators**: Implements matching strategies for request validation. 
Provides JSON body comparison (exact and partial matching) and regex-based pattern matching. 
These are the building blocks used by Expectations to determine if a request matches.

**Subscriptions**: Defines the business rules for queue subscription management. 
Determines what queues the mockserver should listen to and manages the lifecycle of subscriptions.

### Application Layer

Orchestrates use cases and coordinates between domain and infrastructure layers.

**Expectations Service**: Manages the complete lifecycle of expectations. 
Stores active and expired expectations, applies priority ordering when multiple expectations match, 
maintains history of all matched and unmatched requests for debugging, 
and provides query capabilities for assertions and verification.

**Subscriptions Service**: Handles subscription operations. 
Manages active subscriptions to RabbitMQ queues and coordinates with the infrastructure layer to start/stop 
listeners dynamically.

### Infrastructure Layer

Adapters that connect the application to external systems.

**AMQP**: Handles all RabbitMQ interactions. Listens to subscribed queues, receives incoming RPC requests, 
delegates to the application layer for processing, and sends responses back to clients. 
Manages connection pooling, channel lifecycle, and error recovery.

**gRPC**: Provides the management API interface. Exposes all control plane operations
(expectations CRUD, subscriptions management, assertions querying) via gRPC protocol.
Includes HTTP/REST gateway through gRPC-Gateway for JSON-based clients.

### Shared Libraries

**lib**: Reusable components that can be used across projects. 
Includes AMQP client utilities (RPC client, publisher, consumer), 
gRPC server and gateway helpers, HTTP server utilities, configuration management, 
and component lifecycle runners. These libraries have no dependencies on the mockserver's business logic.

## Request Flow

The MockServer operates on two separate flows: the **Data Plane** (handling RPC requests) and the **Control Plane** (managing expectations and subscriptions).

### Data Plane: RPC Request Flow

This is the main flow for incoming RabbitMQ RPC requests:

```
┌──────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                     │
│  ┌─────────────┐                            ┌──────────────┐ │
│  │  RabbitMQ   │──RPC Request──────────────>│ AMQP Listener│ │
│  │   Queues    │                            └──────┬───────┘ │
│  └─────────────┘                                   │         │
└────────────────────────────────────────────────────┼─────────┘
                                                     │
                    ┌────────────────────────────────┘
                    │ Extract: exchange, routing_key, body
                    v
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
│                   ┌──────────────────┐                      │
│                   │  Expectations    │                      │
│                   │     Service      │                      │
│                   └────────┬─────────┘                      │
│                            │                                │
│                            │ Find matching expectations     │
│                            │ Record assertion               │
│                            v                                │
└─────────────────────────────────────────────────────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                v                         v
┌─────────────────────────────────────────────────────────────┐
│                       Domain Layer                          │
│  ┌──────────────────┐           ┌────────────────┐          │
│  │   Expectations   │           │  Comparators   │          │
│  │                  │  uses     │  - JSON Match  │          │
│  │  Match & Select  │─────────> │  - Regex Match │          │
│  │  Manage Lifetime │           └────────────────┘          │
│  │  Generate Response│                                      │
│  └────────┬─────────┘                                       │
└───────────┼─────────────────────────────────────────────────┘
            │
            │ Response body + metadata
            v
┌─────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                    │
│  ┌──────────────┐                         ┌─────────────┐   │
│  │ AMQP Listener│─────RPC Response───────>│  RabbitMQ   │   │
│  └──────────────┘                         │   Queues    │   │
│                                           └─────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

**Flow breakdown:**

1. **Infrastructure → Application**: AMQP listener receives message from subscribed queue and forwards request details (exchange, routing key, body) to Expectations Service

2. **Application → Domain**: Expectations Service asks the Domain layer to find a matching expectation using the configured comparators

3. **Domain Processing**: 
   - Expectations component matches request against all active expectations using Comparators
   - Applies priority-based selection if multiple matches exist
   - Manages expectation lifetime (decrements usage count, checks TTL)
   - Generates response based on the matched expectation

4. **Domain → Application**: Returns matched expectation and response data to Expectations Service

5. **Application Processing**: Records assertion linking the request to the matched expectation (or records unmatched request)

6. **Application → Infrastructure**: Returns response to AMQP listener

7. **Infrastructure → Client**: AMQP listener sends response back via RabbitMQ reply-to queue

### Control Plane: Management API

Configuration and queries flow through a separate path:

```
┌─────────────┐     ┌──────────────┐     ┌──────────────────┐
│  HTTP/REST  │────>│  gRPC        │────>│  Application     │
│  Client     │     │  Gateway     │     │  Services        │
└─────────────┘     └──────────────┘     └──────────────────┘
       or                                          │
┌─────────────┐                                    │
│  gRPC       │────────────────────────────────────┘
│  Client     │
└─────────────┘
```

The Management API allows clients to:
- Create/update/delete expectations
- Manage queue subscriptions
- Query assertions for verification
- Retrieve service information

These operations modify the state used by the Data Plane but don't directly handle RPC requests.

## Matching Logic

### Exchange and Routing Key Matching

All expectations must specify an exchange and routing key. These are matched exactly against incoming messages.

### Body Matching Strategies

#### 1. JSON Body Matching

**Exact Match** (`MATCH_TYPE_EXACT`):
- The entire request body must match the expectation body exactly
- All fields must be present and have the same values
- No additional fields are allowed

**Partial Match** (`MATCH_TYPE_PARTIAL`):
- The request body must contain at least the fields specified in the expectation
- Field values must match
- Additional fields in the request are ignored

Example:

```
Expectation body: {"userId": 123, "action": "create"}
Match type: PARTIAL

Request 1: {"userId": 123, "action": "create"}
Result: ✅ MATCH

Request 2: {"userId": 123, "action": "create", "timestamp": "2025-01-01"}
Result: ✅ MATCH (extra field ignored)

Request 3: {"userId": 123, "action": "delete"}
Result: ❌ NO MATCH (action differs)

Request 4: {"userId": 123}
Result: ❌ NO MATCH (action missing)
```

#### 2. Regex Body Matching

- The request body (as a string) is matched against a regular expression
- Useful for flexible pattern matching
- Can match any text-based format (JSON, XML, plain text, etc.)

Example:

```
Expectation regex: ".*userId.*:\\s*123.*"

Request 1: {"userId": 123, "action": "create"}
Result: ✅ MATCH

Request 2: {"action": "create", "userId": 123}
Result: ✅ MATCH

Request 3: {"userId": 456}
Result: ❌ NO MATCH
```

### Priority-based Matching

When multiple expectations match a request:

1. All matching expectations are collected
2. Expectations are sorted by priority (highest to lowest)
3. The highest priority expectation is selected
4. If priorities are equal, the first created expectation wins

This allows for:
- **Specific overrides**: High-priority expectations for specific cases
- **Default handlers**: Low-priority expectations for general cases

Example:

```
Expectation 1: userId=999, priority=100, response="VIP user"
Expectation 2: action=create, priority=0, response="Standard user"

Request: {"userId": 999, "action": "create"}
Result: Matches both, but Expectation 1 wins (priority 100 > 0)
```

### Lifetime Management

Expectations can be configured with two types of lifetime constraints:

#### 1. Usage-based Lifetime (`times`)

- **remaining_times**: Expectation is valid for N matches
- **unlimited**: Expectation is valid indefinitely

Each time an expectation is matched:
- `remaining_times` is decremented
- When `remaining_times` reaches 0, the expectation is marked as expired
- Expired expectations are removed from matching

#### 2. Time-based Lifetime (`time_to_live_seconds`)

- Expectation expires after a specified number of seconds
- Checked on each matching attempt
- Expired expectations are automatically removed

**Both constraints can be used together**: The expectation expires when either condition is met.
