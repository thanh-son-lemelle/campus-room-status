package rooms

import (
	"net/http"
	"time"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

// NewScheduleHandler godoc
// @Summary Get room schedule for a time period
// @Tags rooms
// @Produce json
// @Param code path string true "Room code"
// @Param start query string true "Start date (YYYY-MM-DD)"
// @Param end query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} api.RoomScheduleResponse
// @Failure 400 {object} api.ErrorEnvelope
// @Failure 404 {object} api.ErrorEnvelope
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/rooms/{code}/schedule [get]
func NewScheduleHandler(service domain.RoomService) gin.HandlerFunc {
	h := &scheduleHandler{
		service: service,
	}

	return h.handle
}

type scheduleHandler struct {
	service domain.RoomService
}

func (h *scheduleHandler) handle(c *gin.Context) {
	if !ensureRoomServiceConfigured(c, h.service) {
		return
	}

	startRaw, endRaw, start, end, err := parseSchedulePeriod(c)
	if err != nil {
		api.WriteError(c, err)
		return
	}

	events, err := h.service.GetRoomSchedule(c.Request.Context(), c.Param("code"), start, end)
	if err != nil {
		writeRoomServiceError(c, err, roomServiceErrorOptions{
			allowInvalidParameter:   true,
			allowRoomNotFound:       true,
			allowServiceUnavailable: true,
		})
		return
	}

	c.JSON(http.StatusOK, api.RoomScheduleResponse{
		RoomCode: c.Param("code"),
		Period: api.PeriodResponse{
			Start: startRaw,
			End:   endRaw,
		},
		Events: mapDomainEventsToAPIEvents(events),
	})
}

func parseSchedulePeriod(c *gin.Context) (string, string, time.Time, time.Time, error) {
	startRaw := c.Query("start")
	if startRaw == "" {
		return "", "", time.Time{}, time.Time{}, &domain.InvalidParameterError{
			Parameter: "start",
		}
	}

	endRaw := c.Query("end")
	if endRaw == "" {
		return "", "", time.Time{}, time.Time{}, &domain.InvalidParameterError{
			Parameter: "end",
		}
	}

	startDate, err := time.Parse(scheduleDateLayout, startRaw)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, &domain.InvalidParameterError{
			Parameter: "start",
			Value:     startRaw,
		}
	}

	endDate, err := time.Parse(scheduleDateLayout, endRaw)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, &domain.InvalidParameterError{
			Parameter: "end",
			Value:     endRaw,
		}
	}

	if startDate.After(endDate) {
		return "", "", time.Time{}, time.Time{}, &domain.InvalidParameterError{
			Parameter: "start",
			Value:     startRaw,
		}
	}

	start := startDate.UTC()
	end := endDate.UTC().Add(24 * time.Hour)
	return startRaw, endRaw, start, end, nil
}
