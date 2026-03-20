package mcp2cli

import (
	"testing"

	"github.com/xiangma9712/mcp2cli/internal/schema"
)

func TestParseFlags(t *testing.T) {
	flags := []schema.Flag{
		{Name: "title", Type: "string", Required: true},
		{Name: "count", Type: "int"},
		{Name: "draft", Type: "bool"},
		{Name: "ratio", Type: "float"},
	}

	tests := []struct {
		name    string
		args    []string
		want    map[string]any
		wantErr bool
	}{
		{
			name: "basic string",
			args: []string{"--title", "hello"},
			want: map[string]any{"title": "hello"},
		},
		{
			name: "int flag",
			args: []string{"--count", "42"},
			want: map[string]any{"count": 42},
		},
		{
			name: "bool flag (no value)",
			args: []string{"--draft"},
			want: map[string]any{"draft": true},
		},
		{
			name: "float flag",
			args: []string{"--ratio", "3.14"},
			want: map[string]any{"ratio": 3.14},
		},
		{
			name: "equals syntax",
			args: []string{"--title=world"},
			want: map[string]any{"title": "world"},
		},
		{
			name: "multiple flags",
			args: []string{"--title", "test", "--count", "5", "--draft"},
			want: map[string]any{"title": "test", "count": 5, "draft": true},
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown", "val"},
			wantErr: true,
		},
		{
			name:    "missing value",
			args:    []string{"--title"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags(flags, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("flag %s: expected %v (%T), got %v (%T)", k, v, v, got[k], got[k])
				}
			}
		})
	}
}
