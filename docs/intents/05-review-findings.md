---
status: requested
date: 2026-03-20
---

# 05: Review findings

Multi-dimensional review of the entire repository (post intent-04).

## Findings

### 1. CLAUDE.md references deleted `example/` directory
- **severity**: medium
- **target**: CLAUDE.md:7
- **proposal**: Remove the `example/` line from the project structure section.
- **flagged by**: Product Manager

### 2. `.golangci.yml` excludes non-existent `example/` directory
- **severity**: medium
- **target**: .golangci.yml:19
- **proposal**: Remove the `exclude-dirs` section.
- **flagged by**: Product Manager, Staff Engineer

### 3. Test helper reimplements `strings.Contains()`
- **severity**: medium
- **target**: internal/auth/oauth_test.go:155-166
- **proposal**: Replace custom `contains()` and `searchString()` with `strings.Contains()` from stdlib.
- **flagged by**: Staff Engineer, QA Engineer, Martin Fowler

### 4. No context timeout in CLI — hang risk with unresponsive servers
- **severity**: medium
- **target**: mcp2cli.go:178,213
- **proposal**: Use `context.WithTimeout(context.Background(), 30*time.Second)` instead of bare `context.Background()` in `showHelp()` and `callTool()`.
- **flagged by**: Staff Engineer, QA Engineer, User

### 5. Encryption key derived from executable path — moving binary breaks token decryption
- **severity**: medium
- **target**: internal/auth/encrypt.go:18-34
- **proposal**: Document this limitation in SECURITY.md and in the `deriveKey()` comment. Consider adding a fallback that tries decryption without executable path on failure.
- **flagged by**: Junior Developer, Martin Fowler

### 6. Silent token load failure in `newClient()` — auth errors hidden
- **severity**: medium
- **target**: mcp2cli.go:148-153
- **proposal**: Log a debug message when token load fails, so users can diagnose auth issues. Keep the silent fallback to unauthenticated but make it observable.
- **flagged by**: Junior Developer, Martin Fowler

### 7. Global `httpClient` in auth package not injectable
- **severity**: medium
- **target**: internal/auth/oauth.go:19
- **proposal**: Accept `*http.Client` as a parameter in `DiscoverOAuth()` and `exchangeCode()` instead of using a package-level variable, improving testability.
- **flagged by**: Staff Engineer, Martin Fowler

### 8. SSE JSON parse errors silently skipped
- **severity**: medium
- **target**: internal/mcp/client.go:137-138
- **proposal**: Log malformed SSE lines at debug level instead of silently continuing. Helps diagnose protocol issues.
- **flagged by**: Staff Engineer, QA Engineer

### 9. Complex types (array/object) passed as string with no user guidance
- **severity**: medium
- **target**: internal/schema/convert.go:93-95
- **proposal**: Append "(JSON)" to the flag description when the underlying type is array or object, so `--help` output guides the user.
- **flagged by**: User, Junior Developer

### 10. Error messages lack server URL context on connection failure
- **severity**: medium
- **target**: mcp2cli.go:167-172
- **proposal**: Include the MCP server URL in the "connect to server" error message so users can verify the endpoint.
- **flagged by**: User, Junior Developer

### 11. `handleAuth()` could be split into per-subcommand functions
- **severity**: medium
- **target**: mcp2cli.go:91-142
- **proposal**: Extract `handleAuthLogin()`, `handleAuthLogout()`, `handleAuthStatus()` for cohesion and testability.
- **flagged by**: Kent Beck, Martin Fowler
