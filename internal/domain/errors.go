package domain

import "fmt"

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

type ServiceUnavailableError struct {
	Service string
}

func (e *ServiceUnavailableError) Error() string {
	if e == nil || e.Service == "" {
		return "service unavailable"
	}

	return fmt.Sprintf("%s service unavailable", e.Service)
}
