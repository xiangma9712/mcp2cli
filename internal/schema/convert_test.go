package schema

import (
	"testing"

	"github.com/xiangma9712/mcp2cli/internal/mcp"
)

func TestConvertTool(t *testing.T) {
	tool := mcp.Tool{
		Name:        "create-issue",
		Description: "Create a new issue",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "Issue title",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Issue body",
				},
				"priority": map[string]any{
					"type":        "integer",
					"description": "Priority level",
				},
				"draft": map[string]any{
					"type":        "boolean",
					"description": "Create as draft",
				},
				"labels": map[string]any{
					"type":        "array",
					"description": "Labels to add",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Status",
					"enum":        []any{"open", "closed"},
				},
			},
			"required": []any{"title"},
		},
	}

	cmd := ConvertTool(tool)

	if cmd.Name != "create-issue" {
		t.Errorf("expected name create-issue, got %s", cmd.Name)
	}
	if cmd.Description != "Create a new issue" {
		t.Errorf("expected description, got %s", cmd.Description)
	}

	if len(cmd.Flags) != 6 {
		t.Fatalf("expected 6 flags, got %d", len(cmd.Flags))
	}

	// Flags are sorted alphabetically
	flagMap := make(map[string]Flag)
	for _, f := range cmd.Flags {
		flagMap[f.Name] = f
	}

	titleFlag := flagMap["title"]
	if titleFlag.Type != "string" {
		t.Errorf("expected title type string, got %s", titleFlag.Type)
	}
	if !titleFlag.Required {
		t.Error("expected title to be required")
	}

	priorityFlag := flagMap["priority"]
	if priorityFlag.Type != "int" {
		t.Errorf("expected priority type int, got %s", priorityFlag.Type)
	}

	draftFlag := flagMap["draft"]
	if draftFlag.Type != "bool" {
		t.Errorf("expected draft type bool, got %s", draftFlag.Type)
	}

	labelsFlag := flagMap["labels"]
	if labelsFlag.Type != "string" {
		t.Errorf("expected labels type string (json), got %s", labelsFlag.Type)
	}
	if labelsFlag.Description != "Labels to add (JSON)" {
		t.Errorf("expected array description with (JSON), got %q", labelsFlag.Description)
	}

	statusFlag := flagMap["status"]
	if statusFlag.Description != "Status [open, closed]" {
		t.Errorf("expected enum in description, got %q", statusFlag.Description)
	}
}
