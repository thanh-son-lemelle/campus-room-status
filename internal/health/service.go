package health

import (
	"context"
	"strings"
	"time"

	"campus-room-status/internal/domain"
)

type inventoryHealthReader interface {
	HealthState() domain.InventoryCacheHealthState
}

type calendarHealthReader interface {
	HealthState() domain.RoomEventsCacheHealthState
}

type service struct {
	inventory inventoryHealthReader
	calendar  calendarHealthReader
	clock     domain.Clock
	version   string
}

var _ domain.HealthService = (*service)(nil)

// NewService creates a new service.
//
// Summary:
// - Creates a new service.
//
// Attributes:
// - inventory (inventoryHealthReader): Input parameter.
// - calendar (calendarHealthReader): Input parameter.
// - clock (domain.Clock): Input parameter.
// - version (string): Input parameter.
//
// Returns:
// - value1 (domain.HealthService): Returned value.
func NewService(
	inventory inventoryHealthReader,
	calendar calendarHealthReader,
	clock domain.Clock,
	version string,
) domain.HealthService {
	if clock == nil {
		clock = serviceClock{}
	}

	trimmedVersion := strings.TrimSpace(version)
	if trimmedVersion == "" {
		trimmedVersion = "dev"
	}

	return &service{
		inventory: inventory,
		calendar:  calendar,
		clock:     clock,
		version:   trimmedVersion,
	}
}

// GetHealth gets health.
//
// Summary:
// - Gets health.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
//
// Returns:
// - value1 (domain.HealthStatus): Returned value.
// - value2 (error): Returned value.
func (s *service) GetHealth(context.Context) (domain.HealthStatus, error) {
	start := s.clock.Now().UTC()

	var inventoryState domain.InventoryCacheHealthState
	hasInventory := s.inventory != nil
	if hasInventory {
		inventoryState = s.inventory.HealthState()
	}

	var calendarState domain.RoomEventsCacheHealthState
	hasCalendar := s.calendar != nil
	if hasCalendar {
		calendarState = s.calendar.HealthState()
	}

	adminConnected := hasInventory && !inventoryState.Degraded
	calendarConnected := hasCalendar && !calendarState.Degraded

	overallStatus := "ok"
	if !adminConnected || !calendarConnected {
		overallStatus = "degraded"
	}

	lastSync := mostRecentSync(inventoryState.LastRefresh, calendarState.LastSuccessfulRefreshAt)

	responseTimeMS := s.clock.Now().UTC().Sub(start).Milliseconds()
	if responseTimeMS <= 0 {
		responseTimeMS = 1
	}

	return domain.HealthStatus{
		Status:                     overallStatus,
		Version:                    s.version,
		GoogleAdminAPIConnected:    adminConnected,
		GoogleCalendarAPIConnected: calendarConnected,
		LastSync:                   lastSync,
		ResponseTimeMS:             responseTimeMS,
	}, nil
}

// mostRecentSync mosts recent sync.
//
// Summary:
// - Mosts recent sync.
//
// Attributes:
// - inventoryRefresh (time.Time): Input parameter.
// - calendarRefresh (*time.Time): Input parameter.
//
// Returns:
// - value1 (*time.Time): Returned value.
func mostRecentSync(inventoryRefresh time.Time, calendarRefresh *time.Time) *time.Time {
	if inventoryRefresh.IsZero() && calendarRefresh == nil {
		return nil
	}

	if calendarRefresh == nil {
		value := inventoryRefresh.UTC()
		return &value
	}

	if inventoryRefresh.IsZero() || calendarRefresh.After(inventoryRefresh) {
		value := calendarRefresh.UTC()
		return &value
	}

	value := inventoryRefresh.UTC()
	return &value
}

type serviceClock struct{}

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
func (serviceClock) Now() time.Time {
	return time.Now().UTC()
}
