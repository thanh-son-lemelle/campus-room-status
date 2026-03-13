package adminsdk

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
)

type ServiceAccountTokenProviderConfig struct {
	CredentialsJSON []byte
	Subject         string
	Scopes          []string
}

func NewServiceAccountTokenProvider(cfg ServiceAccountTokenProviderConfig) (TokenProvider, error) {
	credentialsJSON := strings.TrimSpace(string(cfg.CredentialsJSON))
	if credentialsJSON == "" {
		return nil, errors.New("service account credentials are required")
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{admin.AdminDirectoryResourceCalendarReadonlyScope}
	}

	jwtConfig, err := google.JWTConfigFromJSON([]byte(credentialsJSON), scopes...)
	if err != nil {
		return nil, fmt.Errorf("parse service account credentials: %w", err)
	}

	subject := strings.TrimSpace(cfg.Subject)
	if subject != "" {
		jwtConfig.Subject = subject
	}

	return &serviceAccountTokenProvider{
		sourceFactory: func(ctx context.Context) oauth2.TokenSource {
			return oauth2.ReuseTokenSource(nil, jwtConfig.TokenSource(ctx))
		},
	}, nil
}

type serviceAccountTokenProvider struct {
	sourceFactory func(context.Context) oauth2.TokenSource

	mu     sync.Mutex
	source oauth2.TokenSource
}

func (p *serviceAccountTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	if p.source == nil {
		p.source = p.sourceFactory(ctx)
	}
	source := p.source
	p.mu.Unlock()

	token, err := source.Token()
	if err != nil {
		return "", fmt.Errorf("retrieve service account token: %w", err)
	}

	accessToken := strings.TrimSpace(token.AccessToken)
	if accessToken == "" {
		return "", errors.New("service account access token is empty")
	}

	return accessToken, nil
}
