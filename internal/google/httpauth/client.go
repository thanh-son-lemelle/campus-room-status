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
