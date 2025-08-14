check-quality:
	make lint
	make fmt

lint:
	golangci-lint run

fmt:
	go fmt ./...

test:
	go test -race -covermode atomic -coverprofile=covprofile ./... 

all:
	make check-quality
	make test