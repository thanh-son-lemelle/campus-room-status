package oauth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/oauth2"
)

type TokenProvider struct {
	oauthConfig *oauth2.Config
	store       RefreshTokenStore

	mu     sync.Mutex
	source oauth2.TokenSource
}

func NewTokenProvider(cfg Config, store RefreshTokenStore) (*TokenProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("refresh token store is required")
	}

	conf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       append([]string(nil), cfg.Scopes...),
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}

	return &TokenProvider{
		oauthConfig: conf,
		store:       store,
	}, nil
}

func (p *TokenProvider) Token(ctx context.Context) (string, error) {
	source, err := p.tokenSource(ctx)
	if err != nil {
		return "", err
	}

	token, err := source.Token()
	if err != nil {
		return "", classifyRefreshError(err)
	}

	accessToken := strings.TrimSpace(token.AccessToken)
	if accessToken == "" {
		return "", errors.New("google access token is empty")
	}

	refreshToken := strings.TrimSpace(token.RefreshToken)
	if refreshToken != "" {
		_ = p.store.Save(ctx, refreshToken)
	}

	return accessToken, nil
}

func (p *TokenProvider) tokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.source != nil {
		return p.source, nil
	}

	refreshToken, err := p.store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load refresh token: %w", err)
	}
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, errors.New("missing refresh token; run OAuth consent flow first")
	}

	source := oauth2.ReuseTokenSource(
		nil,
		p.oauthConfig.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}),
	)
	p.source = source
	return source, nil
}

func classifyRefreshError(err error) error {
	if err == nil {
		return nil
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "invalid_grant"),
		strings.Contains(lower, "revoked"),
		strings.Contains(lower, "expired"),
		strings.Contains(lower, "token has been expired or revoked"),
		strings.Contains(lower, "admin_policy_enforced"):
		return errors.New("refresh token is invalid, expired, revoked, or refused by admin policy")
	case strings.Contains(lower, "invalid_client"):
		return errors.New("oauth client credentials are invalid")
	default:
		return fmt.Errorf("refresh access token: %w", err)
	}
}
