# Getting Started: Go Package

Use `mcp2cli` as a library to build a custom CLI for your MCP server.

## Install

```bash
go get github.com/xiangma9712/mcp2cli@latest
```

## Minimal example

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

This gives you a full CLI with:
- Subcommands auto-generated from `tools/list`
- Flags derived from each tool's JSON Schema
- `auth login` / `auth logout` / `auth status`
- `--help` for all commands

## Options

### Hide tools

Exclude specific tools from the CLI:

```go
cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp",
    mcp2cli.WithHiddenTools("internal-debug", "admin-reset"),
)
```

### Custom help text

Append additional information to help output:

```go
cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp",
    mcp2cli.WithExtraHelp("\nDocumentation: https://example.com/docs"),
)
```

### Custom config directory

Override the default config directory (`~/.config/mcp2cli/`):

```go
cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp",
    mcp2cli.WithConfigDir("/custom/config/path"),
)
```

## How it works

1. `Run()` connects to the MCP server via Streamable HTTP transport
2. Performs the `initialize` handshake (MCP protocol 2025-03-26)
3. Calls `tools/list` to discover available tools
4. Maps the user's subcommand to a tool and parses `--flags` from the JSON Schema
5. Sends `tools/call` with the parsed arguments
6. Prints the result to stdout

## Error handling

`Run()` returns an `error` rather than calling `os.Exit()`, so you can wrap it with custom error handling, logging, or recovery logic.

```go
if err := cli.Run(os.Args); err != nil {
    log.Printf("command failed: %v", err)
    // custom cleanup, metrics, etc.
    os.Exit(1)
}
```

## Authentication

OAuth 2.1 with PKCE is handled automatically. Tokens are stored at `~/.config/mcp2cli/<tool-name>/token.json`. If a valid token exists, it is attached as a Bearer token to all MCP requests.
