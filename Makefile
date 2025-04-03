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

.PHONY: up down linter tests benchmark

up:
	docker-compose -f configs/docker-compose/docker-compose.local.yaml up -d --build

down:
	docker-compose -f configs/docker-compose/docker-compose.local.yaml down -v

linter:
	docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint run -v

tests:
	go test ./internal/...

benchmark:
	go test -bench=. ./app/internal/policy/mitigator/... -run=^#