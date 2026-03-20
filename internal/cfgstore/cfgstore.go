package cfgstore

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultDir returns the default config directory (~/.config/mcp2cli).
func DefaultDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mcp2cli")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mcp2cli")
}

// ToolConfig holds the configuration for a single tool.
type ToolConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Save writes the tool configuration to the config directory.
func Save(configDir string, cfg *ToolConfig) error {
	dir := filepath.Join(configDir, cfg.Name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0600)
}

// Load reads the tool configuration from the config directory.
func Load(configDir, name string) (*ToolConfig, error) {
	path := filepath.Join(configDir, name, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ToolConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
