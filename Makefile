.PHONY: help build build-all build-windows build-linux build-linux-amd64 build-linux-arm64 \
         build-macos build-macos-amd64 build-macos-arm64 run dev test clean lint fmt \
         docker-build docker-run deploy test-unit test-integration coverage fmt-check vet deps vendor

# Variables
BINARY_NAME=neighborhood-api
MAIN_PATH=./cmd/api
BUILD_DIR=./bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
GOFMT=gofmt
GOVET=$(GOCMD) vet

# LDFLAGS
LDFLAGS=-ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

## help: Display this help message
help:
	@echo "Neighborhood API - Available commands:"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'

## build: Build the application for current OS
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for all platforms (Windows, Linux, macOS)
build-all: build-windows build-linux build-macos
	@echo "✅ All builds complete!"

## build-windows: Build for Windows (AMD64)
build-windows:
	@echo "Building for Windows (AMD64)..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(MAIN_PATH)
	@echo "✓ Windows build complete"

## build-linux: Build for Linux (AMD64 and ARM64)
build-linux: build-linux-amd64 build-linux-arm64

## build-linux-amd64: Build for Linux (AMD64)
build-linux-amd64:
	@echo "Building for Linux (AMD64)..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(MAIN_PATH)
	@echo "✓ Linux (AMD64) build complete"

## build-linux-arm64: Build for Linux (ARM64)
build-linux-arm64:
	@echo "Building for Linux (ARM64)..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 $(MAIN_PATH)
	@echo "✓ Linux (ARM64) build complete"

## build-macos: Build for macOS (Intel and Apple Silicon)
build-macos: build-macos-amd64 build-macos-arm64

## build-macos-amd64: Build for macOS (Intel)
build-macos-amd64:
	@echo "Building for macOS (AMD64/Intel)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(MAIN_PATH)
	@echo "✓ macOS (Intel) build complete"

## build-macos-arm64: Build for macOS (Apple Silicon)
build-macos-arm64:
	@echo "Building for macOS (ARM64/Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 $(MAIN_PATH)
	@echo "✓ macOS (Apple Silicon) build complete"

## run: Run the application
run:
	$(GOCMD) run $(MAIN_PATH)

## dev: Run with live reload (requires air)
dev:
	air

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

## deps: Download dependencies
deps:
	$(GOGET) -v ./...
	$(GOCMD) mod tidy

## test: Run tests
test:
	$(GOTEST) -v ./...
