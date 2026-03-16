package health

import (
	"net/http"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

var defaultHandler = NewHandler(NewService(nil, nil, nil, "dev"))

// Handler handlers function behavior.
//
// Summary:
// - Handlers function behavior.
//
// Attributes:
// - c (*gin.Context): Input parameter.
//
// Returns:
// - None.
func Handler(c *gin.Context) {
	defaultHandler(c)
}

// NewHandler godoc
// @Summary Get API health status
// @Tags health
// @Produce json
// @Success 200 {object} api.HealthResponse
// @Router /api/v1/health [get]
func NewHandler(service domain.HealthService) gin.HandlerFunc {
	h := &handler{
		service: service,
	}

	return h.handle
}

type handler struct {
	service domain.HealthService
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
func (h *handler) handle(c *gin.Context) {
	if h.service == nil {
		api.WriteError(c, api.NewHTTPError(
			http.StatusInternalServerError,
			api.ErrorCodeInternalServerError,
			"Health service is not configured",
		))
		return
	}

	status, err := h.service.GetHealth(c.Request.Context())
	if err != nil {
		api.WriteError(c, api.NewHTTPError(
			http.StatusInternalServerError,
			api.ErrorCodeInternalServerError,
			"Une erreur interne est survenue",
		))
		return
	}

	c.JSON(http.StatusOK, api.HealthResponse{
		Status:                     status.Status,
		Version:                    status.Version,
		GoogleAdminAPIConnected:    status.GoogleAdminAPIConnected,
		GoogleCalendarAPIConnected: status.GoogleCalendarAPIConnected,
		LastSync:                   status.LastSync,
		ResponseTimeMS:             status.ResponseTimeMS,
	})
}
