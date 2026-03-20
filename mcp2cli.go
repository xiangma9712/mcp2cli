package mcp2cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xiangma9712/mcp2cli/internal/auth"
	"github.com/xiangma9712/mcp2cli/internal/cfgstore"
	"github.com/xiangma9712/mcp2cli/internal/mcp"
	"github.com/xiangma9712/mcp2cli/internal/schema"
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
	hiddenTools map[string]bool
	extraHelp   string
}

// New creates a new CLI instance.
func New(name, url string, opts ...Option) *CLI {
	c := &CLI{
		name:        name,
		url:         url,
		configDir:   cfgstore.DefaultDir(),
		hiddenTools: make(map[string]bool),
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

// WithExtraHelp appends additional text to the help output.
func WithExtraHelp(text string) Option {
	return func(c *CLI) { c.extraHelp = text }
}

// Run executes the CLI with the given arguments and returns any error.
func (c *CLI) Run(args []string) error {
	if len(args) < 2 {
		return c.showHelp()
	}

	switch args[1] {
	case "auth":
		return c.handleAuth(args[2:])
	case "--help", "-h", "help":
		return c.showHelp()
	case "--version", "-v":
		fmt.Printf("%s %s\n", c.name, version)
		return nil
	default:
		return c.callTool(args[1], args[2:])
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
			fmt.Fprintf(os.Stderr, "Token expired. Please run: %s auth login\n", c.name)
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

	token, err := auth.LoadToken(c.configDir, c.name)
	if err == nil && !token.IsExpired() {
		client.SetHTTPClient(auth.AuthenticatedHTTPClient(token))
	}

	return client
}

// initAndListTools connects to the MCP server, performs the handshake,
// and returns all available tools.
func (c *CLI) initAndListTools(ctx context.Context) ([]mcp.Tool, *mcp.Client, error) {
	client := c.newClient()

	if _, err := client.Initialize(ctx, c.name, version); err != nil {
		return nil, nil, fmt.Errorf("connect to server: %w", err)
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list tools: %w", err)
	}

	return tools, client, nil
}

func (c *CLI) showHelp() error {
	ctx := context.Background()
	tools, _, err := c.initAndListTools(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n\n", c.name)
	fmt.Fprintln(os.Stderr, "Commands:")

	for _, t := range tools {
		if c.hiddenTools[t.Name] {
			continue
		}
		cmd := schema.ConvertTool(t)
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
	tools, client, err := c.initAndListTools(ctx)
	if err != nil {
		return err
	}

	tool := findTool(tools, toolName)
	if tool == nil {
		return fmt.Errorf("unknown command: %s", toolName)
	}

	cmd := schema.ConvertTool(*tool)

	// Check for --help before parsing flags
	for _, a := range args {
		if a == "--help" || a == "-h" {
			c.showToolHelp(cmd)
			return nil
		}
	}

	arguments, err := parseFlags(cmd.Flags, args)
	if err != nil {
		return fmt.Errorf("%w\nRun '%s %s --help' for usage", err, c.name, toolName)
	}

	if err := validateRequired(cmd.Flags, arguments, c.name, toolName); err != nil {
		return err
	}

	result, err := client.CallTool(ctx, toolName, arguments)
	if err != nil {
		return err
	}

	return printToolOutput(result)
}

func findTool(tools []mcp.Tool, name string) *mcp.Tool {
	for i := range tools {
		if tools[i].Name == name {
			return &tools[i]
		}
	}
	return nil
}

func validateRequired(flags []schema.Flag, arguments map[string]any, cliName, toolName string) error {
	for _, f := range flags {
		if f.Required {
			if _, ok := arguments[f.Name]; !ok {
				return fmt.Errorf("required flag --%s is missing\nRun '%s %s --help' for usage", f.Name, cliName, toolName)
			}
		}
	}
	return nil
}

func printToolOutput(result *mcp.ToolCallResult) error {
	if result.IsError {
		for _, item := range result.Content {
			if item.Type == "text" {
				fmt.Fprintln(os.Stderr, item.Text)
			}
		}
		return fmt.Errorf("tool returned an error")
	}

	for _, item := range result.Content {
		switch item.Type {
		case "text":
			fmt.Println(item.Text)
		case "image", "audio":
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
