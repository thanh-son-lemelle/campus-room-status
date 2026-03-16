package oauth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFileRefreshTokenStore_PersistsRefreshToken(t *testing.T) {
	path := t.TempDir() + "/oauth-refresh-token.json"
	store := NewFileRefreshTokenStore(path)

	if err := store.Save(context.Background(), "refresh-token-1"); err != nil {
		t.Fatalf("expected save to succeed, got %v", err)
	}

	token, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("expected load to succeed, got %v", err)
	}
	if token != "refresh-token-1" {
		t.Fatalf("expected persisted token refresh-token-1, got %q", token)
	}
}

func TestTokenProvider_RefreshesAccessTokenFromStoredRefreshToken(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}

		if got := r.Form.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("expected refresh_token grant_type, got %q", got)
		}
		if got := r.Form.Get("refresh_token"); got != "refresh-token" {
			t.Fatalf("expected refresh token refresh-token, got %q", got)
		}

		_, _ = io.WriteString(w, `{
			"access_token": "access-token-from-refresh",
			"expires_in": 3600,
			"token_type": "Bearer"
		}`)
	}))
	defer tokenServer.Close()

	store := &memoryStore{token: "refresh-token"}
	provider, err := NewTokenProvider(Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		RedirectURI:      "http://localhost:8080/callback",
		Scopes:           []string{"scope-a"},
		RefreshTokenFile: "ignored",
		AuthURL:          "https://accounts.example.test/o/oauth2/auth",
		TokenURL:         tokenServer.URL,
	}, store)
	if err != nil {
		t.Fatalf("expected provider creation to succeed, got %v", err)
	}

	token, err := provider.Token(context.Background())
	if err != nil {
		t.Fatalf("expected token refresh to succeed, got %v", err)
	}
	if token != "access-token-from-refresh" {
		t.Fatalf("unexpected access token %q", token)
	}
}

func TestTokenProvider_FailsReadablyWhenRefreshTokenMissing(t *testing.T) {
	store := &memoryStore{}
	provider, err := NewTokenProvider(Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		RedirectURI:      "http://localhost:8080/callback",
		Scopes:           []string{"scope-a"},
		RefreshTokenFile: "ignored",
		AuthURL:          "https://accounts.example.test/o/oauth2/auth",
		TokenURL:         "https://oauth2.example.test/token",
	}, store)
	if err != nil {
		t.Fatalf("expected provider creation to succeed, got %v", err)
	}

	_, err = provider.Token(context.Background())
	if err == nil {
		t.Fatalf("expected missing refresh token error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "refresh token") {
		t.Fatalf("expected readable refresh token error, got %v", err)
	}
}

func TestTokenProvider_FailsReadablyWhenRefreshTokenRevoked(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{
			"error": "invalid_grant",
			"error_description": "Token has been expired or revoked."
		}`)
	}))
	defer tokenServer.Close()

	store := &memoryStore{token: "revoked-refresh-token"}
	provider, err := NewTokenProvider(Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		RedirectURI:      "http://localhost:8080/callback",
		Scopes:           []string{"scope-a"},
		RefreshTokenFile: "ignored",
		AuthURL:          "https://accounts.example.test/o/oauth2/auth",
		TokenURL:         tokenServer.URL,
	}, store)
	if err != nil {
		t.Fatalf("expected provider creation to succeed, got %v", err)
	}

	_, err = provider.Token(context.Background())
	if err == nil {
		t.Fatalf("expected revoked refresh token error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "revoked") &&
		!strings.Contains(strings.ToLower(err.Error()), "expired") &&
		!strings.Contains(strings.ToLower(err.Error()), "policy") {
		t.Fatalf("expected readable revoked/expired/policy error, got %v", err)
	}
}

type memoryStore struct {
	token string
}

func (s *memoryStore) Save(_ context.Context, token string) error {
	s.token = token
	return nil
}

func (s *memoryStore) Load(context.Context) (string, error) {
	return s.token, nil
}
