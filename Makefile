.PHONY: build test lint fmt ci clean

BIN_DIR := bin
RUNNER_BIN := $(BIN_DIR)/mcp2cli-runner

build:
	go build -o $(RUNNER_BIN) ./cmd/mcp2cli-runner

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

ci: lint test build

clean:
	rm -rf $(BIN_DIR)
