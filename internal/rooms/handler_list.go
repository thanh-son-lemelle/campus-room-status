package rooms

import (
	"net/http"
	"strconv"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

// NewListHandler godoc
// @Summary List rooms with current and next events
// @Tags rooms
// @Produce json
// @Param building query string false "Building code or ID"
// @Param type query string false "Room type"
// @Param status query string false "Room status (available|occupied|upcoming|maintenance)"
// @Param capacity_min query int false "Minimum capacity"
// @Param capacity_max query int false "Maximum capacity"
// @Param sort query string false "Sort field (name|capacity|status)"
// @Param order query string false "Sort order (asc|desc)"
// @Success 200 {object} api.RoomsListResponse
// @Failure 400 {object} api.ErrorEnvelope
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/rooms [get]
func NewListHandler(service domain.RoomService, clock domain.Clock) gin.HandlerFunc {
	h := &listHandler{
		service: service,
		clock:   clock,
	}
	if h.clock == nil {
		h.clock = listHandlerClock{}
	}

	return h.handle
}

type listHandler struct {
	service domain.RoomService
	clock   domain.Clock
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
func (h *listHandler) handle(c *gin.Context) {
	if !ensureRoomServiceConfigured(c, h.service) {
		return
	}

	queryFilters, responseFilters, err := parseListFilters(c)
	if err != nil {
		api.WriteError(c, err)
		return
	}

	rooms, err := h.service.ListRooms(c.Request.Context(), queryFilters)
	if err != nil {
		writeRoomServiceError(c, err, roomServiceErrorOptions{
			allowInvalidParameter:   true,
			allowServiceUnavailable: true,
		})
		return
	}

	responseRooms := make([]api.RoomResponse, len(rooms))
	for i := range rooms {
		responseRooms[i] = domainRoomToAPIRoom(rooms[i])
	}

	c.JSON(http.StatusOK, api.RoomsListResponse{
		Timestamp: h.clock.Now().UTC(),
		Filters:   responseFilters,
		Count:     len(responseRooms),
		Rooms:     responseRooms,
	})
}

// parseListFilters parses list filters.
//
// Summary:
// - Parses list filters.
//
// Attributes:
// - c (*gin.Context): Input parameter.
//
// Returns:
// - value1 (domain.RoomFilters): Returned value.
// - value2 (map[string]any): Returned value.
// - value3 (error): Returned value.
func parseListFilters(c *gin.Context) (domain.RoomFilters, map[string]any, error) {
	responseFilters := make(map[string]any)
	queryFilters := domain.RoomFilters{}

	if building := c.Query("building"); building != "" {
		queryFilters.Building = &building
		responseFilters["building"] = building
	}

	if roomType := c.Query("type"); roomType != "" {
		queryFilters.Type = &roomType
		responseFilters["type"] = roomType
	}

	if status := c.Query("status"); status != "" {
		queryFilters.Status = &status
		responseFilters["status"] = status
	}

	if raw := c.Query("capacity_min"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return domain.RoomFilters{}, nil, &domain.InvalidParameterError{
				Parameter: "capacity_min",
				Value:     raw,
			}
		}

		queryFilters.CapacityMin = &parsed
		responseFilters["capacity_min"] = parsed
	}

	if raw := c.Query("capacity_max"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return domain.RoomFilters{}, nil, &domain.InvalidParameterError{
				Parameter: "capacity_max",
				Value:     raw,
			}
		}

		queryFilters.CapacityMax = &parsed
		responseFilters["capacity_max"] = parsed
	}

	if sortField := c.Query("sort"); sortField != "" {
		queryFilters.Sort = &sortField
		responseFilters["sort"] = sortField
	}

	if order := c.Query("order"); order != "" {
		queryFilters.Order = &order
		responseFilters["order"] = order
	}

	return queryFilters, responseFilters, nil
}
