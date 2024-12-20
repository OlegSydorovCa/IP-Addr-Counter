BINARY_NAME=ip_calc

BUILD_DIR=build

GENERATOR_DIR=generator

GENERATOR_BINARY=$(BUILD_DIR)/generator

LDFLAGS=-ldflags "-s -w"

.PHONY: all build build-linux lint run test clean generator help

all: build

build:
	@echo "==> Building the binary for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)
	@echo "==> Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux:
	@echo "==> Building the binary for Linux/amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux
	@echo "==> Linux build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux"

run: build
	@echo "==> Running the project..."
	@$(BUILD_DIR)/$(BINARY_NAME)

generator-build:
	@echo "==> Building the generator..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(GENERATOR_BINARY) $(GENERATOR_DIR)/main.go
	@echo "==> Generator built: $(GENERATOR_BINARY)"

generator: generator-build
	@echo "==> Running the generator..."
	@$(GENERATOR_BINARY)

lint:
	@echo "==> Running linter..."
	@golangci-lint run
	@echo "==> Linting complete."

test:
	@echo "==> Running tests..."
	@go test ./... -v
	@echo "==> Tests complete."

clean:
	@echo "==> Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "==> Cleanup complete."

help:
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          Build the project (default target)"
	@echo "  build        Build the binary for the current system"
	@echo "  build-linux  Build the binary for Linux/amd64"
	@echo "  run          Build and run the project"
	@echo "  generator    Build and run the test file generator"
	@echo "  lint         Run the linter (requires golangci-lint)"
	@echo "  test         Run all tests"
	@echo "  clean        Remove build artifacts"
	@echo "  help         Show this help message"
