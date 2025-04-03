# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ../app-server/go.mod .
COPY ../app-server/go.sum .

RUN go mod download

COPY ../app-server/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./app/cmd/server/main.go

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server /app/server
COPY ../configs/config.server.local.yaml /app/configs/

CMD ["/app/server", "-config=/app/configs/config.server.local.yaml"]