package main

import (
	"fmt"
	"os"

	"github.com/xiangma9712/mcp2cli"
)

func main() {
	url := os.Getenv("MCP2CLI_NOTION_URL")
	if url == "" {
		url = "https://mcp.notion.com/mcp"
	}

	cli := mcp2cli.New("mcp2cli-notion", url,
		mcp2cli.WithExtraHelp("\nSet MCP2CLI_NOTION_URL to override the MCP server endpoint."),
	)
	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
