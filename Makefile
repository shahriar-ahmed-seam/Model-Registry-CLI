# Model Registry CLI Makefile

.PHONY: build test clean install lint

BINARY_NAME=ml-reg
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")

build:
	@echo "Building ${BINARY_NAME} ${VERSION}"
	@go build -ldflags "-s -w -X main.version=${VERSION}" -o ${BINARY_NAME} .

test:
	@echo "Running tests..."
	@go test ./...

test-verbose:
	@echo "Running verbose tests..."
	@go test -v ./...

integration-test:
	@echo "Running integration tests..."
	@go test ./cmd_test.go

coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	@rm -f ${BINARY_NAME} ${BINARY_NAME}-*
	@rm -f coverage.out coverage.html
	@rm -rf dist/

install:
	@echo "Installing ${BINARY_NAME} to $(GOPATH)/bin"
	@go install -ldflags "-s -w -X main.version=${VERSION}" .

lint:
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...

release:
	@echo "Building release binaries..."
	@goreleaser release --snapshot --clean

all: lint test build

help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run all tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  integration-test - Run integration tests"
	@echo "  coverage       - Generate test coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install to GOPATH/bin"
	@echo "  lint           - Run linters and formatters"
	@echo "  release        - Build release binaries (snapshot)"
	@echo "  all            - Run lint, tests, and build"
	@echo "  help           - Show this help message"