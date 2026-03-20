package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverOAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/oauth-authorization-server" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"authorization_endpoint": "https://auth.example.com/authorize",
			"token_endpoint":         "https://auth.example.com/token",
		})
	}))
	defer server.Close()

	cfg, err := DiscoverOAuth(context.Background(), server.URL+"/mcp")
	if err != nil {
		t.Fatalf("DiscoverOAuth: %v", err)
	}
	if cfg.AuthorizationEndpoint != "https://auth.example.com/authorize" {
		t.Errorf("unexpected auth endpoint: %s", cfg.AuthorizationEndpoint)
	}
	if cfg.TokenEndpoint != "https://auth.example.com/token" {
		t.Errorf("unexpected token endpoint: %s", cfg.TokenEndpoint)
	}
}

func TestDiscoverOAuthNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := DiscoverOAuth(context.Background(), server.URL+"/mcp")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestSaveLoadRemoveToken(t *testing.T) {
	dir := t.TempDir()
	toolName := "test-tool"

	token := &Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    9999999999,
	}

	if err := SaveToken(dir, toolName, token); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	// Verify file permissions and that content is encrypted (not plaintext JSON)
	path := filepath.Join(dir, toolName, "token.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("token file not found: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read token file: %v", err)
	}
	if len(raw) > 0 && raw[0] == '{' {
		t.Error("token file appears to be plaintext JSON; expected encrypted data")
	}

	loaded, err := LoadToken(dir, toolName)
	if err != nil {
		t.Fatalf("LoadToken: %v", err)
	}
	if loaded.AccessToken != token.AccessToken {
		t.Errorf("access token mismatch: got %s", loaded.AccessToken)
	}
	if loaded.IsExpired() {
		t.Error("token should not be expired")
	}

	if err := RemoveToken(dir, toolName); err != nil {
		t.Fatalf("RemoveToken: %v", err)
	}

	_, err = LoadToken(dir, toolName)
	if err == nil {
		t.Error("expected error after removing token")
	}
}

func TestTokenExpired(t *testing.T) {
	token := &Token{ExpiresAt: 1}
	if !token.IsExpired() {
		t.Error("token with ExpiresAt=1 should be expired")
	}

	token2 := &Token{ExpiresAt: 0}
	if token2.IsExpired() {
		t.Error("token with ExpiresAt=0 should not be expired")
	}
}

func TestGenerateCodeVerifierAndChallenge(t *testing.T) {
	v1, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("generateCodeVerifier: %v", err)
	}
	v2, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("generateCodeVerifier: %v", err)
	}
	if v1 == v2 {
		t.Error("expected unique verifiers")
	}

	c1 := computeCodeChallenge(v1)
	c2 := computeCodeChallenge(v1)
	if c1 != c2 {
		t.Error("same verifier should produce same challenge")
	}
	if c1 == "" {
		t.Error("challenge should not be empty")
	}
}

func TestBuildAuthURL(t *testing.T) {
	cfg := &OAuthConfig{
		AuthorizationEndpoint: "https://auth.example.com/authorize",
		ClientID:              "test-client",
		Scopes:                []string{"read", "write"},
	}
	url := buildAuthURL(cfg, "http://localhost:8080/callback", "test-state", "test-challenge")
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	for _, expected := range []string{"client_id=test-client", "state=test-state", "code_challenge=test-challenge", "scope=read+write"} {
		if !contains(url, expected) {
			t.Errorf("URL missing %q: %s", expected, url)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
