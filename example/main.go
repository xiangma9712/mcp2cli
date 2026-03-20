package main

import (
	"os"

	"github.com/xiangma9712/mcp2cli"
)

func main() {
	cli := mcp2cli.New("my-tool", "https://mcp.example.com/mcp")
	cli.Run(os.Args)
}
