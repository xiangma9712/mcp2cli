package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xiangma9712/mcp2cli/mcp"
)

// Flag represents a CLI flag derived from a JSON Schema property.
type Flag struct {
	Name        string
	Description string
	Type        string // "string", "int", "float", "bool"
	Required    bool
	Default     any
}

// ToolCommand represents a CLI subcommand derived from an MCP tool.
type ToolCommand struct {
	Name        string
	Description string
	Flags       []Flag
}

// ConvertTool converts an MCP tool definition to a CLI subcommand definition.
func ConvertTool(tool mcp.Tool) ToolCommand {
	cmd := ToolCommand{
		Name:        tool.Name,
		Description: tool.Description,
	}

	props, _ := tool.InputSchema["properties"].(map[string]any)
	requiredList := extractRequired(tool.InputSchema)

	var names []string
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		propRaw := props[name]
		prop, ok := propRaw.(map[string]any)
		if !ok {
			continue
		}

		flag := Flag{
			Name:        name,
			Description: descFromProp(prop),
			Type:        typeFromProp(prop),
			Required:    contains(requiredList, name),
			Default:     prop["default"],
		}
		cmd.Flags = append(cmd.Flags, flag)
	}

	return cmd
}

func extractRequired(schema map[string]any) []string {
	raw, ok := schema["required"].([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func typeFromProp(prop map[string]any) string {
	t, _ := prop["type"].(string)
	switch t {
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "string" // JSON array passed as string
	case "object":
		return "string" // JSON object passed as string
	default:
		return "string"
	}
}

func descFromProp(prop map[string]any) string {
	desc, _ := prop["description"].(string)
	if enum, ok := prop["enum"].([]any); ok && len(enum) > 0 {
		var vals []string
		for _, v := range enum {
			vals = append(vals, fmt.Sprintf("%v", v))
		}
		if desc != "" {
			desc += " "
		}
		desc += fmt.Sprintf("[%s]", strings.Join(vals, ", "))
	}
	return desc
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
