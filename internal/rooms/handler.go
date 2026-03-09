package rooms

import (
	"net/http"
	"time"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"
)

var scheduleFixture = []api.EventResponse{
	{
		Title:     "Advanced Networks",
		Start:     time.Date(2026, time.March, 9, 9, 0, 0, 0, time.UTC),
		End:       time.Date(2026, time.March, 9, 11, 0, 0, 0, time.UTC),
		Organizer: "IT Department",
	},
	{
		Title:     "Distributed Systems",
		Start:     time.Date(2026, time.March, 9, 14, 0, 0, 0, time.UTC),
		End:       time.Date(2026, time.March, 9, 16, 0, 0, 0, time.UTC),
		Organizer: "Engineering Office",
	},
}

var nextEventFixture = &api.EventResponse{
	Title:     "Capstone Review",
	Start:     time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC),
	End:       time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
	Organizer: "Academic Board",
}

func DetailHandler(c *gin.Context) {
	c.JSON(http.StatusOK, api.RoomDetailResponse{
		Code:          c.Param("code"),
		Name:          "Amphitheater A",
		Building:      "B1",
		Floor:         1,
		Capacity:      180,
		Type:          "amphitheater",
		Status:        "available",
		CurrentEvent:  nil,
		NextEvent:     nextEventFixture,
		ScheduleToday: scheduleFixture,
	})
}

func ScheduleHandler(c *gin.Context) {
	c.JSON(http.StatusOK, api.RoomScheduleResponse{
		RoomCode: c.Param("code"),
		Period: api.PeriodResponse{
			Start: c.Query("start"),
			End:   c.Query("end"),
		},
		Events: scheduleFixture,
	})
}
