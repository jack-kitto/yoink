.PHONY: build clean test install dev help

# Default target
.DEFAULT_GOAL := build

# Build variables
BINARY_NAME=yoink
VERSION?=dev
BUILD_DIR=./bin
GO_FILES=$(shell find . -type f -name '*.go' -not -path './vendor/*')

# Build the binary
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean

# Run tests
test:
	go test -v ./...

# Install to GOPATH/bin
install: build
	go install -ldflags="-X main.version=$(VERSION)" .

# Development build with race detection
dev:
	go build -race -ldflags="-X main.version=$(VERSION)-dev" -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
build-all: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  dev        - Development build with race detection"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  tidy       - Tidy dependencies"
	@echo "  help       - Show this help"
