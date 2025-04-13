.PHONY: build test tidy

build:
	go build ./...

test:
	go test ./... -v

tidy:
	go mod tidy

doc:
	go doc ./...

vet:
	go vet ./...

lint:
	golangci-lint run

ci: tidy vet test lint build
