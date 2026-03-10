package domain

import (
	"context"
	"time"
)

type RoomService interface {
	ListRooms(ctx context.Context, filters RoomFilters) ([]Room, error)
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
