# Makefile for selfsign-path-tool

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=selfsign-path-tool
BINARY_LINUX=$(BINARY_NAME)-linux-amd64
BINARY_WINDOWS=$(BINARY_NAME)-windows-amd64.exe

# Build flags
LDFLAGS=-ldflags="-s -w"

# Default target
.PHONY: all
all: clean test build

# Build for current platform
.PHONY: build
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-windows

# Build for Linux
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_LINUX)

# Build for Windows
.PHONY: build-windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS)

# Build for Windows with GUI support (includes icon)
.PHONY: build-windows-gui
build-windows-gui:
	@echo "Building Windows GUI version with icon support..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS)

# Test
.PHONY: test
test:
	$(GOTEST) -v ./...

# Clean
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX)
	rm -f $(BINARY_WINDOWS)

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run
.PHONY: run
run:
	$(GOBUILD) -o $(BINARY_NAME) && ./$(BINARY_NAME)

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, test, and build for current platform"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all platforms (Linux, Windows)"
	@echo "  build-linux  - Build for Linux x86_64"
	@echo "  build-windows- Build for Windows x86_64"
	@echo "  build-windows-gui - Build for Windows x86_64 with GUI and icon support"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  run          - Build and run the application"
	@echo "  help         - Show this help message"