package app

import (
	"errors"
	"fmt"

	"campus-room-status/internal/domain"
	"campus-room-status/internal/google/adminsdk"
	gcalendar "campus-room-status/internal/google/calendar"
	goauth "campus-room-status/internal/google/oauth"
)

func newRuntimeInventorySource() domain.InventorySource {
	return newRuntimeInventorySourceWithConfig(loadRuntimeConfigFromEnv())
}

func newRuntimeInventorySourceWithConfig(cfg runtimeConfig) domain.InventorySource {
	if cfg.dataSource != runtimeDataSourceGoogle {
		return staticInventorySource{}
	}

	tokenProvider, ok := newRuntimeAdminTokenProviderWithConfig(cfg)
	if !ok {
		return unavailableInventorySource{
			err: errors.New("google inventory source selected but no Google token provider is configured"),
		}
	}

	source, err := adminsdk.NewInventorySource(
		nil,
		tokenProvider,
		adminsdk.InventorySourceConfig{
			BaseURL:  cfg.googleAdminBaseURL,
			Customer: cfg.googleAdminCustomer,
			PageSize: cfg.googleAdminPageSize,
			Timeout:  cfg.googleAdminTimeout,
		},
	)
	if err != nil {
		return unavailableInventorySource{
			err: fmt.Errorf("create Google inventory source: %w", err),
		}
	}

	return oauthBootstrapInventorySource{
		primary: source,
	}
}

func newRuntimeCalendarClient() domain.CalendarClient {
	return newRuntimeCalendarClientWithConfig(loadRuntimeConfigFromEnv())
}

func newRuntimeCalendarClientWithConfig(cfg runtimeConfig) domain.CalendarClient {
	if cfg.dataSource != runtimeDataSourceGoogle {
		return staticCalendarClient{}
	}

	tokenProvider, ok := newRuntimeAdminTokenProviderWithConfig(cfg)
	if !ok {
		return unavailableCalendarClient{
			err: errors.New("google calendar client selected but no Google token provider is configured"),
		}
	}

	client, err := gcalendar.NewClient(
		nil,
		tokenProvider,
		gcalendar.ClientConfig{
			BaseURL:  cfg.googleCalendarBaseURL,
			Timeout:  cfg.googleCalendarTimeout,
			PageSize: cfg.googleCalendarPageSize,
		},
	)
	if err != nil {
		return unavailableCalendarClient{
			err: fmt.Errorf("create Google calendar client: %w", err),
		}
	}

	return client
}

func newRuntimeAdminTokenProvider() (adminsdk.TokenProvider, bool) {
	return newRuntimeAdminTokenProviderWithConfig(loadRuntimeConfigFromEnv())
}

func newRuntimeAdminTokenProviderWithConfig(cfg runtimeConfig) (adminsdk.TokenProvider, bool) {
	if provider, ok := newRuntimeOAuthTokenProvider(); ok {
		return provider, true
	}

	credentialsJSON, hasCredentials := readServiceAccountCredentials()
	if hasCredentials {
		provider, err := adminsdk.NewServiceAccountTokenProvider(adminsdk.ServiceAccountTokenProviderConfig{
			CredentialsJSON: credentialsJSON,
			Subject:         cfg.googleAdminImpersonatedUser,
		})
		if err == nil {
			return provider, true
		}
	}

	token := cfg.googleAdminBearerToken
	if token != "" {
		return staticTokenProvider{token: token}, true
	}

	return nil, false
}

func newRuntimeOAuthTokenProvider() (adminsdk.TokenProvider, bool) {
	cfg, err := goauth.LoadConfigFromEnv()
	if err != nil {
		return nil, false
	}

	// TODO(prod): replace file token store with secret manager or env-backed secure store.
	store := goauth.NewFileRefreshTokenStore(cfg.RefreshTokenFile)

	provider, err := goauth.NewTokenProvider(cfg, store)
	if err != nil {
		return nil, false
	}

	return provider, true
}

func newRuntimeOAuthFlow() *goauth.AuthorizationFlow {
	cfg, err := goauth.LoadConfigFromEnv()
	if err != nil {
		return nil
	}

	// TODO(prod): replace file token store with secret manager or env-backed secure store.
	store := goauth.NewFileRefreshTokenStore(cfg.RefreshTokenFile)
	flow, err := goauth.NewAuthorizationFlow(cfg, store)
	if err != nil {
		return nil
	}

	return flow
}
