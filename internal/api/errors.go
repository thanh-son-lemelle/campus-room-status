package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

const (
	ErrorCodeGoogleServiceUnavailable = "GOOGLE_SERVICE_UNAVAILABLE"
	ErrorCodeInternalServerError      = "INTERNAL_SERVER_ERROR"
	ErrorCodeInvalidParameters        = "INVALID_PARAMETERS"
	ErrorCodeNotFound                 = "NOT_FOUND"
	ErrorCodeRoomNotFound             = "ROOM_NOT_FOUND"
	ErrorCodeServiceUnavailable       = "SERVICE_UNAVAILABLE"
)

type HTTPError struct {
	Status  int
	Code    string
	Message string
}

// Error errors function behavior.
//
// Summary:
// - Errors function behavior.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (string): Returned value.
func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// NewHTTPError creates a new http error.
//
// Summary:
// - Creates a new http error.
//
// Attributes:
// - status (int): Input parameter.
// - code (string): Input parameter.
// - message (string): Input parameter.
//
// Returns:
// - value1 (*HTTPError): Returned value.
func NewHTTPError(status int, code string, message string) *HTTPError {
	return &HTTPError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// WriteError writes error.
//
// Summary:
// - Writes error.
//
// Attributes:
// - c (*gin.Context): Input parameter.
// - err (error): Input parameter.
//
// Returns:
// - None.
func WriteError(c *gin.Context, err error) {
	status, payload := mapError(err)
	c.AbortWithStatusJSON(status, payload)
}

// mapError maps error.
//
// Summary:
// - Maps error.
//
// Attributes:
// - err (error): Input parameter.
//
// Returns:
// - value1 (int): Returned value.
// - value2 (ErrorEnvelope): Returned value.
func mapError(err error) (int, ErrorEnvelope) {
	if err == nil {
		return http.StatusInternalServerError, errorEnvelope(ErrorCodeInternalServerError, "Une erreur interne est survenue")
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		status := httpErr.Status
		if status == 0 {
			status = http.StatusInternalServerError
		}

		code := httpErr.Code
		if code == "" {
			code = ErrorCodeInternalServerError
		}

		message := httpErr.Message
		if strings.TrimSpace(message) == "" {
			message = "Une erreur interne est survenue"
		}

		return status, errorEnvelope(code, message)
	}

	var invalidParamErr *domain.InvalidParameterError
	if errors.As(err, &invalidParamErr) {
		message := "Parametres invalides"
		if invalidParamErr.Parameter != "" {
			message = "Le parametre '" + invalidParamErr.Parameter + "' est invalide"
		}
		return http.StatusBadRequest, errorEnvelope(ErrorCodeInvalidParameters, message)
	}

	var roomNotFoundErr *domain.RoomNotFoundError
	if errors.As(err, &roomNotFoundErr) {
		message := "La salle demandee n'existe pas"
		if roomNotFoundErr.RoomCode != "" {
			message = "La salle avec le code '" + roomNotFoundErr.RoomCode + "' n'existe pas"
		}
		return http.StatusNotFound, errorEnvelope(ErrorCodeRoomNotFound, message)
	}

	var serviceUnavailableErr *domain.ServiceUnavailableError
	if errors.As(err, &serviceUnavailableErr) {
		if domain.IsServiceUnavailableFromProvider(err, domain.UnavailableProviderGoogle) {
			return http.StatusServiceUnavailable, errorEnvelope(ErrorCodeGoogleServiceUnavailable, "Service Google indisponible")
		}
		return http.StatusServiceUnavailable, errorEnvelope(ErrorCodeServiceUnavailable, "Service indisponible")
	}

	return http.StatusInternalServerError, errorEnvelope(ErrorCodeInternalServerError, "Une erreur interne est survenue")
}

// errorEnvelope errors envelope.
//
// Summary:
// - Errors envelope.
//
// Attributes:
// - code (string): Input parameter.
// - message (string): Input parameter.
//
// Returns:
// - value1 (ErrorEnvelope): Returned value.
func errorEnvelope(code string, message string) ErrorEnvelope {
	return ErrorEnvelope{
		Error: ErrorResponse{
			Code:      code,
			Message:   message,
			Timestamp: time.Now().UTC(),
		},
	}
}
