package auth

import (
	"net/http"
	"time"
)

// AuthenticatedHTTPClient returns an http.Client that attaches the Bearer token.
func AuthenticatedHTTPClient(token *Token) *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &tokenTransport{
			token: token,
			base:  http.DefaultTransport,
		},
	}
}

type tokenTransport struct {
	token *Token
	base  http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token.AccessToken)
	return t.base.RoundTrip(req)
}
