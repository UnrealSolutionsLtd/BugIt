# BugIt Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION)"

.PHONY: all build test clean docker run

all: build

# Build the binary
build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o bugit ./cmd/bugit

# Build for Linux (cross-compile)
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bugit-linux ./cmd/bugit

# Run tests
test:
	go test -v ./...

# Run the server locally
run: build
	./bugit serve --port 8080 --data-dir ./data

# Clean build artifacts
clean:
	rm -f bugit bugit-linux
	rm -rf ./data

# Build Docker image
docker:
	docker build -t bugit:$(VERSION) -t bugit:latest --build-arg VERSION=$(VERSION) .

# Run in Docker
docker-run:
	docker run -p 8080:8080 -v $(PWD)/data:/app/data bugit:latest

# Initialize development data directory
init:
	mkdir -p data/bundles data/tmp

# Create a sample test bundle (for development)
test-bundle:
	@mkdir -p /tmp/test_bundle
	@echo '{"schema_version":"1.0","build_id":"TestGame-1.0.0","platform":"Win64","timestamp":"2026-01-21T10:00:00Z","artifacts":[{"filename":"test.log","type":"log","mime_type":"text/plain"}]}' > /tmp/test_bundle/manifest.json
	@echo "Test log file" > /tmp/test_bundle/test.log
	@cd /tmp/test_bundle && zip -r ../test_bundle.zip .
	@echo "Created /tmp/test_bundle.zip"

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate go.sum
deps:
	go mod tidy

.DEFAULT_GOAL := build
