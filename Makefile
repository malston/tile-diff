.PHONY: build test clean install lint fmt vet

# Build the binary
build:
	go build -o tile-diff ./cmd/tile-diff

# Run tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.txt -o coverage.html

# Clean build artifacts
clean:
	rm -f tile-diff coverage.txt coverage.html

# Install the binary
install:
	go install ./cmd/tile-diff

# Run linters
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run all checks
check: fmt vet test

# Download dependencies
deps:
	go mod download
	go mod tidy
