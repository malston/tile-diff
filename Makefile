.PHONY: build test acceptance-test acceptance-test-with-token acceptance-test-fast acceptance-test-fast-with-token clean install lint fmt vet

# Build the binary
build:
	go build -o tile-diff ./cmd/tile-diff

# Run unit tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./pkg/...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.txt -o coverage.html

# Run Ginkgo acceptance tests (requires PIVNET_TOKEN env var)
acceptance-test: build
	@if [ -z "$$PIVNET_TOKEN" ]; then \
		echo "Error: PIVNET_TOKEN environment variable is required"; \
		echo "Export it or use 'make acceptance-test-with-token'"; \
		exit 1; \
	fi
	TILE_DIFF_BIN="$(PWD)/tile-diff" ginkgo -v ./test

# Run Ginkgo acceptance tests with PIVNET_TOKEN from command line
# Usage: make acceptance-test-with-token PIVNET_TOKEN=your-token-here
acceptance-test-with-token: build
	@if [ -z "$(PIVNET_TOKEN)" ]; then \
		echo "Error: PIVNET_TOKEN is required"; \
		echo "Usage: make acceptance-test-with-token PIVNET_TOKEN=your-token-here"; \
		exit 1; \
	fi
	TILE_DIFF_BIN="$(PWD)/tile-diff" PIVNET_TOKEN=$(PIVNET_TOKEN) ginkgo -v ./test

# Run fast acceptance tests only (skip slow download tests)
# Usage: make acceptance-test-fast PIVNET_TOKEN=your-token-here
acceptance-test-fast: build
	@if [ -z "$$PIVNET_TOKEN" ]; then \
		echo "Error: PIVNET_TOKEN environment variable is required"; \
		echo "Export it or use 'make acceptance-test-fast-with-token'"; \
		exit 1; \
	fi
	TILE_DIFF_BIN="$(PWD)/tile-diff" ginkgo -v --label-filter='!slow' ./test

# Run fast acceptance tests with PIVNET_TOKEN from command line
# Usage: make acceptance-test-fast-with-token PIVNET_TOKEN=your-token-here
acceptance-test-fast-with-token: build
	@if [ -z "$(PIVNET_TOKEN)" ]; then \
		echo "Error: PIVNET_TOKEN is required"; \
		echo "Usage: make acceptance-test-fast-with-token PIVNET_TOKEN=your-token-here"; \
		exit 1; \
	fi
	TILE_DIFF_BIN="$(PWD)/tile-diff" PIVNET_TOKEN=$(PIVNET_TOKEN) ginkgo -v --label-filter='!slow' ./test

# Run all tests (unit + acceptance)
test-all: test acceptance-test

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
