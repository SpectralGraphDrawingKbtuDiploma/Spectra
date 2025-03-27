# Define variables
COMPOSE_FILE = docker-compose.yaml
GO          = go
GOTEST      = $(GO) test
GOBUILD     = $(GO) build
GOMOD       = $(GO) mod
GOCLEAN     = $(GO) clean

.PHONY: build up down restart ps logs clean test tidy help

# Build all services
build:
	docker-compose -f $(COMPOSE_FILE) build

# Start all services in detached mode
up:
	docker-compose -f $(COMPOSE_FILE) up --build -d

# Stop and remove all containers, networks
down:
	docker-compose -f $(COMPOSE_FILE) down --remove-orphans -v

# Restart all services
restart: down up

# Show running containers
ps:
	docker-compose -f $(COMPOSE_FILE) ps

# Show logs of all services
logs:
	docker-compose -f $(COMPOSE_FILE) logs -f

# View logs of a specific service
logs-%:
	docker-compose -f $(COMPOSE_FILE) logs -f $*

# Run tests for mtx_server
test:
	cd mtx_server && $(GOTEST) ./...

# Update go.mod and download dependencies
tidy:
	cd mtx_server && $(GOMOD) tidy && $(GOMOD) download

# Build mtx_server locally for development
build-local:
	cd mtx_server && $(GOBUILD) -o bin/server ./cmd/server

# Clean up docker resources
clean: down
	docker-compose -f $(COMPOSE_FILE) down -v
	docker system prune -f
	cd mtx_server && $(GOCLEAN)
	rm -rf mtx_server/bin/*

# Show help information
help:
	@echo "Available commands:"
	@echo "  make build           - Build all services"
	@echo "  make up              - Start all services"
	@echo "  make down            - Stop and remove all services"
	@echo "  make restart         - Restart all services"
	@echo "  make ps              - Show running containers"
	@echo "  make logs            - Show logs of all services"
	@echo "  make logs-SERVICE    - Show logs of a specific service (e.g., make logs-mtx_server)"
	@echo "  make test            - Run tests for mtx_server"
	@echo "  make tidy            - Update go.mod and download dependencies"
	@echo "  make build-local     - Build mtx_server locally for development"
	@echo "  make clean           - Clean up all Docker resources"
	@echo "  make help            - Show this help information"
