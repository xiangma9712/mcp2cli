---
status: requested
date: 2026-03-20
---

# 06: Review findings

Multi-dimensional review of the entire repository (post feat/notion-e2e).

## Findings

### 1. `AuthenticatedHTTPClient` creates client with no timeout
- **severity**: high
- **target**: internal/auth/transport.go:7-12
- **proposal**: Set `Timeout: 30 * time.Second` on the `http.Client` returned by `AuthenticatedHTTPClient`, matching the default used in the MCP client.
- **flagged by**: Staff Engineer

### 2. DEBUG logging exposes request bodies including secrets/tokens
- **severity**: high
- **target**: internal/mcp/client.go:86
- **proposal**: Redact or truncate request body logging when it contains sensitive fields. At minimum, skip logging bodies for `initialize` and `tools/call` that may contain auth tokens in headers.
- **flagged by**: QA Engineer

### 3. Auth error hint duplicated in 4 places
- **severity**: high
- **target**: mcp2cli.go:196,201,284,298
- **proposal**: Extract `"\n\nIf authentication is required, run: %s auth login"` into a helper function (e.g. `authHint()`) to eliminate duplication.
- **flagged by**: Kent Beck

### 4. `loadTools()` returns nil client on cache hit — implicit contract
- **severity**: medium
- **target**: mcp2cli.go:180-210, 291-300
- **proposal**: Return an explicit type or use a separate code path. Currently `callTool()` must check `client == nil` and lazily initialize, which is error-prone.
- **flagged by**: Junior Developer, Martin Fowler

### 5. Cache invalidation scattered across auth handlers
- **severity**: medium
- **target**: mcp2cli.go:128, 141
- **proposal**: Move cache invalidation into the auth layer (e.g. `auth.Logout()` calls `cfgstore.InvalidateToolsCache` internally) so callers don't need to remember.
- **flagged by**: Martin Fowler, Junior Developer

### 6. Context timeout pattern repeated — extract helper
- **severity**: medium
- **target**: mcp2cli.go:113, 229, 259
- **proposal**: Add `func (c *CLI) requestContext() (context.Context, context.CancelFunc)` to wrap `context.WithTimeout(context.Background(), defaultRequestTimeout)`.
- **flagged by**: Kent Beck, Martin Fowler

### 7. `schema/convert.go` uses `log.Printf` bypassing debuglog
- **severity**: medium
- **target**: internal/schema/convert.go:39
- **proposal**: Replace `log.Printf` with `debuglog.New().Printf` or remove the warning (callers handle empty flags already).
- **flagged by**: Staff Engineer, QA Engineer

### 8. `"tool returned an error"` message is too generic
- **severity**: medium
- **target**: mcp2cli.go:337
- **proposal**: Remove the redundant wrapper error. The tool's own error content is already printed to stderr; the additional generic message adds no value.
- **flagged by**: User, Junior Developer

### 9. Timeout values undocumented — users see "context deadline exceeded"
- **severity**: medium
- **target**: mcp2cli.go:23, 113
- **proposal**: Document default timeouts (30s for commands, 5min for auth login) in README and SECURITY.md. Wrap timeout errors with a user-friendly message.
- **flagged by**: User, Junior Developer

### 10. `firstSentence()` panics if `maxLen < 4`
- **severity**: medium
- **target**: mcp2cli.go:213-226
- **proposal**: Add guard: `if maxLen < 4 { return s }` before the truncation logic to prevent negative slice index.
- **flagged by**: QA Engineer
