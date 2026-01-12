# RMQ RPC MockServer API

This directory contains the gRPC API definition and generated Go code for the RMQ RPC MockServer.

## 📖 Documentation

For complete API documentation, including usage examples, code samples, and endpoint reference, please see:

**[API Documentation](../docs/API.md)**

## Directory Structure

```
api/
├── v1/
│   ├── mockserver.proto        # Proto definition file
│   ├── mockserver.pb.go        # Generated Go protobuf messages
│   ├── mockserver_grpc.pb.go   # Generated Go gRPC client/server code
│   └── mockserver.pb.gw.go     # Generated gRPC-Gateway HTTP/REST bindings
└── README.md                    # This file
```

## Code Generation

If you modify the proto file, regenerate the Go code:

```bash
# Install proto generation tools (first time only)
make proto-install

# Generate Go code from proto files
make proto-gen
```

See the [Development Guide](../docs/DEVELOPMENT.md) for detailed instructions on proto code generation.
