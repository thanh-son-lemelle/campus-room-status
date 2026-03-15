package buildings

import (
	"net/http"
	"time"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

// NewHandler godoc
// @Summary List campus buildings
// @Tags buildings
// @Produce json
// @Success 200 {object} api.BuildingsResponse
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/buildings [get]
func NewHandler(service domain.BuildingService, clock domain.Clock) gin.HandlerFunc {
	h := &handler{
		service: service,
		clock:   clock,
	}
	if h.clock == nil {
		h.clock = handlerClock{}
	}

	return h.handle
}

type handler struct {
	service domain.BuildingService
	clock   domain.Clock
}

func (h *handler) handle(c *gin.Context) {
	if h.service == nil {
		api.WriteError(c, api.NewHTTPError(
			http.StatusInternalServerError,
			api.ErrorCodeInternalServerError,
			"Building service is not configured",
		))
		return
	}

	buildings, err := h.service.ListBuildings(c.Request.Context())
	if err != nil {
		api.WriteError(c, api.NewHTTPError(
			http.StatusServiceUnavailable,
			api.ErrorCodeServiceUnavailable,
			"Service Google indisponible",
		))
		return
	}

	c.JSON(http.StatusOK, api.BuildingsResponse{
		Timestamp: timestampFromClock(h.clock),
		Buildings: mapDomainBuildingsToResponse(buildings),
	})
}

func timestampFromClock(clock domain.Clock) time.Time {
	if clock == nil {
		return time.Now().UTC()
	}
	return clock.Now().UTC()
}

func mapDomainBuildingsToResponse(buildings []domain.Building) []api.BuildingResponse {
	out := make([]api.BuildingResponse, len(buildings))
	for i := range buildings {
		out[i] = api.BuildingResponse{
			ID:      buildings[i].ID,
			Name:    buildings[i].Name,
			Address: buildings[i].Address,
			Floors:  append([]string(nil), buildings[i].Floors...),
		}
	}

	return out
}

type handlerClock struct{}

func (handlerClock) Now() time.Time {
	return time.Now().UTC()
}
