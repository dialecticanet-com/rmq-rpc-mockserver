FROM public.ecr.aws/docker/library/golang:1.24 AS builder

ARG APPLICATION_VERSION=dev

WORKDIR /app
COPY . ./
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=$APPLICATION_VERSION -X main.commitHash=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) -s -w" -o rmq-rpc-mockserver ./cmd/rmq-rpc-mockserver/main.go

FROM public.ecr.aws/docker/library/alpine:3.20
COPY --from=builder /app/rmq-rpc-mockserver .
CMD ["./rmq-rpc-mockserver"]
