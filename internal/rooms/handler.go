package rooms

import (
	"errors"
	"net/http"
	"time"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

const scheduleDateLayout = "2006-01-02"

type roomServiceErrorOptions struct {
	allowInvalidParameter   bool
	allowRoomNotFound       bool
	allowServiceUnavailable bool
}

func ensureRoomServiceConfigured(c *gin.Context, service domain.RoomService) bool {
	if service != nil {
		return true
	}

	api.WriteError(c, api.NewHTTPError(
		http.StatusInternalServerError,
		api.ErrorCodeInternalServerError,
		"Room service is not configured",
	))
	return false
}

func writeRoomServiceError(c *gin.Context, err error, opts roomServiceErrorOptions) {
	if err == nil {
		return
	}
	if opts.allowInvalidParameter && isInvalidParameterError(err) {
		api.WriteError(c, err)
		return
	}
	if opts.allowRoomNotFound && isRoomNotFoundError(err) {
		api.WriteError(c, err)
		return
	}
	if opts.allowServiceUnavailable && isServiceUnavailableError(err) {
		api.WriteError(c, err)
		return
	}

	api.WriteError(c, api.NewHTTPError(
		http.StatusInternalServerError,
		api.ErrorCodeInternalServerError,
		"Une erreur interne est survenue",
	))
}

func isInvalidParameterError(err error) bool {
	var target *domain.InvalidParameterError
	return errors.As(err, &target)
}

func isRoomNotFoundError(err error) bool {
	var target *domain.RoomNotFoundError
	return errors.As(err, &target)
}

func isServiceUnavailableError(err error) bool {
	var target *domain.ServiceUnavailableError
	return errors.As(err, &target)
}

type listHandlerClock struct{}

func (listHandlerClock) Now() time.Time {
	return time.Now().UTC()
}
