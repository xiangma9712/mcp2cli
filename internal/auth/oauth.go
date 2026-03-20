package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultHTTPClient is the default HTTP client used for OAuth operations.
// It can be replaced for testing.
var DefaultHTTPClient = &http.Client{Timeout: 30 * time.Second}

// Token represents an OAuth 2.1 token set.
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
}

// IsExpired returns true if the token has expired.
func (t *Token) IsExpired() bool {
	if t.ExpiresAt == 0 {
		return false
	}
	return time.Now().Unix() >= t.ExpiresAt
}

// OAuthConfig holds OAuth 2.1 configuration discovered from the MCP server.
type OAuthConfig struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
	ClientID              string
	ClientSecret          string
	Scopes                []string
}

// DiscoverOAuth fetches OAuth metadata from the MCP server's well-known endpoint.
func DiscoverOAuth(ctx context.Context, mcpURL string) (*OAuthConfig, error) {
	u, err := url.Parse(mcpURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s (only http and https are allowed)", u.Scheme)
	}

	wellKnown := fmt.Sprintf("%s://%s/.well-known/oauth-authorization-server", u.Scheme, u.Host)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return nil, err
	}

	resp, err := DefaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch oauth metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth metadata returned %d", resp.StatusCode)
	}

	var meta struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		RegistrationEndpoint  string `json:"registration_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decode oauth metadata: %w", err)
	}

	cfg := &OAuthConfig{
		AuthorizationEndpoint: meta.AuthorizationEndpoint,
		TokenEndpoint:         meta.TokenEndpoint,
	}

	return cfg, nil
}

// Login performs the OAuth 2.1 authorization code flow with PKCE.
func Login(ctx context.Context, cfg *OAuthConfig) (*Token, error) {
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}
	codeChallenge := computeCodeChallenge(codeVerifier)

	// Start local callback server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen for callback: %w", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	state, err := generateState()
	if err != nil {
		return nil, err
	}

	authURL := buildAuthURL(cfg, redirectURI, state, codeChallenge)
	fmt.Fprintf(os.Stderr, "Open this URL in your browser:\n\n  %s\n\nWaiting for authorization...\n", authURL)

	codeCh, errCh, server := startCallbackServer(listener, state)
	defer server.Shutdown(ctx)

	code, err := waitForAuthorizationCode(ctx, codeCh, errCh)
	if err != nil {
		return nil, err
	}

	return exchangeCode(ctx, cfg, code, redirectURI, codeVerifier)
}

func startCallbackServer(listener net.Listener, state string) (codeCh chan string, errCh chan error, server *http.Server) {
	codeCh = make(chan string, 1)
	errCh = make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errCh <- fmt.Errorf("oauth error: %s", errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			http.Error(w, "no code", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "<html><body><h1>Authorization successful!</h1><p>You can close this tab.</p></body></html>")
		codeCh <- code
	})

	server = &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	return
}

func waitForAuthorizationCode(ctx context.Context, codeCh chan string, errCh chan error) (string, error) {
	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func exchangeCode(ctx context.Context, cfg *OAuthConfig, code, redirectURI, codeVerifier string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}
	if cfg.ClientID != "" {
		data.Set("client_id", cfg.ClientID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := DefaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	token := &Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
	}
	if tokenResp.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Unix() + tokenResp.ExpiresIn
	}
	return token, nil
}

// SaveToken encrypts and writes the token to the config directory.
func SaveToken(configDir, toolName string, token *Token) error {
	dir := filepath.Join(configDir, toolName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	plaintext, err := json.Marshal(token)
	if err != nil {
		return err
	}
	encrypted, err := encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "token.json"), encrypted, 0600)
}

// LoadToken reads and decrypts the token from the config directory.
func LoadToken(configDir, toolName string) (*Token, error) {
	path := filepath.Join(configDir, toolName, "token.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	plaintext, err := decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("decrypt token: %w", err)
	}
	var token Token
	if err := json.Unmarshal(plaintext, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// RemoveToken removes the stored token.
func RemoveToken(configDir, toolName string) error {
	path := filepath.Join(configDir, toolName, "token.json")
	return os.Remove(path)
}

func buildAuthURL(cfg *OAuthConfig, redirectURI, state, codeChallenge string) string {
	params := url.Values{
		"response_type":         {"code"},
		"redirect_uri":          {redirectURI},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}
	if cfg.ClientID != "" {
		params.Set("client_id", cfg.ClientID)
	}
	if len(cfg.Scopes) > 0 {
		params.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	return cfg.AuthorizationEndpoint + "?" + params.Encode()
}

func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func computeCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
