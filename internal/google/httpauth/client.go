package httpauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// NewAuthorizedHTTPClient wraps an HTTP client with bearer token authorization.
//
// Summary:
// - Clones the input client and injects an authorization transport.
// - Uses the default HTTP transport when the input client has none.
//
// Attributes:
// - client: Base HTTP client to clone and wrap.
// - tokenProvider: Provider used to fetch bearer tokens for outbound requests.
//
// Returns:
// - *http.Client: Cloned HTTP client that injects Authorization headers on requests.
func NewAuthorizedHTTPClient(client *http.Client, tokenProvider TokenProvider) *http.Client {
	if client == nil {
		client = &http.Client{}
	}

	clone := *client
	transport := clone.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	clone.Transport = authorizedTransport{
		base:          transport,
		tokenProvider: tokenProvider,
	}

	return &clone
}

type authorizedTransport struct {
	base          http.RoundTripper
	tokenProvider TokenProvider
}

// RoundTrip sends a request with an optional bearer token.
//
// Summary:
// - Clones the request to avoid mutating the original headers.
// - Retrieves a token and sets the Authorization header when non-empty.
//
// Attributes:
// - req: Outbound HTTP request to authorize and forward.
//
// Returns:
// - *http.Response: HTTP response from the wrapped transport when successful.
// - error: Token retrieval or transport execution error.
func (t authorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())

	token, err := t.tokenProvider.Token(req.Context())
	if err != nil {
		return nil, fmt.Errorf("retrieve access token: %w", err)
	}

	if strings.TrimSpace(token) != "" {
		cloned.Header.Set("Authorization", "Bearer "+token)
	}

	return t.base.RoundTrip(cloned)
}
