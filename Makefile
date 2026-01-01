# GoGoldenHour Makefile

# Application name
APP_NAME := gogoldenhour

# Build directory
BUILD_DIR := build

# Source directories
CMD_DIR := cmd/gogoldenhour

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOMOD := $(GOCMD) mod
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-s -w"

.PHONY: all build clean deps test vet run

# Default target
all: deps build

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build the application
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

# Build for development (with debug symbols)
build-dev:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

# Run the application
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run go vet
vet:
	$(GOVET) ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	$(GOCMD) clean

# Install dependencies for the development environment
install-deps:
	@echo "Installing Qt6 development dependencies..."
	@echo "On Arch Linux: sudo pacman -S qt6-base qt6-webengine qt6-webchannel"
	@echo "On Debian/Ubuntu: sudo apt install qt6-base-dev qt6-webengine-dev"

# Show help
help:
	@echo "GoGoldenHour Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make          - Download deps and build"
	@echo "  make deps     - Download Go module dependencies"
	@echo "  make build    - Build the application"
	@echo "  make build-dev- Build with debug symbols"
	@echo "  make run      - Build and run the application"
	@echo "  make test     - Run tests"
	@echo "  make vet      - Run go vet"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make help     - Show this help"
