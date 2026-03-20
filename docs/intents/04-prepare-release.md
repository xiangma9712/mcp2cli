---
status: requested
date: 2026-03-20
---

# 04: Prepare release

Make the project release-ready with documentation, CI/CD, and security hardening.

## Documentation

- `CONTRIBUTING.md` — development setup, coding conventions, PR process
- `LICENSE` — MIT license
- `docs/getting-started-cli.md` — mcp2cli-runner install, auth, usage guide
- `docs/getting-started-pkg.md` — Go package integration guide with examples and options

## CD workflow

- `.github/workflows/release.yml`
  - Triggered by pushing a tag (`v*`)
  - Build multi-platform binaries (linux/darwin, amd64/arm64)
  - Create GitHub Release with changelog via `gh release create`
- `CHANGELOG.md`
  - Keep a Changelog format
  - Populate with initial v0.1.0 entries

## Security review and hardening

- Token storage security
  - Evaluate plaintext JSON token storage (`~/.config/mcp2cli/<name>/token.json`)
  - Consider OS keychain integration or encrypted storage
  - At minimum, document the threat model and limitations
- OAuth flow hardening
  - Validate redirect URI origin in callback handler
  - Add CSRF protection review for state parameter handling
  - Ensure callback server binds only to loopback (already done, verify)
- Input validation
  - Review flag parsing for injection risks
  - Validate MCP server URLs before connecting
- Dependency audit
  - Verify zero external dependencies is intentional and maintained
  - Document security posture in README or SECURITY.md
