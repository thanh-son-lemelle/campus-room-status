package domain

import (
	"context"
	"time"
)

// TODO(issue-10): when implementing the concrete room service, make inventory cache mandatory:
// warmup at startup, serve from cache while TTL is valid, refresh on expiration, and keep stale data when refresh fails.
type RoomService interface {
	ListRooms(ctx context.Context, filters RoomFilters) ([]Room, error)
	GetRoomDetail(ctx context.Context, code string) (Room, []Event, error)
	GetRoomSchedule(ctx context.Context, code string, start time.Time, end time.Time) ([]Event, error)
}

type BuildingService interface {
	ListBuildings(ctx context.Context) ([]Building, error)
}

type HealthService interface {
	GetHealth(ctx context.Context) (HealthStatus, error)
}

type AdminDirectoryClient interface {
	ListRooms(ctx context.Context) ([]DirectoryRoom, error)
}

type CalendarClient interface {
	ListRoomEvents(ctx context.Context, resourceEmail string, start time.Time, end time.Time) ([]Event, error)
}

type StatusInterpreter interface {
	Resolve(ctx context.Context, room DirectoryRoom, events []Event) string
}

type Clock interface {
	Now() time.Time
}

// UnavailabilitySource is an optional abstraction for external closure/outage signals.
// Implement only when a complementary source exists in the project context.
type UnavailabilitySource interface {
	IsRoomUnavailable(ctx context.Context, roomCode string, at time.Time) (bool, error)
}
