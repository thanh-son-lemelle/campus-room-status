package app

import (
	"net/http"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"
)

// newRouterEngine creates a new router engine.
//
// Summary:
// - Creates a new router engine.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (*gin.Engine): Returned value.
func newRouterEngine() *gin.Engine {
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

	return r
}
