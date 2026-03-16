package oauth

import (
	"strings"
	"testing"
)

func TestLoadConfigFromEnv_LoadsRequiredValues(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "http://localhost:8080/api/v1/auth/google/callback")
	t.Setenv("GOOGLE_OAUTH_SCOPES", "scope-a, scope-b")
	t.Setenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE", "tokens/oauth-refresh.json")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected config loading to succeed, got %v", err)
	}

	if cfg.ClientID != "client-id" {
		t.Fatalf("expected client id to be loaded, got %q", cfg.ClientID)
	}
	if cfg.ClientSecret != "client-secret" {
		t.Fatalf("expected client secret to be loaded, got %q", cfg.ClientSecret)
	}
	if cfg.RedirectURI != "http://localhost:8080/api/v1/auth/google/callback" {
		t.Fatalf("unexpected redirect uri: %q", cfg.RedirectURI)
	}
	if len(cfg.Scopes) != 2 || cfg.Scopes[0] != "scope-a" || cfg.Scopes[1] != "scope-b" {
		t.Fatalf("unexpected scopes: %v", cfg.Scopes)
	}
	if cfg.RefreshTokenFile != "tokens/oauth-refresh.json" {
		t.Fatalf("unexpected refresh token file: %q", cfg.RefreshTokenFile)
	}
}

func TestLoadConfigFromEnv_FailsWhenClientIDMissing(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "http://localhost/callback")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected missing client id error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "client_id") {
		t.Fatalf("expected client_id mention in error, got %v", err)
	}
}

func TestLoadConfigFromEnv_FailsWhenClientSecretMissing(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "http://localhost/callback")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected missing client secret error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "client_secret") {
		t.Fatalf("expected client_secret mention in error, got %v", err)
	}
}

func TestLoadConfigFromEnv_FailsWhenRedirectURIMissing(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected missing redirect uri error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "redirect_uri") {
		t.Fatalf("expected redirect_uri mention in error, got %v", err)
	}
}
