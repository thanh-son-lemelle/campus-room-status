package domain

import "context"

type RoomService interface {
	ListRooms(ctx context.Context, filters RoomFilters) ([]Room, error)
}

type HealthService interface {
	GetHealth(ctx context.Context) (HealthStatus, error)
}
