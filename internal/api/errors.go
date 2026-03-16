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

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func NewHTTPError(status int, code string, message string) *HTTPError {
	return &HTTPError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func WriteError(c *gin.Context, err error) {
	status, payload := mapError(err)
	c.AbortWithStatusJSON(status, payload)
}

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

func errorEnvelope(code string, message string) ErrorEnvelope {
	return ErrorEnvelope{
		Error: ErrorResponse{
			Code:      code,
			Message:   message,
			Timestamp: time.Now().UTC(),
		},
	}
}
