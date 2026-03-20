# mcp2cli

Convert remote [MCP](https://modelcontextprotocol.io/) server tools into local CLI commands.

mcp2cli connects to an MCP server via Streamable HTTP transport, fetches the available tools, and generates a CLI interface with subcommands and flags derived from each tool's JSON Schema.

## Install

```bash
go install github.com/xiangma9712/mcp2cli/cmd/mcp2cli-runner@latest
```

## Usage

### As a Go package

Build a custom CLI in ~5 lines:

```go
package main

import (
    "os"
    "github.com/xiangma9712/mcp2cli"
)

func main() {
    cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp")
    cli.Run(os.Args)
}
```

Customize with options:

```go
cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp",
    mcp2cli.WithHiddenTools("internal-debug"),
    mcp2cli.WithExtraHelp("\nSee https://example.com/docs for more info."),
)
```

### As mcp2cli-runner

Install a remote MCP server as a local CLI:

```bash
# Register a tool
mcp2cli-runner install --name github-tool --url https://mcp.github.com/mcp

# Add the alias to your shell profile
alias github-tool='mcp2cli-runner run github-tool'

# Authenticate
github-tool auth login

# List available commands
github-tool --help

# Run a command
github-tool create-issue --repo owner/repo --title "Bug report"
```

Manage installed tools:

```bash
mcp2cli-runner list
mcp2cli-runner uninstall --name github-tool
```

## Authentication

mcp2cli supports OAuth 2.1 with PKCE. Credentials are stored in `~/.config/mcp2cli/<tool-name>/`.

```bash
my-tool auth login    # Start OAuth flow
my-tool auth status   # Check token status
my-tool auth logout   # Remove stored token
```

## Development

Prerequisites: [mise](https://mise.jdx.dev/)

```bash
mise install          # Install Go + tools
make test             # Run tests
make lint             # Run linter
make ci               # Full CI check
```

## License

MIT
