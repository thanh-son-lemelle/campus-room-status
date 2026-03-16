package app

import (
	"log"

	goauth "campus-room-status/internal/google/oauth"
)

type routerDependencies struct {
	services  runtimeServices
	oauthFlow *goauth.AuthorizationFlow
}

// bootstrapRouterDependencies initializes router dependencies from environment configuration.
//
// Summary:
// - Loads runtime configuration from environment variables and bootstraps router dependencies.
//
// Attributes:
// - None.
//
// Returns:
// - routerDependencies: Services and OAuth flow wiring for router setup.
func bootstrapRouterDependencies() routerDependencies {
	return bootstrapRouterDependenciesWithConfig(loadRuntimeConfigFromEnv())
}

// bootstrapRouterDependenciesWithConfig initializes router dependencies from an explicit config.
//
// Summary:
// - Builds runtime services from the provided configuration.
// - Falls back to degraded services when service bootstrap fails.
//
// Attributes:
// - cfg: Runtime configuration used to initialize services.
//
// Returns:
// - routerDependencies: Services and OAuth flow wiring for router setup.
func bootstrapRouterDependenciesWithConfig(cfg runtimeConfig) routerDependencies {
	services, err := newRuntimeServicesWithConfig(cfg)
	if err != nil {
		log.Printf("runtime bootstrap failed; starting in degraded mode: %v", err)
		services = newUnavailableRuntimeServicesWithConfig(err, cfg)
	}

	return routerDependencies{
		services:  services,
		oauthFlow: newRuntimeOAuthFlow(),
	}
}
