package mcp2cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xiangma9712/mcp2cli/auth"
	"github.com/xiangma9712/mcp2cli/cfgstore"
	"github.com/xiangma9712/mcp2cli/mcp"
	"github.com/xiangma9712/mcp2cli/schema"
)

const version = "0.1.0"

// Option configures a CLI instance.
type Option func(*CLI)

// CLI converts MCP tools into a command-line interface.
type CLI struct {
	name      string
	url       string
	configDir string

	// Customization
	hiddenTools  map[string]bool
	toolOverride map[string]func(*schema.ToolCommand)
	extraHelp    string
}

// New creates a new CLI instance.
func New(name, url string, opts ...Option) *CLI {
	c := &CLI{
		name:         name,
		url:          url,
		configDir:    cfgstore.DefaultDir(),
		hiddenTools:  make(map[string]bool),
		toolOverride: make(map[string]func(*schema.ToolCommand)),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithConfigDir overrides the config directory.
func WithConfigDir(dir string) Option {
	return func(c *CLI) { c.configDir = dir }
}

// WithHiddenTools marks tools to be excluded from the CLI.
func WithHiddenTools(names ...string) Option {
	return func(c *CLI) {
		for _, n := range names {
			c.hiddenTools[n] = true
		}
	}
}

// WithToolOverride registers a function to modify a tool's CLI definition.
func WithToolOverride(toolName string, fn func(*schema.ToolCommand)) Option {
	return func(c *CLI) {
		c.toolOverride[toolName] = fn
	}
}

// WithExtraHelp appends additional text to the help output.
func WithExtraHelp(text string) Option {
	return func(c *CLI) { c.extraHelp = text }
}

// Run executes the CLI with the given arguments.
func (c *CLI) Run(args []string) {
	if err := c.run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func (c *CLI) run(args []string) error {
	if len(args) < 2 {
		return c.showHelp()
	}

	subcmd := args[1]

	switch subcmd {
	case "auth":
		return c.handleAuth(args[2:])
	case "--help", "-h", "help":
		return c.showHelp()
	case "--version", "-v":
		fmt.Printf("%s %s\n", c.name, version)
		return nil
	default:
		return c.callTool(subcmd, args[2:])
	}
}

func (c *CLI) handleAuth(args []string) error {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s auth <login|logout|status>\n", c.name)
		return nil
	}

	ctx := context.Background()

	switch args[0] {
	case "login":
		oauthCfg, err := auth.DiscoverOAuth(ctx, c.url)
		if err != nil {
			return fmt.Errorf("discover oauth: %w", err)
		}
		token, err := auth.Login(ctx, oauthCfg)
		if err != nil {
			return fmt.Errorf("login: %w", err)
		}
		if err := auth.SaveToken(c.configDir, c.name, token); err != nil {
			return fmt.Errorf("save token: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Login successful.")
		return nil

	case "logout":
		if err := auth.RemoveToken(c.configDir, c.name); err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "Not logged in.")
				return nil
			}
			return err
		}
		fmt.Fprintln(os.Stderr, "Logged out.")
		return nil

	case "status":
		token, err := auth.LoadToken(c.configDir, c.name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Not logged in.")
			return nil
		}
		if token.IsExpired() {
			fmt.Fprintln(os.Stderr, "Token expired. Please run: auth login")
		} else {
			fmt.Fprintln(os.Stderr, "Logged in.")
		}
		return nil

	default:
		return fmt.Errorf("unknown auth command: %s", args[0])
	}
}

func (c *CLI) newClient() *mcp.Client {
	client := mcp.NewClient(c.url)

	// Attach auth token if available
	token, err := auth.LoadToken(c.configDir, c.name)
	if err == nil && !token.IsExpired() {
		client.SetHTTPClient(auth.AuthenticatedHTTPClient(token))
	}

	return client
}

