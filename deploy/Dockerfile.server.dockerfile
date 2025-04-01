# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project source
COPY . .

# Build the server application
# Make sure the path to main.go is correct relative to WORKDIR
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./app-server/app/cmd/server

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

# Copy the executable from the builder stage
COPY --from=builder /app/server /app/server

# Copy configuration file(s)
# Ensure the config path matches the default or argument in CMD
COPY configs/config.server.local.yaml /app/configs/config.server.local.yaml

# Expose the port the server listens on (must match config.server.local.yaml)
EXPOSE 8081

# Command to run the server
# The config path here matches the default in main.go
CMD ["/app/server", "-config=/app/configs/config.server.local.yaml"]