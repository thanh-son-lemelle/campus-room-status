package buildings

import (
	"net/http"
	"time"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"
)

var buildingsFixture = []api.BuildingResponse{
	{
		ID:      "B1",
		Name:    "Building A",
		Address: "1 Campus Street",
		Floors:  []int{0, 1, 2},
	},
}

func Handler(c *gin.Context) {
	c.JSON(http.StatusOK, api.BuildingsResponse{
		Timestamp: time.Now().UTC(),
		Buildings: buildingsFixture,
	})
}