func (c *CLI) showHelp() error {
	ctx := context.Background()
	client := c.newClient()

	if _, err := client.Initialize(ctx, c.name, version); err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n\n", c.name)
	fmt.Fprintln(os.Stderr, "Commands:")

	for _, t := range tools {
		if c.hiddenTools[t.Name] {
			continue
		}
		cmd := schema.ConvertTool(t)
		if fn, ok := c.toolOverride[t.Name]; ok {
			fn(&cmd)
		}
		fmt.Fprintf(os.Stderr, "  %-20s %s\n", cmd.Name, cmd.Description)
	}

	fmt.Fprintf(os.Stderr, "\n  %-20s %s\n", "auth", "Manage authentication (login, logout, status)")
	fmt.Fprintf(os.Stderr, "\nRun '%s <command> --help' for more information on a command.\n", c.name)

	if c.extraHelp != "" {
		fmt.Fprintln(os.Stderr, c.extraHelp)
	}

	return nil
}

func (c *CLI) callTool(toolName string, args []string) error {
	ctx := context.Background()
	client := c.newClient()

	if _, err := client.Initialize(ctx, c.name, version); err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	var tool *mcp.Tool
	for i := range tools {
		if tools[i].Name == toolName {
			tool = &tools[i]
			break
		}
	}
	if tool == nil {
		return fmt.Errorf("unknown command: %s", toolName)
	}

	cmd := schema.ConvertTool(*tool)
	if fn, ok := c.toolOverride[toolName]; ok {
		fn(&cmd)
	}

	// Check for --help
	for _, a := range args {
		if a == "--help" || a == "-h" {
			c.showToolHelp(cmd)
			return nil
		}
	}

	arguments, err := parseFlags(cmd.Flags, args)
	if err != nil {
		return err
	}

	// Validate required flags
	for _, f := range cmd.Flags {
		if f.Required {
			if _, ok := arguments[f.Name]; !ok {
				return fmt.Errorf("required flag --%s is missing", f.Name)
			}
		}
	}

	result, err := client.CallTool(ctx, toolName, arguments)
	if err != nil {
		return err
	}

	if result.IsError {
		for _, item := range result.Content {
			if item.Type == "text" {
				fmt.Fprintln(os.Stderr, item.Text)
			}
		}
		os.Exit(1)
	}

	for _, item := range result.Content {
		switch item.Type {
		case "text":
			fmt.Println(item.Text)
		case "image", "audio":
			// Write binary data to stdout
			data, err := json.Marshal(item)
			if err == nil {
				fmt.Println(string(data))
			}
		}
	}

	return nil
}

func (c *CLI) showToolHelp(cmd schema.ToolCommand) {
	fmt.Fprintf(os.Stderr, "Usage: %s %s [flags]\n\n", c.name, cmd.Name)
	if cmd.Description != "" {
		fmt.Fprintln(os.Stderr, cmd.Description)
		fmt.Fprintln(os.Stderr)
	}
	if len(cmd.Flags) > 0 {
		fmt.Fprintln(os.Stderr, "Flags:")
		for _, f := range cmd.Flags {
			req := ""
			if f.Required {
				req = " (required)"
			}
			fmt.Fprintf(os.Stderr, "  --%-20s %s%s\n", f.Name, f.Description, req)
		}
	}
}

func parseFlags(flags []schema.Flag, args []string) (map[string]any, error) {
	result := make(map[string]any)
	flagMap := make(map[string]schema.Flag)
	for _, f := range flags {
		flagMap[f.Name] = f
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return nil, fmt.Errorf("unexpected argument: %s", arg)
		}

		name := strings.TrimPrefix(arg, "--")

		// Handle --flag=value
		if idx := strings.Index(name, "="); idx >= 0 {
			value := name[idx+1:]
			name = name[:idx]
			f, ok := flagMap[name]
			if !ok {
				return nil, fmt.Errorf("unknown flag: --%s", name)
			}
			parsed, err := parseValue(f.Type, value)
			if err != nil {
				return nil, fmt.Errorf("flag --%s: %w", name, err)
			}
			result[name] = parsed
			continue
		}

		f, ok := flagMap[name]
		if !ok {
			return nil, fmt.Errorf("unknown flag: --%s", name)
		}

		if f.Type == "bool" {
			result[name] = true
			continue
		}

		if i+1 >= len(args) {
			return nil, fmt.Errorf("flag --%s requires a value", name)
		}
		i++
		parsed, err := parseValue(f.Type, args[i])
		if err != nil {
			return nil, fmt.Errorf("flag --%s: %w", name, err)
		}
		result[name] = parsed
	}

	return result, nil
}

func parseValue(typ, value string) (any, error) {
	switch typ {
	case "int":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	default:
		return value, nil
	}
}
