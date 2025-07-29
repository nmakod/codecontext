# CodeContext Makefile for Building and Distribution

VERSION ?= $(shell cat VERSION)
BINARY_NAME = codecontext
BUILD_DIR = dist
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.gitCommit=$(shell git rev-parse --short HEAD)"

# Default target
all: clean build

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build for current platform
build: $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/codecontext

# Build for multiple platforms (CGO-enabled builds for Tree-sitter support)
build-all: $(BUILD_DIR)
	# macOS (current architecture - native build)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-$(shell go env GOARCH) ./cmd/codecontext
	# Note: Cross-compilation with CGO for Tree-sitter requires platform-specific build environments
	# For production releases, use GitHub Actions or platform-specific builders

# Create release tarballs
release: build-all
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-$(shell go env GOARCH).tar.gz $(BINARY_NAME)-darwin-$(shell go env GOARCH)

# Generate checksums for release files
checksums: release
	cd $(BUILD_DIR) && \
	shasum -a 256 *.tar.gz *.zip > checksums.txt

# Install locally (for testing)
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Uninstall
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Prepare for Homebrew (native build only - Homebrew will build from source)
homebrew: $(BUILD_DIR)
	# Build for current platform
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/codecontext
	# Create tarball for Homebrew
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-$(shell go env GOARCH).tar.gz $(BINARY_NAME)
	# Generate checksum
	cd $(BUILD_DIR) && shasum -a 256 $(BINARY_NAME)-$(VERSION)-darwin-$(shell go env GOARCH).tar.gz > $(BINARY_NAME)-$(VERSION)-darwin-$(shell go env GOARCH).tar.gz.sha256

# Development build (with debug symbols)
dev-build: $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev ./cmd/codecontext

# Show help
help:
	@echo "Available targets:"
	@echo "  all         - Clean and build for current platform"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all supported platforms"
	@echo "  release     - Create release tarballs for all platforms"
	@echo "  checksums   - Generate checksums for release files"
	@echo "  homebrew    - Build universal macOS binary for Homebrew"
	@echo "  install     - Install binary locally"
	@echo "  uninstall   - Remove installed binary"
	@echo "  test        - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  clean       - Clean build artifacts"
	@echo "  help        - Show this help"

.PHONY: all clean build build-all release checksums install uninstall test test-coverage fmt lint homebrew dev-build help