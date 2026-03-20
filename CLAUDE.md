# mcp2cli

Remote MCP server tools → local CLI commands.

## Project Structure

- `mcp2cli.go` — Public API (root package)
- `internal/mcp/` — MCP Streamable HTTP client
- `internal/schema/` — JSON Schema → CLI flag conversion
- `internal/auth/` — OAuth 2.1 + PKCE authentication
- `internal/cfgstore/` — Config persistence (~/.config/mcp2cli/)
- `cmd/mcp2cli-runner/` — Generic runner binary
- `docs/intents/` — ADR-style intent records

## Commands

- `make build` — Build runner binary
- `make test` — Run tests
- `make lint` — Run golangci-lint
- `make ci` — Lint + test + build

## Conventions

- Go module: `github.com/xiangma9712/mcp2cli`
- Internal packages are not part of the public API
- MCP protocol version: 2025-03-26 (Streamable HTTP transport)
