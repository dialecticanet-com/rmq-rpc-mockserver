FROM public.ecr.aws/docker/library/alpine:3.20

# GoReleaser (dockers) will place the built binary in the Docker build context
# as rmq-rpc-mockserver, so we just copy it directly.
COPY rmq-rpc-mockserver /usr/local/bin/rmq-rpc-mockserver

ENTRYPOINT ["/usr/local/bin/rmq-rpc-mockserver"]