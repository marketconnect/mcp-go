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

cover-html:
	go tool cover -html=coverage.out

sonar:
	go test -coverprofile=coverage.out ./...
	sed -i 's|github.com/marketconnect/mcp-go/||g' coverage.out
	sonar-scanner