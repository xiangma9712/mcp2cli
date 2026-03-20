# Contributing to mcp2cli

## Development setup

Prerequisites: [mise](https://mise.jdx.dev/)

```bash
git clone https://github.com/xiangma9712/mcp2cli.git
cd mcp2cli
mise install
```

## Building and testing

```bash
make build    # Build mcp2cli-runner binary to bin/
make test     # Run all tests
make lint     # Run golangci-lint
make fmt      # Format code (gofmt + goimports)
make ci       # Full CI check: lint + test + build
```

## Project structure

```
mcp2cli.go                  # Public API (the only exported package)
internal/mcp/               # MCP Streamable HTTP client
internal/schema/             # JSON Schema → CLI flag conversion
internal/auth/               # OAuth 2.1 + PKCE authentication
internal/cfgstore/           # Config persistence
cmd/mcp2cli-runner/          # Generic runner binary
```

Packages under `internal/` are not part of the public API. Only `mcp2cli.go` (the root package) is exported.

## Conventions

- Zero external dependencies — only the Go standard library
- Run `make ci` before submitting a PR
- Keep test coverage for new code
- Follow existing code style (enforced by `golangci-lint`)

## Pull requests

1. Fork and create a feature branch from `main`
2. Make your changes with tests
3. Run `make ci` to verify
4. Open a PR against `main`

Keep PRs focused — one concern per PR.
