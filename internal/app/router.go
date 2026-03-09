package app

import (
	"github.com/gin-gonic/gin"

	"campus-room-status/internal/buildings"
	"campus-room-status/internal/health"
	"campus-room-status/internal/rooms"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api/v1")
	api.GET("/buildings", buildings.Handler)
	api.GET("/health", health.Handler)
	api.GET("/rooms", rooms.ListHandler)
	api.GET("/rooms/:code", rooms.DetailHandler)
	api.GET("/rooms/:code/schedule", rooms.ScheduleHandler)

	return r
}
