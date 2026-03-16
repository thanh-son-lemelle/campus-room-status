package domain

import (
	"errors"
	"fmt"
)

type InvalidParameterError struct {
	Parameter string
	Value     string
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
func (e *InvalidParameterError) Error() string {
	if e == nil || e.Parameter == "" {
		return "invalid parameters"
	}

	return fmt.Sprintf("invalid parameter: %s", e.Parameter)
}

type RoomNotFoundError struct {
	RoomCode string
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
func (e *ServiceUnavailableError) Error() string {
	if e == nil || e.Provider == UnavailableProviderUnknown {
		return "service unavailable"
	}

	return fmt.Sprintf("%s service unavailable", e.Provider)
}

// NewServiceUnavailableError creates a new service unavailable error.
//
// Summary:
// - Creates a new service unavailable error.
//
// Attributes:
// - provider (UnavailableProvider): Input parameter.
//
// Returns:
// - value1 (*ServiceUnavailableError): Returned value.
func NewServiceUnavailableError(provider UnavailableProvider) *ServiceUnavailableError {
	return &ServiceUnavailableError{Provider: provider}
}

// IsServiceUnavailableFromProvider is service unavailable from provider.
//
// Summary:
// - Is service unavailable from provider.
//
// Attributes:
// - err (error): Input parameter.
// - provider (UnavailableProvider): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
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
