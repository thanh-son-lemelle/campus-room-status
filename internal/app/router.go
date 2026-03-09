package app

import (
	"github.com/gin-gonic/gin"

	"campus-room-status/internal/health"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api/v1")
	api.GET("/health", health.Handler)

	return r
}