FROM alpine:latest
WORKDIR /test
COPY configs/config.client.local.yaml /test/config.yaml
CMD ["ls", "-l", "/test"]