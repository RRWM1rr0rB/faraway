# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project source
COPY . .

# Build the client application
# Make sure the path to main.go is correct relative to WORKDIR
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/client ./app-client/app/cmd/client

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

# Copy the executable from the builder stage
COPY --from=builder /app/client /app/client

# Copy configuration file(s)
# Ensure the config path matches the default or argument in CMD
COPY configs/config.client.local.yaml /app/configs/config.client.local.yaml

# Command to run the client
# The config path here matches the default in main.go
# Note: When running with Docker Compose, the server address in the config
# might need to be the service name (e.g., "server:8081")
CMD ["/app/client", "-config=/app/configs/config.client.local.yaml"]