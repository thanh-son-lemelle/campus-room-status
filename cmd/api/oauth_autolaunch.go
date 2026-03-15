package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	goauth "campus-room-status/internal/google/oauth"
)

func maybeAutoLaunchOAuthConsent() {
	startURL, ok := oauthAutoLaunchStartURLFromEnv()
	if !ok {
		return
	}

	// TODO(prod): disable browser auto-launch on server deployments (headless/no local user).
	log.Printf("oauth auto-launch enabled, opening consent flow at %s", startURL)

	go func(url string) {
		// Small delay so the HTTP server has time to start listening.
		time.Sleep(1200 * time.Millisecond)

		if err := openBrowser(url); err != nil {
			log.Printf("oauth auto-launch failed: %v; open manually: %s", err, url)
		}
	}(startURL)
}

func oauthAutoLaunchStartURLFromEnv() (string, bool) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("DATA_SOURCE")), "google") {
		return "", false
	}

	cfg, err := goauth.LoadConfigFromEnv()
	if err != nil {
		log.Printf("oauth auto-launch disabled: invalid oauth config: %v", err)
		return "", false
	}

	store := goauth.NewFileRefreshTokenStore(cfg.RefreshTokenFile)
	refreshToken, err := store.Load(context.Background())
	if err != nil {
		log.Printf("oauth auto-launch disabled: cannot read refresh token store: %v", err)
		return "", false
	}
	if strings.TrimSpace(refreshToken) != "" {
		return "", false
	}

	return oauthStartEndpointURLFromEnv(), true
}

func oauthStartEndpointURLFromEnv() string {
	baseURL := strings.TrimSpace(os.Getenv("APP_BASE_URL"))
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return strings.TrimRight(baseURL, "/") + "/api/v1/auth/google/start"
}

func openBrowser(url string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("browser URL is required")
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
