# mcp2cli

Convert remote [MCP](https://modelcontextprotocol.io/) server tools into local CLI commands.

mcp2cli connects to an MCP server via Streamable HTTP transport, fetches the available tools, and generates a CLI interface with subcommands and flags derived from each tool's JSON Schema.

## Quick Start

There are two ways to use mcp2cli:

| | **Go package** | **mcp2cli-runner** |
|---|---|---|
| **Use when** | Building a branded CLI for your MCP server | Trying out any MCP server quickly |
| **How** | Import `github.com/xiangma9712/mcp2cli` | Install the runner binary |
| **Guide** | [Getting Started: Go Package](docs/getting-started-pkg.md) | [Getting Started: CLI](docs/getting-started-cli.md) |

### Go package

```go
package main

import (
    "fmt"
    "os"
    "github.com/xiangma9712/mcp2cli"
)

func main() {
    cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp")
    if err := cli.Run(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### mcp2cli-runner

```bash
go install github.com/xiangma9712/mcp2cli/cmd/mcp2cli-runner@latest

mcp2cli-runner install --name my-tool --url https://mcp.example.com/mcp
my-tool auth login
my-tool --help
```

## Authentication

OAuth 2.1 with PKCE is handled automatically. On `auth login`, the browser opens for authorization and tokens are stored encrypted at `~/.config/mcp2cli/<tool-name>/`.

```bash
my-tool auth login    # Opens browser for OAuth flow
my-tool auth status   # Check token status
my-tool auth logout   # Remove stored token
```

Each tool has its own token stored independently.

## Documentation

| Document | Description |
|---|---|
| [Getting Started: CLI](docs/getting-started-cli.md) | Install, authenticate, and use mcp2cli-runner |
| [Getting Started: Go Package](docs/getting-started-pkg.md) | Build a custom CLI with the Go library |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Development setup, conventions, PR process |
| [SECURITY.md](SECURITY.md) | Threat model, token storage, timeouts |
| [CHANGELOG.md](CHANGELOG.md) | Release history |

## Development

Prerequisites: [mise](https://mise.jdx.dev/)

```bash
mise install          # Install Go + tools
make test             # Run tests
make lint             # Run golangci-lint
make build            # Build runner binary
make ci               # Lint + test + build
```

Zero external dependencies — stdlib only.

## License

[MIT](LICENSE)
