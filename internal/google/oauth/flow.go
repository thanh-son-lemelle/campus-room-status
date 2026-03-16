package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type AuthorizationFlow struct {
	oauthConfig *oauth2.Config
	store       RefreshTokenStore
	states      *stateStore
}

// NewAuthorizationFlow creates a new authorization flow.
//
// Summary:
// - Creates a new authorization flow.
//
// Attributes:
// - cfg (Config): Input parameter.
// - store (RefreshTokenStore): Input parameter.
//
// Returns:
// - value1 (*AuthorizationFlow): Returned value.
// - value2 (error): Returned value.
func NewAuthorizationFlow(cfg Config, store RefreshTokenStore) (*AuthorizationFlow, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("refresh token store is required")
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       append([]string(nil), cfg.Scopes...),
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}

	return &AuthorizationFlow{
		oauthConfig: oauthConfig,
		store:       store,
		states:      newStateStore(10 * time.Minute),
	}, nil
}

// Start starts function behavior.
//
// Summary:
// - Starts function behavior.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (string): Returned value.
// - value2 (string): Returned value.
// - value3 (error): Returned value.
func (f *AuthorizationFlow) Start() (string, string, error) {
	if f == nil || f.oauthConfig == nil {
		return "", "", errors.New("oauth flow is not configured")
	}

	state, err := randomState()
	if err != nil {
		return "", "", err
	}
	f.states.Put(state)

	authURL := f.oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("include_granted_scopes", "true"),
	)

	return authURL, state, nil
}

// Callback callbacks function behavior.
//
// Summary:
// - Callbacks function behavior.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - state (string): Input parameter.
// - code (string): Input parameter.
//
// Returns:
// - value1 (error): Returned value.
func (f *AuthorizationFlow) Callback(ctx context.Context, state, code string) error {
	if f == nil || f.oauthConfig == nil {
		return errors.New("oauth flow is not configured")
	}
	if !f.states.Consume(state) {
		return errors.New("invalid or expired oauth state")
	}
	if strings.TrimSpace(code) == "" {
		return errors.New("authorization code is required")
	}

	token, err := f.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("authorization code exchange failed: %w", classifyExchangeError(err))
	}

	refreshToken := strings.TrimSpace(token.RefreshToken)
	if refreshToken == "" {
		existing, loadErr := f.store.Load(ctx)
		if loadErr != nil {
			return fmt.Errorf("load persisted refresh token: %w", loadErr)
		}
		refreshToken = strings.TrimSpace(existing)
	}
	if refreshToken == "" {
		return errors.New("authorization succeeded but no refresh token was returned; re-consent with prompt=consent and offline access")
	}

	if err := f.store.Save(ctx, refreshToken); err != nil {
		return fmt.Errorf("persist refresh token: %w", err)
	}

	return nil
}

// randomState randoms state.
//
// Summary:
// - Randoms state.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (string): Returned value.
// - value2 (error): Returned value.
func randomState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

type stateStore struct {
	ttl    time.Duration
	mu     sync.Mutex
	values map[string]time.Time
}

// newStateStore creates a new state store.
//
// Summary:
// - Creates a new state store.
//
// Attributes:
// - ttl (time.Duration): Input parameter.
//
// Returns:
// - value1 (*stateStore): Returned value.
func newStateStore(ttl time.Duration) *stateStore {
	return &stateStore{
		ttl:    ttl,
		values: make(map[string]time.Time),
	}
}

// Put puts function behavior.
//
// Summary:
// - Puts function behavior.
//
// Attributes:
// - state (string): Input parameter.
//
// Returns:
// - None.
func (s *stateStore) Put(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	s.cleanupExpiredLocked(now)
	s.values[strings.TrimSpace(state)] = now.Add(s.ttl)
}

// Consume consumes function behavior.
//
// Summary:
// - Consumes function behavior.
//
// Attributes:
// - state (string): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
func (s *stateStore) Consume(state string) bool {
	trimmed := strings.TrimSpace(state)
	if trimmed == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	s.cleanupExpiredLocked(now)

	expiresAt, exists := s.values[trimmed]
	if !exists || !expiresAt.After(now) {
		return false
	}

	delete(s.values, trimmed)
	return true
}

// cleanupExpiredLocked cleanups expired locked.
//
// Summary:
// - Cleanups expired locked.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - None.
func (s *stateStore) cleanupExpiredLocked(now time.Time) {
	for key, expiresAt := range s.values {
		if !expiresAt.After(now) {
			delete(s.values, key)
		}
	}
}

// classifyExchangeError classifies exchange error.
//
// Summary:
// - Classifies exchange error.
//
// Attributes:
// - err (error): Input parameter.
//
// Returns:
// - value1 (error): Returned value.
func classifyExchangeError(err error) error {
	if err == nil {
		return nil
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "invalid_grant"):
		return errors.New("invalid authorization code or consent was denied/expired")
	case strings.Contains(lower, "access_denied"):
		return errors.New("oauth consent was denied")
	default:
		if strings.Contains(lower, "token has been expired or revoked") {
			return errors.New("google refused the authorization code")
		}
		return err
	}
}

// BuildAuthorizationURLPreview builds authorization url preview.
//
// Summary:
// - Builds authorization url preview.
//
// Attributes:
// - cfg (Config): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
// - value2 (error): Returned value.
func BuildAuthorizationURLPreview(cfg Config) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", err
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

	state := "preview-state"
	authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	_, err := url.Parse(authURL)
	if err != nil {
		return "", err
	}
	return authURL, nil
}
