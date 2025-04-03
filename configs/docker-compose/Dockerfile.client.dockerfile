# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ../app-client/go.mod .
COPY ../app-client/go.sum .

RUN go mod download

COPY ../app-client/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/client ./app/cmd/client/main.go

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/client /app/client
COPY ../configs/config.client.local.yaml /app/configs/

CMD ["/app/client", "-config=/app/configs/config.client.local.yaml"]