package app

import (
	"github.com/gin-gonic/gin"
)

// NewRouter godoc
// @Summary Get Swagger specification
// @Tags docs
// @Produce json
// @Success 200 {string} string "Swagger JSON document"
// @Router /api/v1/docs/openapi.json [get]
func NewRouter() *gin.Engine {
	r := newRouterEngine()
	deps := bootstrapRouterDependencies()
	registerAPIRoutes(r, deps)
	return r
}
