package cfgstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	cfg := &ToolConfig{Name: "test-tool", URL: "https://example.com/mcp"}
	if err := Save(dir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path := filepath.Join(dir, "test-tool", "config.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := Load(dir, "test-tool")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != cfg.Name || loaded.URL != cfg.URL {
		t.Errorf("loaded config mismatch: got %+v, want %+v", loaded, cfg)
	}
}

func TestLoadNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}
