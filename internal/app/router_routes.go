package app

import (
	"campus-room-status/internal/buildings"
	"campus-room-status/internal/docs"
	goauth "campus-room-status/internal/google/oauth"
	"campus-room-status/internal/health"
	"campus-room-status/internal/rooms"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// registerAPIRoutes registers api routes.
//
// Summary:
// - Registers api routes.
//
// Attributes:
// - r (*gin.Engine): Input parameter.
// - deps (routerDependencies): Input parameter.
//
// Returns:
// - None.
func registerAPIRoutes(r *gin.Engine, deps routerDependencies) {
	apiGroup := r.Group("/api/v1")
	apiGroup.GET("/docs/openapi.json", docs.NewOpenAPIHandler())
	apiGroup.GET("/docs/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/api/v1/docs/openapi.json"),
	))

	// TODO(prod): protect or disable OAuth consent endpoints outside trusted admin network.
	apiGroup.GET("/auth/google/start", goauth.NewStartHandler(deps.oauthFlow))
	apiGroup.GET("/auth/google/callback", goauth.NewCallbackHandlerWithHook(deps.oauthFlow, deps.services.refreshCachesAfterOAuth))

	apiGroup.GET("/buildings", buildings.NewHandler(deps.services.buildingService, nil))
	apiGroup.GET("/health", health.NewHandler(deps.services.healthService))
	apiGroup.GET("/rooms", rooms.NewListHandler(deps.services.roomService, nil))
	apiGroup.GET("/rooms/:code", rooms.NewDetailHandler(deps.services.roomService))
	apiGroup.GET("/rooms/:code/schedule", rooms.NewScheduleHandler(deps.services.roomService))
}
