package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiangma9712/mcp2cli"
	"github.com/xiangma9712/mcp2cli/cfgstore"
)

func main() {
	configDir := cfgstore.DefaultDir()

	// Detect if invoked as an alias (argv[0] is the tool name)
	execName := filepath.Base(os.Args[0])
	if execName != "mcp2cli-runner" {
		runAsTool(execName, configDir)
		return
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		cmdInstall(configDir, os.Args[2:])
	case "run":
		cmdRun(configDir, os.Args[2:])
	case "list":
		cmdList(configDir)
	case "uninstall":
		cmdUninstall(configDir, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: mcp2cli-runner <command> [args]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  install   --name <name> --url <url>   Install a tool")
	fmt.Fprintln(os.Stderr, "  uninstall --name <name>               Uninstall a tool")
	fmt.Fprintln(os.Stderr, "  run       <name> [args...]            Run a tool")
	fmt.Fprintln(os.Stderr, "  list                                  List installed tools")
}

func cmdInstall(configDir string, args []string) {
	var name, url string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 < len(args) {
				i++
				name = args[i]
			}
		case "--url":
			if i+1 < len(args) {
				i++
				url = args[i]
			}
		}
	}

	if name == "" || url == "" {
		fmt.Fprintln(os.Stderr, "Usage: mcp2cli-runner install --name <name> --url <url>")
		os.Exit(1)
	}

	cfg := &cfgstore.ToolConfig{Name: name, URL: url}
	if err := cfgstore.Save(configDir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	// Print alias instruction
	self, _ := os.Executable()
	if self == "" {
		self = "mcp2cli-runner"
	}
	alias := fmt.Sprintf("alias %s='%s run %s'", name, self, name)
	fmt.Fprintf(os.Stderr, "Installed %s.\n\nAdd this to your shell profile:\n\n  %s\n\n", name, alias)
	fmt.Fprintf(os.Stderr, "Or create a symlink:\n\n  ln -s %s %s/%s\n\n", self, filepath.Dir(self), name)
}

func cmdRun(configDir string, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: mcp2cli-runner run <name> [args...]")
		os.Exit(1)
	}

	name := args[0]
	toolArgs := []string{name} // fake argv[0]
	toolArgs = append(toolArgs, args[1:]...)

	runAsTool(name, configDir, toolArgs...)
}

func cmdList(configDir string) {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "No tools installed.")
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cfg, err := cfgstore.Load(configDir, e.Name())
		if err != nil {
			continue
		}
		fmt.Printf("%-20s %s\n", cfg.Name, cfg.URL)
	}
}

func cmdUninstall(configDir string, args []string) {
	var name string
	for i := 0; i < len(args); i++ {
		if args[i] == "--name" && i+1 < len(args) {
			i++
			name = args[i]
		}
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "Usage: mcp2cli-runner uninstall --name <name>")
		os.Exit(1)
	}

	dir := filepath.Join(configDir, name)
	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Uninstalled %s.\n", name)
	fmt.Fprintf(os.Stderr, "Don't forget to remove the alias from your shell profile.\n")
}

func runAsTool(name, configDir string, overrideArgs ...string) {
	cfg, err := cfgstore.Load(configDir, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Tool %q not installed. Run: mcp2cli-runner install --name %s --url <url>\n", name, name)
		os.Exit(1)
	}

	cli := mcp2cli.New(name, cfg.URL, mcp2cli.WithConfigDir(configDir))

	var args []string
	if len(overrideArgs) > 0 {
		args = overrideArgs
	} else {
		args = os.Args
	}

	// Ensure args[0] looks right
	if len(args) > 0 && !strings.Contains(args[0], name) {
		args[0] = name
	}

	cli.Run(args)
}
