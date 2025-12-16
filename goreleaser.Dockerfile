FROM public.ecr.aws/docker/library/alpine:3.20

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/rmq-rpc-mockserver /usr/local/bin/rmq-rpc-mockserver

ENTRYPOINT ["/usr/local/bin/rmq-rpc-mockserver"]
