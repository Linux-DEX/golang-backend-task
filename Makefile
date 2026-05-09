.PHONY: help build run test clean docker-up docker-down k8s-deploy k8s-clean

# Variables
BINARY_NAME=edi-api
WORKER_BINARY=edi-worker

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the API server
	@echo "Building API server..."
	go build -o $(BINARY_NAME) cmd/api/main.go

build-worker: ## Build the worker
	@echo "Building worker..."
	go build -o $(WORKER_BINARY) cmd/worker/main.go

run: ## Run the API server
	@echo "Running API server..."
	go run cmd/api/main.go

run-worker: ## Run the worker
	@echo "Running worker..."
	go run cmd/worker/main.go

test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) $(WORKER_BINARY)
	rm -f coverage.out coverage.html

docker-up: ## Start all services with docker-compose
	@echo "Starting services..."
	docker-compose up -d --build

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down

docker-logs: ## View docker logs
	docker-compose logs -f

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	./k8s/deploy.sh

k8s-clean: ## Clean up Kubernetes deployment
	@echo "Cleaning up Kubernetes..."
	./k8s/cleanup.sh
