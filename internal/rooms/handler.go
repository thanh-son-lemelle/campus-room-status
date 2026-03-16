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

// ensureRoomServiceConfigured ensures room service configured.
//
// Summary:
// - Ensures room service configured.
//
// Attributes:
// - c (*gin.Context): Input parameter.
// - service (domain.RoomService): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
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

// writeRoomServiceError writes room service error.
//
// Summary:
// - Writes room service error.
//
// Attributes:
// - c (*gin.Context): Input parameter.
// - err (error): Input parameter.
// - opts (roomServiceErrorOptions): Input parameter.
//
// Returns:
// - None.
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

// isInvalidParameterError is invalid parameter error.
//
// Summary:
// - Is invalid parameter error.
//
// Attributes:
// - err (error): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
func isInvalidParameterError(err error) bool {
	var target *domain.InvalidParameterError
	return errors.As(err, &target)
}

// isRoomNotFoundError is room not found error.
//
// Summary:
// - Is room not found error.
//
// Attributes:
// - err (error): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
func isRoomNotFoundError(err error) bool {
	var target *domain.RoomNotFoundError
	return errors.As(err, &target)
}

// isServiceUnavailableError is service unavailable error.
//
// Summary:
// - Is service unavailable error.
//
// Attributes:
// - err (error): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
func isServiceUnavailableError(err error) bool {
	var target *domain.ServiceUnavailableError
	return errors.As(err, &target)
}

type listHandlerClock struct{}

// Now nows function behavior.
//
// Summary:
// - Nows function behavior.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (time.Time): Returned value.
func (listHandlerClock) Now() time.Time {
	return time.Now().UTC()
}
