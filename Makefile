.PHONY: build test lint fmt ci clean

BIN_DIR := bin
RUNNER_BIN := $(BIN_DIR)/mcp2cli-runner
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/xiangma9712/mcp2cli.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(RUNNER_BIN) ./cmd/mcp2cli-runner

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
