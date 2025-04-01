# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOMODTIDY=$(GOCMD) mod tidy
BINARY_DIR=./bin
SERVER_BINARY_NAME=faraway-server
CLIENT_BINARY_NAME=faraway-client
SERVER_BINARY=$(BINARY_DIR)/$(SERVER_BINARY_NAME)
CLIENT_BINARY=$(BINARY_DIR)/$(CLIENT_BINARY_NAME)
SERVER_MAIN=./app-server/app/cmd/server/main.go
CLIENT_MAIN=./app-client/app/cmd/client/main.go
SERVER_CONFIG=./configs/config.server.local.yaml
CLIENT_CONFIG=./configs/config.client.local.yaml

# Docker parameters
DOCKER_CMD=docker
DOCKER_COMPOSE_CMD=docker-compose
SERVER_IMAGE_NAME=faraway-server
CLIENT_IMAGE_NAME=faraway-client
DOCKER_NETWORK_NAME=faraway-net

.PHONY: all build build-server build-client run-server run-client clean tidy docker-build docker-build-server docker-build-client docker-up docker-down docker-logs docker-network

all: build

# Build Commands
build: tidy build-server build-client
	@echo "Build complete."

build-server:
	@echo "Building server..."
	mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(SERVER_BINARY) $(SERVER_MAIN)

build-client:
	@echo "Building client..."
	mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(CLIENT_BINARY) $(CLIENT_MAIN)

# Run Commands (Locally)
run-server: build-server
	@echo "Running server locally..."
	$(SERVER_BINARY) -config=$(SERVER_CONFIG)

run-client: build-client
	@echo "Running client locally..."
	# Give server a moment to start if run concurrently
	sleep 1
	$(CLIENT_BINARY) -config=$(CLIENT_CONFIG)

# Dependency Management
tidy:
	@echo "Running go mod tidy..."
	$(GOMODTIDY)

# Docker Commands
docker-build: docker-build-server docker-build-client
	@echo "Docker images built."

docker-build-server:
	@echo "Building server Docker image..."
	$(DOCKER_CMD) build -t $(SERVER_IMAGE_NAME):latest -f Dockerfile.server .

docker-build-client:
	@echo "Building client Docker image..."
	$(DOCKER_CMD) build -t $(CLIENT_IMAGE_NAME):latest -f Dockerfile.client .

# Docker Compose Commands (Requires docker-compose.yml)
docker-network:
	@echo "Creating Docker network $(DOCKER_NETWORK_NAME)..."
	-$(DOCKER_CMD) network create $(DOCKER_NETWORK_NAME) || true # Create network if it doesn't exist

# Start services using Docker Compose
docker-up: docker-network docker-build
	@echo "Starting services with Docker Compose..."
	$(DOCKER_COMPOSE_CMD) up -d # -d runs in detached mode

# Stop services started with Docker Compose
docker-down:
	@echo "Stopping services with Docker Compose..."
	$(DOCKER_COMPOSE_CMD) down

# View logs of services running via Docker Compose
docker-logs:
	@echo "Tailing logs from Docker Compose..."
	$(DOCKER_COMPOSE_CMD) logs -f

# Clean Commands
clean:
	@echo "Cleaning up..."
	rm -rf $(BINARY_DIR)
	$(GOCLEAN)
	# Optional: Remove docker images/containers if needed (use with caution)
	# -$(DOCKER_CMD) rm $$(docker ps -a -q --filter ancestor=$(SERVER_IMAGE_NAME):latest)
	# -$(DOCKER_CMD) rm $$(docker ps -a -q --filter ancestor=$(CLIENT_IMAGE_NAME):latest)
	# -$(DOCKER_CMD) rmi $(SERVER_IMAGE_NAME):latest $(CLIENT_IMAGE_NAME):latest