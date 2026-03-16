package app

import (
	"log"

	goauth "campus-room-status/internal/google/oauth"
)

type routerDependencies struct {
	services  runtimeServices
	oauthFlow *goauth.AuthorizationFlow
}

func bootstrapRouterDependencies() routerDependencies {
	return bootstrapRouterDependenciesWithConfig(loadRuntimeConfigFromEnv())
}

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
