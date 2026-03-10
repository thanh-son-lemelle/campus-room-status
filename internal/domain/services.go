package domain

import "context"

// TODO(issue-10): when implementing the concrete room service, make inventory cache mandatory:
// warmup at startup, serve from cache while TTL is valid, refresh on expiration, and keep stale data when refresh fails.
type RoomService interface {
	ListRooms(ctx context.Context, filters RoomFilters) ([]Room, error)
}

type HealthService interface {
	GetHealth(ctx context.Context) (HealthStatus, error)
}
