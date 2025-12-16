#!/usr/bin/make

default: help

ci:
	@make lint
	@make test

## lint: Run linter
lint:
	golangci-lint run

## test: Run tests
test:
	@echo "NOTE: Tests require a running RabbitMQ instance on default port (5672) with default credentials (guest:guest)"
	go test -v -race -count=1 ./...

## rmq: Run RabbitMQ container
rmq:
	docker run -p 15672:15672 -p 5672:5672 rabbitmq:management

## test-cover: Run tests with coverage
test-cover:
	go test -v -race -count=1 -coverprofile=coverage.txt ./...

## run: Runs the application
run:
	go run ./cmd/mockserver/main.go --env=.env.local

## proto-install: Install proto generation tools
proto-install:
	@echo "Installing proto generation tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "Proto tools installed successfully!"

## proto-gen: Generate Go code from proto files
proto-gen:
	@echo "Generating Go code from proto files..."
	@PATH="$(shell go env GOPATH)/bin:$$PATH" protoc \
		--proto_path=api/v1 \
		--proto_path=third_party/googleapis \
		--go_out=api/v1 \
		--go_opt=paths=source_relative \
		--go-grpc_out=api/v1 \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=api/v1 \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		api/v1/*.proto
	@echo "Proto generation completed successfully!"

## proto-clean: Remove generated proto files
proto-clean:
	@echo "Cleaning generated proto files..."
	@rm -f api/v1/*.pb.go
	@rm -f api/v1/*.pb.gw.go
	@echo "Cleaned successfully!"

## help: Show this help
help: Makefile
	@echo
	@echo "Available targets:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /' | LANG=C sort
	@echo

.PHONY: default lint test help ci run proto-install proto-gen proto-clean
