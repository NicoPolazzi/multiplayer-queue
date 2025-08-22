TESTFOLDER := $(shell go list ./... | grep -v '/cmd$$' | grep -v '/internal/models$$' | grep -v '/gen/')


check-quality:
	make lint
	make fmt

lint:
	golangci-lint run

fmt:
	go fmt ./...

test:
	go test -race -covermode atomic -coverprofile=coverage.txt $(TESTFOLDER)

all:
	make check-quality
	make test