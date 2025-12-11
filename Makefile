.PHONY: build test test-acceptance test-acceptance-verbose test-smoke clean install lint fmt vet

# Build the binary
build:
	go build -o tile-diff ./cmd/tile-diff

# Run unit tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.txt -o coverage.html

# Run acceptance tests
test-acceptance: build
	@echo "Running acceptance tests..."
	@./test/acceptance/run_acceptance_tests.sh

# Run acceptance tests with verbose output
test-acceptance-verbose: build
	@echo "Running acceptance tests (verbose)..."
	@./test/acceptance/run_acceptance_tests.sh --verbose

# Run all tests (unit + acceptance)
test-all: test test-acceptance

# Run smoke test (quick validation without PIVNET_TOKEN)
test-smoke: build
	@echo "Running smoke test..."
	@./test/acceptance/smoke_test.sh

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
