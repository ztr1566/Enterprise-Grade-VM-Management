# Web-Based VM Management Platform - Backend Makefile

.PHONY: build run test clean seed benchmark

# Default build target
build:
	mkdir -p bin
	go build -o bin/server cmd/server/main.go

# Run specific provisioning integration tests
test-provision:
	go test -v ./tests/integration/provisioning_test.go

# Run the server with race detection
run:
	go run cmd/server/main.go

# Run all tests
test:
	go test -v ./...

# Run benchmarks for performance validation (SC-001)
benchmark:
	go test -bench=. -benchmem ./internal/api/handlers/...

# Seed the database with initial admin user
seed:
	go run cmd/seed/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f vm-manager.db

# Install dependencies
deps:
	go mod download
	go mod tidy
