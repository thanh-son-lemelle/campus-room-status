package oauth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthorizationFlow_StartBuildsAuthorizationURLWithExpectedScopes(t *testing.T) {
	store := &memoryRefreshTokenStore{}
	flow, err := NewAuthorizationFlow(Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		RedirectURI:      "http://localhost:8080/callback",
		Scopes:           []string{"scope-a", "scope-b"},
		RefreshTokenFile: "ignored",
		AuthURL:          "https://accounts.example.test/o/oauth2/auth",
		TokenURL:         "https://oauth2.example.test/token",
	}, store)
	if err != nil {
		t.Fatalf("expected flow creation to succeed, got %v", err)
	}

	authURL, state, err := flow.Start()
	if err != nil {
		t.Fatalf("expected start flow to succeed, got %v", err)
	}
	if strings.TrimSpace(state) == "" {
		t.Fatalf("expected non-empty state")
	}

	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}

	query := parsed.Query()
	if query.Get("client_id") != "client-id" {
		t.Fatalf("expected client_id in auth url, got %q", query.Get("client_id"))
	}
	if query.Get("redirect_uri") != "http://localhost:8080/callback" {
		t.Fatalf("expected redirect_uri in auth url, got %q", query.Get("redirect_uri"))
	}
	if query.Get("response_type") != "code" {
		t.Fatalf("expected response_type=code, got %q", query.Get("response_type"))
	}
	if query.Get("state") != state {
		t.Fatalf("expected state in auth url to match generated value")
	}

	scopeValue := query.Get("scope")
	if !strings.Contains(scopeValue, "scope-a") || !strings.Contains(scopeValue, "scope-b") {
		t.Fatalf("expected both scopes in auth url, got %q", scopeValue)
	}
}

func TestAuthorizationFlow_CallbackExchangesCodeAndPersistsRefreshToken(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}

		if got := r.Form.Get("grant_type"); got != "authorization_code" {
			t.Fatalf("expected authorization_code grant_type, got %q", got)
		}
		if got := r.Form.Get("code"); got != "valid-code" {
			t.Fatalf("expected code valid-code, got %q", got)
		}

		_, _ = io.WriteString(w, `{
			"access_token": "access-token",
			"refresh_token": "refresh-token",
			"expires_in": 3600,
			"token_type": "Bearer"
		}`)
	}))
	defer tokenServer.Close()

	store := &memoryRefreshTokenStore{}
	flow, err := NewAuthorizationFlow(Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		RedirectURI:      "http://localhost:8080/callback",
		Scopes:           []string{"scope-a"},
		RefreshTokenFile: "ignored",
		AuthURL:          "https://accounts.example.test/o/oauth2/auth",
		TokenURL:         tokenServer.URL,
	}, store)
	if err != nil {
		t.Fatalf("expected flow creation to succeed, got %v", err)
	}

	_, state, err := flow.Start()
	if err != nil {
		t.Fatalf("expected start flow to succeed, got %v", err)
	}

	if err := flow.Callback(context.Background(), state, "valid-code"); err != nil {
		t.Fatalf("expected callback exchange to succeed, got %v", err)
	}
	if store.token != "refresh-token" {
		t.Fatalf("expected refresh token persisted, got %q", store.token)
	}
}

func TestAuthorizationFlow_CallbackFailsWhenCodeInvalid(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{
			"error": "invalid_grant",
			"error_description": "Bad Request"
		}`)
	}))
	defer tokenServer.Close()

	store := &memoryRefreshTokenStore{}
	flow, err := NewAuthorizationFlow(Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		RedirectURI:      "http://localhost:8080/callback",
		Scopes:           []string{"scope-a"},
		RefreshTokenFile: "ignored",
		AuthURL:          "https://accounts.example.test/o/oauth2/auth",
		TokenURL:         tokenServer.URL,
	}, store)
	if err != nil {
		t.Fatalf("expected flow creation to succeed, got %v", err)
	}

	_, state, err := flow.Start()
	if err != nil {
		t.Fatalf("expected start flow to succeed, got %v", err)
	}

	err = flow.Callback(context.Background(), state, "invalid-code")
	if err == nil {
		t.Fatalf("expected callback error for invalid code")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "authorization code") {
		t.Fatalf("expected readable authorization code error, got %v", err)
	}
}

type memoryRefreshTokenStore struct {
	token string
}

func (s *memoryRefreshTokenStore) Save(_ context.Context, token string) error {
	s.token = token
	return nil
}

func (s *memoryRefreshTokenStore) Load(context.Context) (string, error) {
	return s.token, nil
}
