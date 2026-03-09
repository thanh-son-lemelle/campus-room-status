package health

import (
	"net/http"
	"time"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	start := time.Now()

	c.JSON(http.StatusOK, api.HealthResponse{
		Status:                     "ok",
		Version:                    "dev",
		GoogleAdminAPIConnected:    false,
		GoogleCalendarAPIConnected: false,
		LastSync:                   nil,
		ResponseTimeMS:             time.Since(start).Milliseconds(),
	})
}
