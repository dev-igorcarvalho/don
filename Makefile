.PHONY: lint lint-fix test build run

GOLANGCI_LINT_VERSION := v1.61.0

lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint >/dev/null; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	golangci-lint run ./...

lint-incremental:
	@echo "Running incremental linter..."
	golangci-lint run --new-from-rev=HEAD ./...

lint-fix:
	@echo "Fixing lint issues..."
	golangci-lint run --fix ./...

test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building api..."
	go build -o bin/api ./cmd/api

run:
	@echo "Running api..."
	go run cmd/api/main.go
