package app

import (
	"net/http"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"

	"campus-room-status/internal/buildings"
	"campus-room-status/internal/health"
	"campus-room-status/internal/rooms"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Logger(),
		gin.CustomRecovery(func(c *gin.Context, _ any) {
			api.WriteError(c, api.NewHTTPError(
				http.StatusInternalServerError,
				api.ErrorCodeInternalServerError,
				"Une erreur interne est survenue",
			))
		}),
	)

	r.NoRoute(func(c *gin.Context) {
		api.WriteError(c, api.NewHTTPError(
			http.StatusNotFound,
			api.ErrorCodeNotFound,
			"La route demandee n'existe pas",
		))
	})
	r.NoMethod(func(c *gin.Context) {
		api.WriteError(c, api.NewHTTPError(
			http.StatusNotFound,
			api.ErrorCodeNotFound,
			"La route demandee n'existe pas",
		))
	})

	apiGroup := r.Group("/api/v1")
	apiGroup.GET("/buildings", buildings.Handler)
	apiGroup.GET("/health", health.Handler)
	apiGroup.GET("/rooms", rooms.ListHandler)
	apiGroup.GET("/rooms/:code", rooms.DetailHandler)
	apiGroup.GET("/rooms/:code/schedule", rooms.ScheduleHandler)

	return r
}
