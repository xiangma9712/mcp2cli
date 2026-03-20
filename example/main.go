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
