# Security

## Threat model

mcp2cli is a CLI tool that connects to remote MCP servers on behalf of the user. The trust model assumes:

- The MCP server URL is provided by the user and trusted
- The local filesystem is controlled by the user (single-user system)
- Network communication may traverse untrusted networks

## Token storage

OAuth tokens are stored at `~/.config/mcp2cli/<tool-name>/token.json` with `0600` file permissions (owner read/write only). Tokens are **encrypted at rest** using AES-256-GCM with a key deterministically derived from the executable path and runtime environment.

**Encryption details:**
- AES-256-GCM with random nonce per write
- Key derived via SHA-256 from: executable path + OS/arch + fixed salt
- This is **obfuscation**, not strong encryption — a determined attacker with access to the binary can reproduce the key

**Limitations:**
- Encryption key is deterministic and derivable from the binary
- Moving or reinstalling the binary changes the derived key, making existing tokens undecryptable — users must re-run `auth login`
- Do not use on shared or multi-user systems where other users may have elevated access
- No automatic token rotation; expired tokens are detected but not refreshed

**Mitigations:**
- Encrypted at rest (prevents casual reading)
- File permissions restrict access to the owner
- Config directory uses `0700` permissions
- Tokens are scoped to specific MCP servers

## OAuth 2.1 implementation

- Uses PKCE (Proof Key for Code Exchange) with S256 challenge method
- State parameter generated with `crypto/rand` (16 bytes) for CSRF protection
- Callback server binds exclusively to `127.0.0.1` (loopback), preventing external access
- Dynamic port allocation (`port 0`) prevents port collision

## Network security

- URL scheme validation enforces `http` or `https` only
- Response body size limited to 10MB to prevent DoS
- Bearer tokens are only sent to the configured MCP server endpoint

## Timeouts

| Operation | Timeout | Notes |
|-----------|---------|-------|
| MCP requests (help, tool calls) | 30 seconds | Covers initialize + tools/list or tools/call |
| OAuth login flow | 5 minutes | Allows time for browser-based authorization |
| HTTP client (per request) | 30 seconds | Applied to all HTTP connections |

If a timeout is exceeded, the error message will include "context deadline exceeded".

## Reporting vulnerabilities

If you discover a security vulnerability, please report it privately via [GitHub Security Advisories](https://github.com/xiangma9712/mcp2cli/security/advisories/new).

Do not open a public issue for security vulnerabilities.
