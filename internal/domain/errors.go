package domain

import (
	"errors"
	"fmt"
)

type InvalidParameterError struct {
	Parameter string
	Value     string
}

func (e *InvalidParameterError) Error() string {
	if e == nil || e.Parameter == "" {
		return "invalid parameters"
	}

	return fmt.Sprintf("invalid parameter: %s", e.Parameter)
}

type RoomNotFoundError struct {
	RoomCode string
}

func (e *RoomNotFoundError) Error() string {
	if e == nil || e.RoomCode == "" {
		return "room not found"
	}

	return fmt.Sprintf("room %s not found", e.RoomCode)
}

type UnavailableProvider string

const (
	UnavailableProviderUnknown UnavailableProvider = ""
	UnavailableProviderGoogle  UnavailableProvider = "google"
)

type ServiceUnavailableError struct {
	Provider UnavailableProvider
}

func (e *ServiceUnavailableError) Error() string {
	if e == nil || e.Provider == UnavailableProviderUnknown {
		return "service unavailable"
	}

	return fmt.Sprintf("%s service unavailable", e.Provider)
}

func NewServiceUnavailableError(provider UnavailableProvider) *ServiceUnavailableError {
	return &ServiceUnavailableError{Provider: provider}
}

func IsServiceUnavailableFromProvider(err error, provider UnavailableProvider) bool {
	if provider == UnavailableProviderUnknown {
		return false
	}

	var unavailableErr *ServiceUnavailableError
	if !errors.As(err, &unavailableErr) {
		return false
	}

	return unavailableErr.Provider == provider
}
