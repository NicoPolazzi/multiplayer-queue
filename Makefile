BINARY_NAME=mqueue
CMD_DIR=./cmd
# Exclude certain folders from tests.
TESTFOLDER := $(shell go list ./... | grep -v '/cmd$$' | grep -v '/internal/models$$' | grep -v '/gen/')


.DEFAULT_GOAL := help

help: ## ✨ Show this help message
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
# Print the help message by extracting comments from the Makefile with 
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


fmt: ## 🎨 Format all Go files
	@echo "--> Formatting code..."
	go fmt ./...

lint: ## 🧐 Run the linter
	@echo "--> Running linter..."
	golangci-lint run

check-quality: fmt lint ## ✅ Check code formatting and linting

test: ## 🧪 Run all tests with coverage
	@echo "--> Running tests..."
	go test -race -covermode=atomic -coverprofile=coverage.txt $(TESTFOLDER)

deps: ## 📦 Install and tidy dependencies
	@echo "--> Tidying go modules..."
	go mod tidy

build: deps ## 🔨 Build the application binary
	@echo "--> Building binary..."
	go build -o $(BINARY_NAME) $(CMD_DIR)

run: build ## 🚀 Build and run the executable
	@echo "--> Running executable..."
	./$(BINARY_NAME)

all: check-quality test build ## ✨ Run all checks, tests, and build the binary

clean: ## 🧹 Clean up build artifacts
	@echo "--> Cleaning up..."
	rm -f $(BINARY_NAME) coverage.txt


.PHONY: help run test fmt lint check-quality build all clean