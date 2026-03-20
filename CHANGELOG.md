# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-03-20

### Added

- Tools cache (1h TTL) — `--help` serves from cache without server round-trip
- Dynamic Client Registration (RFC 7591) for OAuth servers that require it
- Browser auto-open on `auth login` (macOS/Linux)
- `DEBUG=1` environment variable for debug logging via `internal/debuglog` package
- Sample CLI: `cmd/mcp2cli-notion` for Notion MCP server
- `(JSON)` hint appended to array/object flag descriptions in `--help`
- Timeout documentation in SECURITY.md

### Changed

- `Run()` returns `error` instead of calling `os.Exit()` — callers handle exit
- Help display: flags shown first, then usage example, then full description as "Details"
- Top-level `--help` shows short descriptions (first sentence); `<tool> --help` shows full details
- Error messages include server URL and auth login hint on connection failure
- Tool error output uses actual error content instead of generic message
- Version injected at build time via ldflags (`-X github.com/xiangma9712/mcp2cli.version`)
- SSE scanner buffer increased to 10MB for large tool lists
- `schema/convert.go` warning uses debuglog instead of unconditional `log.Printf`

### Fixed

- Nil pointer dereference when server returns empty result
- Resource leak in OAuth `Login()` — TCP listener now closed on all error paths
- `AuthenticatedHTTPClient` missing timeout (30s added)
- `firstSentence()` panic when `maxLen < 4`

### Removed

- `WithToolOverride` option (premature abstraction)
- Request body from debug logging (sensitive data exposure)
- `example/` directory (covered by docs/getting-started-pkg.md)

### Security

- Token files encrypted at rest with AES-256-GCM
- URL scheme validation on MCP client (`NewClient` returns error for non-http/https)
- Context timeouts on all operations (30s default, 5min for auth login)
- `client_secret` sent during token exchange when available

## [0.1.0] - 2026-03-20

### Added

- Go package API: `mcp2cli.New()` and `cli.Run()` for building CLIs from MCP servers
- mcp2cli-runner binary with `install`, `run`, `list`, `uninstall` commands
- MCP Streamable HTTP transport client (protocol version 2025-03-26)
- JSON Schema to CLI flag conversion (string, int, float, bool, enum)
- OAuth 2.1 authentication with PKCE (`auth login`, `auth logout`, `auth status`)
- Configuration persistence at `~/.config/mcp2cli/`
- Options: `WithHiddenTools`, `WithExtraHelp`, `WithConfigDir`

### Security

- URL scheme validation (http/https only)
- Response body size limit (10MB)
- HTTP client timeouts (30s)
- Token files stored with 0600 permissions
- OAuth callback server binds to loopback only

[Unreleased]: https://github.com/xiangma9712/mcp2cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/xiangma9712/mcp2cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/xiangma9712/mcp2cli/releases/tag/v0.1.0
