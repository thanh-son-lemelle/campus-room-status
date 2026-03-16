package main

import (
	"path/filepath"
	"testing"

	goauth "campus-room-status/internal/google/oauth"
)

func TestOAuthAutoLaunchStartURLFromEnv_ReturnsFalseWhenDataSourceIsNotGoogle(t *testing.T) {
	setOAuthAutoLaunchEnv(t)
	t.Setenv("DATA_SOURCE", "static")

	startURL, ok := oauthAutoLaunchStartURLFromEnv()
	if ok {
		t.Fatalf("expected auto-launch disabled for non-google data source, got URL %q", startURL)
	}
}

func TestOAuthAutoLaunchStartURLFromEnv_ReturnsFalseWhenOAuthConfigMissing(t *testing.T) {
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "")

	startURL, ok := oauthAutoLaunchStartURLFromEnv()
	if ok {
		t.Fatalf("expected auto-launch disabled when oauth config is missing, got URL %q", startURL)
	}
}

func TestOAuthAutoLaunchStartURLFromEnv_ReturnsFalseWhenRefreshTokenExists(t *testing.T) {
	setOAuthAutoLaunchEnv(t)
	t.Setenv("DATA_SOURCE", "google")

	tokenFile := filepath.Join(t.TempDir(), "oauth-refresh.json")
	t.Setenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE", tokenFile)

	store := goauth.NewFileRefreshTokenStore(tokenFile)
	if err := store.Save(t.Context(), "already-present"); err != nil {
		t.Fatalf("seed refresh token file: %v", err)
	}

	startURL, ok := oauthAutoLaunchStartURLFromEnv()
	if ok {
		t.Fatalf("expected auto-launch disabled when refresh token exists, got URL %q", startURL)
	}
}

func TestOAuthAutoLaunchStartURLFromEnv_ReturnsURLWhenGoogleAndTokenMissing(t *testing.T) {
	setOAuthAutoLaunchEnv(t)
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
	t.Setenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE", filepath.Join(t.TempDir(), "missing-refresh.json"))

	startURL, ok := oauthAutoLaunchStartURLFromEnv()
	if !ok {
		t.Fatalf("expected auto-launch enabled")
	}
	if startURL != "http://localhost:8080/api/v1/auth/google/start" {
		t.Fatalf("unexpected start URL: %q", startURL)
	}
}

func TestOAuthStartEndpointURLFromEnv_UsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("APP_BASE_URL", "")

	startURL := oauthStartEndpointURLFromEnv()
	if startURL != "http://localhost:8080/api/v1/auth/google/start" {
		t.Fatalf("unexpected default start URL: %q", startURL)
	}
}

func TestOAuthStartEndpointURLFromEnv_TrimsTrailingSlash(t *testing.T) {
	t.Setenv("APP_BASE_URL", "http://127.0.0.1:9999/")

	startURL := oauthStartEndpointURLFromEnv()
	if startURL != "http://127.0.0.1:9999/api/v1/auth/google/start" {
		t.Fatalf("unexpected start URL with trailing slash: %q", startURL)
	}
}

func setOAuthAutoLaunchEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "http://localhost:8080/api/v1/auth/google/callback")
	t.Setenv("GOOGLE_OAUTH_SCOPES", "scope-a,scope-b")
}
