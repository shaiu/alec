# Alec - Script-to-CLI TUI System
# Makefile for build, test, lint, and install targets

.PHONY: build test lint install clean deps fmt vet cover bench

# Build configuration
BINARY_NAME=alec
BUILD_DIR=bin
MAIN_PACKAGE=./cmd/alec

# Version information (will be injected at build time)
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags="-X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.BuildTime=$(BUILD_TIME)'"

# Default target
all: deps fmt vet lint test build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run linter (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Run tests
test:
	go test -v ./...

# Run tests with coverage
cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test -bench=. -benchmem ./tests/performance/...

# Build the application
build: deps
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Install the application to GOPATH/bin
install:
	go install $(LDFLAGS) $(MAIN_PACKAGE)

# Build for multiple platforms
build-all: deps
	mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run development version
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with sample scripts
run-sample: build
	mkdir -p ./example-scripts
	echo '#!/bin/bash\necho "Hello from sample script!"' > ./example-scripts/hello.sh
	chmod +x ./example-scripts/hello.sh
	echo '#!/usr/bin/env python3\nprint("Hello from Python!")' > ./example-scripts/hello.py
	chmod +x ./example-scripts/hello.py
	./$(BUILD_DIR)/$(BINARY_NAME) --script-dirs ./example-scripts

# Development targets
dev-deps:
	# Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# CI targets
ci-test: deps fmt vet lint
	go test -v -race -coverprofile=coverage.out ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  cover       - Run tests with coverage"
	@echo "  bench       - Run benchmarks"
	@echo "  build-all   - Build for all platforms"
	@echo "  run         - Build and run"
	@echo "  run-sample  - Run with sample scripts"
	@echo "  dev-deps    - Install development tools"
	@echo "  ci-test     - Run CI tests"