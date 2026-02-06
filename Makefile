.PHONY: help build run test clean docker-build docker-run migrate lint

# Default target
help:
	@echo "MCP Context Engine - Available Commands:"
	@echo ""
	@echo "  make build              - Build the Go binary"
	@echo "  make run                - Run the server"
	@echo "  make test               - Run all tests"
	@echo "  make test-go            - Run Go tests"
	@echo "  make test-node          - Run Node.js tests"
	@echo "  make test-integration   - Run all integration tests"
	@echo "  make test-e2e           - Run end-to-end tests"
	@echo "  make test-performance   - Run performance tests"
	@echo "  make test-security      - Run security tests"
	@echo "  make setup-test-db      - Setup test database"
	@echo "  make clean-test-db      - Clean test database"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Run with Docker Compose"
	@echo "  make migrate            - Run database migrations"
	@echo "  make lint               - Run linters"
	@echo "  make fmt                - Format code"
	@echo ""

# Build the Go binary
build:
	@echo "Building Go binary..."
	go build -o bin/server ./cmd/server

# Run the server
run:
	@echo "Starting server..."
	go run ./cmd/server/main.go

# Run all tests
test: test-go test-node

# Run Go tests
test-go:
	@echo "Running Go tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run Node.js tests
test-node:
	@echo "Running Node.js tests..."
	npm test

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	./scripts/run-integration-tests.sh

# Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	go test -v -timeout 5m ./test/e2e_integration_test.go

# Run performance tests
test-performance:
	@echo "Running performance tests..."
	go test -v -timeout 10m ./test/performance_test.go

# Run security tests
test-security:
	@echo "Running security tests..."
	go test -v -timeout 5m ./test/security_test.go

# Setup test database
setup-test-db:
	@echo "Setting up test database..."
	createdb contextkeeper_test || true
	psql contextkeeper_test < schema.sql || true

# Clean test database
clean-test-db:
	@echo "Cleaning test database..."
	dropdb contextkeeper_test || true

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf dist/
	rm -rf coverage.out
	rm -f server
	rm -f contextkeeper-backend

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker-compose build

docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

docker-clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

# Run database migrations
migrate:
	@echo "Running database migrations..."
	go run ./cmd/server/main.go migrate

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...
	npm run lint

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	npm run format

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	npm install

# Development mode with hot reload
dev:
	@echo "Starting development mode..."
	air

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060
