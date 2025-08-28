TESTFOLDER := $(shell go list ./... | grep -v '/cmd$$' | grep -v '/internal/models$$' | grep -v '/gen/')


check-quality:
	make lint
	make fmt

lint:
	@echo "Running linter..."
	golangci-lint run

fmt:
	@echo "Running formatter..."
	go fmt ./...

test:
	@echo "Running tests..."
	go test -race -covermode atomic -coverprofile=coverage.txt $(TESTFOLDER)

all:
	make check-quality
	make test