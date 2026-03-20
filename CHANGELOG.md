# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/xiangma9712/mcp2cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/xiangma9712/mcp2cli/releases/tag/v0.1.0
