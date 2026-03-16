package rooms

import (
	"net/http"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

// NewDetailHandler godoc
// @Summary Get room detail and today's schedule
// @Tags rooms
// @Produce json
// @Param code path string true "Room code"
// @Success 200 {object} api.RoomDetailResponse
// @Failure 404 {object} api.ErrorEnvelope
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/rooms/{code} [get]
func NewDetailHandler(service domain.RoomService) gin.HandlerFunc {
	h := &detailHandler{
		service: service,
	}

	return h.handle
}

type detailHandler struct {
	service domain.RoomService
}

// handle handles function behavior.
//
// Summary:
// - Handles function behavior.
//
// Attributes:
// - c (*gin.Context): Input parameter.
//
// Returns:
// - None.
func (h *detailHandler) handle(c *gin.Context) {
	if !ensureRoomServiceConfigured(c, h.service) {
		return
	}

	room, scheduleToday, err := h.service.GetRoomDetail(c.Request.Context(), c.Param("code"))
	if err != nil {
		writeRoomServiceError(c, err, roomServiceErrorOptions{
			allowRoomNotFound:       true,
			allowServiceUnavailable: true,
		})
		return
	}

	c.JSON(http.StatusOK, api.RoomDetailResponse{
		Code:          room.Code,
		ResourceEmail: room.ResourceEmail,
		Name:          room.Name,
		Building:      room.Building,
		Floor:         room.Floor,
		Capacity:      room.Capacity,
		Type:          room.Type,
		Status:        room.Status,
		CurrentEvent:  domainEventToAPIEvent(room.CurrentEvent),
		NextEvent:     domainEventToAPIEvent(room.NextEvent),
		ScheduleToday: mapDomainEventsToAPIEvents(scheduleToday),
	})
}
