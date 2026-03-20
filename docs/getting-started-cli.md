# Getting Started: CLI (mcp2cli-runner)

mcp2cli-runner is a standalone binary that lets you use any remote MCP server as a local CLI tool.

## Install

```bash
go install github.com/xiangma9712/mcp2cli/cmd/mcp2cli-runner@latest
```

## Register a tool

```bash
mcp2cli-runner install --name my-tool --url https://mcp.example.com/mcp
```

This saves the tool configuration to `~/.config/mcp2cli/my-tool/config.json`.

Follow the printed instructions to add a shell alias or symlink so you can invoke it directly:

```bash
# Option 1: Shell alias (add to ~/.bashrc or ~/.zshrc)
alias my-tool='mcp2cli-runner run my-tool'

# Option 2: Symlink
ln -s $(which mcp2cli-runner) /usr/local/bin/my-tool
```

## Authenticate

If the MCP server requires OAuth:

```bash
my-tool auth login     # Opens browser for OAuth flow
my-tool auth status    # Check if you're logged in
my-tool auth logout    # Remove stored credentials
```

## Usage

```bash
# List available commands (fetched from MCP server)
my-tool --help

# Get help for a specific command
my-tool create-issue --help

# Run a command
my-tool create-issue --repo owner/repo --title "Bug report" --body "Description"
```

Flags are derived from the tool's JSON Schema. Required flags are marked in `--help` output.

## Manage installed tools

```bash
mcp2cli-runner list                        # List all installed tools
mcp2cli-runner uninstall --name my-tool    # Remove a tool
```

## Configuration

Tool configurations are stored in `~/.config/mcp2cli/<tool-name>/`:

- `config.json` — tool name and MCP server URL
- `token.json` — OAuth tokens (created after `auth login`)

The config directory respects `XDG_CONFIG_HOME` if set.
