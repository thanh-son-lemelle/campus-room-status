package oauth

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
)

var defaultScopes = []string{
	"https://www.googleapis.com/auth/admin.directory.resource.calendar.readonly",
	"https://www.googleapis.com/auth/calendar.readonly",
}

type Config struct {
	ClientID         string
	ClientSecret     string
	RedirectURI      string
	Scopes           []string
	RefreshTokenFile string
	AuthURL          string
	TokenURL         string
}

// LoadConfigFromEnv loads config from env.
//
// Summary:
// - Loads config from env.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Config): Returned value.
// - value2 (error): Returned value.
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		ClientID:         strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID")),
		ClientSecret:     strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")),
		RedirectURI:      strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")),
		Scopes:           parseScopes(strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_SCOPES"))),
		RefreshTokenFile: strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE")),
		AuthURL:          google.Endpoint.AuthURL,
		TokenURL:         google.Endpoint.TokenURL,
	}

	if len(cfg.Scopes) == 0 {
		cfg.Scopes = append([]string(nil), defaultScopes...)
	}
	if cfg.RefreshTokenFile == "" {
		cfg.RefreshTokenFile = ".secrets/google_oauth_refresh_token.json"
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate validates function behavior.
//
// Summary:
// - Validates function behavior.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (error): Returned value.
func (c Config) Validate() error {
	if strings.TrimSpace(c.ClientID) == "" {
		return errors.New("missing GOOGLE_OAUTH_CLIENT_ID")
	}
	if strings.TrimSpace(c.ClientSecret) == "" {
		return errors.New("missing GOOGLE_OAUTH_CLIENT_SECRET")
	}
	if strings.TrimSpace(c.RedirectURI) == "" {
		return errors.New("missing GOOGLE_OAUTH_REDIRECT_URI")
	}
	if len(c.Scopes) == 0 {
		return errors.New("at least one OAuth scope is required")
	}
	if strings.TrimSpace(c.AuthURL) == "" || strings.TrimSpace(c.TokenURL) == "" {
		return errors.New("oauth endpoints are required")
	}

	return nil
}

// parseScopes parses scopes.
//
// Summary:
// - Parses scopes.
//
// Attributes:
// - raw (string): Input parameter.
//
// Returns:
// - value1 ([]string): Returned value.
func parseScopes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	normalized := strings.NewReplacer("\n", ",", "\t", ",", " ", ",").Replace(raw)
	parts := strings.Split(normalized, ",")

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		scope := strings.TrimSpace(part)
		if scope == "" {
			continue
		}
		out = append(out, scope)
	}
	return out
}
