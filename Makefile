# Document Processing System Makefile

.PHONY: help build test clean docker-build docker-run docker-stop lint format

# Default target
help:
	@echo "Document Processing System - Available commands:"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker Compose services"
	@echo "  lint         - Run linter"
	@echo "  format       - Format code"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t document-processor .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Create test data directory
test-data:
	@echo "Creating test data directory..."
	mkdir -p test-data
	@echo "Add .txt files to test-data/ directory for testing"

# Database operations
db-migrate:
	@echo "Running database migrations..."
	psql -h localhost -U postgres -d document_processor -f migrations/001_initial_schema.sql

# Development server (requires Go installed)
dev:
	@echo "Starting development server..."
	go run cmd/server/main.go

# Health check
health:
	@echo "Checking service health..."
	curl -f http://localhost:8080/health || echo "Service is not healthy" 