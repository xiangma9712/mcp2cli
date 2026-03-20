---
status: requested
date: 2026-03-20
---

# 03: Review findings

Multi-dimensional review of the entire repository.

## Findings

### 1. Nil pointer dereference on `resp.Result`
- **severity**: high
- **target**: internal/mcp/client.go:135,180,216
- **proposal**: Add nil check before dereferencing `*resp.Result`. `post()` can return `(nil, nil)` on 202 Accepted; callers must guard against nil response.
- **flagged by**: QA Engineer, Staff Engineer

### 2. Resource leak in `Login()` — TCP listener never closed on error paths
- **severity**: high
- **target**: internal/auth/oauth.go:94-152
- **proposal**: Add `defer listener.Close()` immediately after listener creation. Currently, if `generateState()` or `buildAuthURL()` fail, the listener leaks.
- **flagged by**: Staff Engineer, QA Engineer, Kent Beck

### 3. Duplicate initialize + listTools in `showHelp()` and `callTool()`
- **severity**: high
- **target**: mcp2cli.go:168-179, 205-216
- **proposal**: Extract shared method (e.g. `initAndListTools()`) to avoid redundant MCP handshakes and duplicated error handling.
- **flagged by**: Staff Engineer, Kent Beck, Junior Developer, Martin Fowler

### 4. `Run()` calls `os.Exit()` — library API should return errors
- **severity**: high
- **target**: mcp2cli.go:78-82
- **proposal**: Rename current `run()` to be the public API returning `error`. Let only the binary wrapper (`cmd/mcp2cli-runner`) handle `os.Exit()`.
- **flagged by**: Product Manager, Martin Fowler

### 5. Auth package has 0% test coverage
- **severity**: high
- **target**: internal/auth/
- **proposal**: Add tests for OAuth discovery, token save/load/remove, PKCE generation, and the callback server flow (using httptest).
- **flagged by**: QA Engineer

### 6. `callTool()` is too long with too many responsibilities
- **severity**: high
- **target**: mcp2cli.go:205-284
- **proposal**: Extract `formatToolOutput()`, `findTool()`, and `prepareToolCommand()` to separate concerns (init, lookup, override, parse, call, output).
- **flagged by**: Martin Fowler, Kent Beck

### 7. `Login()` is too long with multiple phases
- **severity**: high
- **target**: internal/auth/oauth.go:85-152
- **proposal**: Extract `startCallbackServer()` and `waitForAuthorizationCode()` to isolate HTTP server setup and channel orchestration.
- **flagged by**: Martin Fowler, Kent Beck

### 8. Schema conversion silently ignores malformed InputSchema
- **severity**: medium
- **target**: internal/schema/convert.go:34
- **proposal**: Return an error or log a warning when `properties` type assertion fails, rather than silently producing zero flags.
- **flagged by**: Staff Engineer, QA Engineer, Junior Developer

### 9. No HTTP timeouts on any client
- **severity**: medium
- **target**: internal/mcp/client.go:27, internal/auth/oauth.go:58,171
- **proposal**: Set default timeouts on `http.Client` (e.g. 30s). Unresponsive servers currently cause indefinite hangs.
- **flagged by**: QA Engineer, Staff Engineer

### 10. Error messages don't guide the user
- **severity**: medium
- **target**: mcp2cli.go:248,316
- **proposal**: Append `run '<tool> <cmd> --help' for usage` to "required flag missing" and "unknown flag" errors.
- **flagged by**: User, Product Manager

### 11. `WithToolOverride` is unclear and possibly premature
- **severity**: medium
- **target**: mcp2cli.go:64-67
- **proposal**: Remove or replace with explicit methods (e.g. `WithFlagDefault()`). Current callback API exposes internal schema types and has no documentation or tests.
- **flagged by**: Product Manager, Martin Fowler, User

### 12. Duplicate flag parsing in runner
- **severity**: medium
- **target**: cmd/mcp2cli-runner/main.go:56-68, 128-135
- **proposal**: Extract a shared `parseNamedArg(args, flagName)` helper to eliminate repeated `--name` / `--url` extraction logic.
- **flagged by**: Staff Engineer, Martin Fowler

### 13. Config file permissions 0644 (world-readable)
- **severity**: medium
- **target**: internal/cfgstore/cfgstore.go:34
- **proposal**: Use 0600 for config files (same as token files). The MCP server URL may be sensitive in multi-user environments.
- **flagged by**: QA Engineer, Junior Developer

### 14. Missing inline documentation for MCP protocol concepts
- **severity**: medium
- **target**: internal/mcp/client.go, internal/mcp/types.go
- **proposal**: Add brief doc comments explaining the MCP handshake flow, Streamable HTTP transport, and session ID semantics for contributors unfamiliar with the protocol.
- **flagged by**: Junior Developer
