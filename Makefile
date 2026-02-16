.PHONY: test lint build

test:
	go vet ./...
	go test -v -race -count=1 ./...

lint:
	golangci-lint run

build:
	go build -o /dev/null .
