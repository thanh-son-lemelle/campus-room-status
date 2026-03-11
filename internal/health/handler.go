package health

import (
	"net/http"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

var defaultHandler = NewHandler(NewService(nil, nil, nil, "dev"))

func Handler(c *gin.Context) {
	defaultHandler(c)
}

func NewHandler(service domain.HealthService) gin.HandlerFunc {
	h := &handler{
		service: service,
	}

	return h.handle
}

type handler struct {
	service domain.HealthService
}

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
