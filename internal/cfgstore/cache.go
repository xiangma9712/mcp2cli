package cfgstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const defaultCacheTTL = 1 * time.Hour

// CachedTools stores a tools/list response with a timestamp.
type CachedTools struct {
	Tools    json.RawMessage `json:"tools"`
	CachedAt int64           `json:"cached_at"`
}

// SaveToolsCache writes the tools/list response to the cache file.
func SaveToolsCache(configDir, name string, toolsJSON []byte) error {
	dir := filepath.Join(configDir, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	cache := CachedTools{
		Tools:    toolsJSON,
		CachedAt: time.Now().Unix(),
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "tools_cache.json"), data, 0600)
}

// LoadToolsCache reads the cached tools/list response.
// Returns nil if the cache is missing or expired.
func LoadToolsCache(configDir, name string) []byte {
	path := filepath.Join(configDir, name, "tools_cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cache CachedTools
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}
	if time.Now().Unix()-cache.CachedAt > int64(defaultCacheTTL.Seconds()) {
		return nil
	}
	return cache.Tools
}

// InvalidateToolsCache removes the cached tools/list response.
func InvalidateToolsCache(configDir, name string) {
	path := filepath.Join(configDir, name, "tools_cache.json")
	os.Remove(path)
}
