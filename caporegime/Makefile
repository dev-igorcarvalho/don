MODULE  := github.com/dev-igorcarvalho/don/caporegime
BINARY  := caporegime
BIN_DIR := bin
COVER_FILE := coverage.out

.PHONY: all
all: test build

.PHONY: build
build:
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/caporegime

.PHONY: run
run:
	go run ./cmd/caporegime

.PHONY: test
test:
	go test ./... -race -count=1

.PHONY: test-verbose
test-verbose:
	go test ./... -race -count=1 -v

.PHONY: cover
cover:
	go test ./... -race -count=1 -coverprofile=$(COVER_FILE)
	go tool cover -func=$(COVER_FILE)

.PHONY: cover-html
cover-html: cover
	go tool cover -html=$(COVER_FILE)

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -l -w .
	goimports -l -w .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: check
check: fmt vet lint test

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(COVER_FILE)
