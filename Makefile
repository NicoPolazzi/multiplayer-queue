TESTFOLDER := $(shell go list ./... | grep -v '/cmd$$' | grep -v '/internal/models$$')


check-quality:
	make lint
	make fmt

lint:
	golangci-lint run

fmt:
	go fmt ./...

test:
	go test -race -covermode atomic -coverprofile=covprofile $(TESTFOLDER)

all:
	make check-quality
	make test