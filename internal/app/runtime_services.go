package app

import (
	"context"
	"fmt"
	"time"

	"campus-room-status/internal/buildings"
	"campus-room-status/internal/domain"
	"campus-room-status/internal/health"
	"campus-room-status/internal/rooms"
)

type runtimeServices struct {
	buildingService domain.BuildingService
	roomService     domain.RoomService
	healthService   domain.HealthService
	inventoryCache  *domain.InventoryCache
	eventsCache     *domain.RoomEventsCache
}

// newRuntimeServices creates a new runtime services.
//
// Summary:
// - Creates a new runtime services.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (runtimeServices): Returned value.
// - value2 (error): Returned value.
func newRuntimeServices() (runtimeServices, error) {
	return newRuntimeServicesWithConfig(loadRuntimeConfigFromEnv())
}

// newRuntimeServicesWithConfig creates a new runtime services with config.
//
// Summary:
// - Creates a new runtime services with config.
//
// Attributes:
// - cfg (runtimeConfig): Input parameter.
//
// Returns:
// - value1 (runtimeServices): Returned value.
// - value2 (error): Returned value.
func newRuntimeServicesWithConfig(cfg runtimeConfig) (runtimeServices, error) {
	cache, err := domain.NewInventoryCache(
		context.Background(),
		newRuntimeInventorySourceWithConfig(cfg),
		cfg.inventoryCacheTTL,
		nil,
	)
	if err != nil {
		return runtimeServices{}, fmt.Errorf("create inventory cache: %w", err)
	}

	eventsCache, err := domain.NewRoomEventsCache(
		newRuntimeCalendarClientWithConfig(cfg),
		cfg.roomEventsCacheTTL,
		nil,
	)
	if err != nil {
		return runtimeServices{}, fmt.Errorf("create room events cache: %w", err)
	}

	buildingService := buildings.NewService(cache)
	roomService := rooms.NewService(cache, eventsCache, nil, nil)
	healthService := health.NewService(cache, eventsCache, nil, cfg.version)

	return runtimeServices{
		buildingService: buildingService,
		roomService:     roomService,
		healthService:   healthService,
		inventoryCache:  cache,
		eventsCache:     eventsCache,
	}, nil
}

// newUnavailableRuntimeServices creates a new unavailable runtime services.
//
// Summary:
// - Creates a new unavailable runtime services.
//
// Attributes:
// - cause (error): Input parameter.
//
// Returns:
// - value1 (runtimeServices): Returned value.
func newUnavailableRuntimeServices(cause error) runtimeServices {
	return newUnavailableRuntimeServicesWithConfig(cause, loadRuntimeConfigFromEnv())
}

// newUnavailableRuntimeServicesWithConfig creates a new unavailable runtime services with config.
//
// Summary:
// - Creates a new unavailable runtime services with config.
//
// Attributes:
// - cause (error): Input parameter.
// - cfg (runtimeConfig): Input parameter.
//
// Returns:
// - value1 (runtimeServices): Returned value.
func newUnavailableRuntimeServicesWithConfig(cause error, cfg runtimeConfig) runtimeServices {
	serviceErr := runtimeServiceUnavailableError(cause)
	return runtimeServices{
		buildingService: unavailableBuildingService{err: serviceErr},
		roomService:     unavailableRoomService{err: serviceErr},
		// Keep /health available to expose degraded startup state.
		healthService: health.NewService(nil, nil, nil, cfg.version),
	}
}

// refreshCachesAfterOAuth refreshes caches after o auth.
//
// Summary:
// - Refreshes caches after o auth.
//
// Attributes:
// - ctx (context.Context): Input parameter.
//
// Returns:
// - None.
func (s runtimeServices) refreshCachesAfterOAuth(ctx context.Context) {
	if s.inventoryCache != nil {
		_ = s.inventoryCache.ForceRefresh(ctx)
	}
	if s.eventsCache != nil {
		s.eventsCache.Clear()
	}
}

type unavailableBuildingService struct {
	err error
}

// ListBuildings lists buildings.
//
// Summary:
// - Lists buildings.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
//
// Returns:
// - value1 ([]domain.Building): Returned value.
// - value2 (error): Returned value.
func (s unavailableBuildingService) ListBuildings(context.Context) ([]domain.Building, error) {
	return nil, s.err
}

type unavailableRoomService struct {
	err error
}

// ListRooms lists rooms.
//
// Summary:
// - Lists rooms.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
// - arg2 (domain.RoomFilters): Input parameter.
//
// Returns:
// - value1 ([]domain.Room): Returned value.
// - value2 (error): Returned value.
func (s unavailableRoomService) ListRooms(context.Context, domain.RoomFilters) ([]domain.Room, error) {
	return nil, s.err
}

// GetRoomDetail gets room detail.
//
// Summary:
// - Gets room detail.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
// - arg2 (string): Input parameter.
//
// Returns:
// - value1 (domain.Room): Returned value.
// - value2 ([]domain.Event): Returned value.
// - value3 (error): Returned value.
func (s unavailableRoomService) GetRoomDetail(context.Context, string) (domain.Room, []domain.Event, error) {
	return domain.Room{}, nil, s.err
}

// GetRoomSchedule gets room schedule.
//
// Summary:
// - Gets room schedule.
//
// Attributes:
// - arg1 (context.Context): Input parameter.
// - arg2 (string): Input parameter.
// - arg3 (time.Time): Input parameter.
// - arg4 (time.Time): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
// - value2 (error): Returned value.
func (s unavailableRoomService) GetRoomSchedule(context.Context, string, time.Time, time.Time) ([]domain.Event, error) {
	return nil, s.err
}

// runtimeServiceUnavailableError handles runtime service unavailable error.
//
// Summary:
// - Handles runtime service unavailable error.
//
// Attributes:
// - cause (error): Input parameter.
//
// Returns:
// - value1 (error): Returned value.
func runtimeServiceUnavailableError(cause error) error {
	base := domain.NewServiceUnavailableError(domain.UnavailableProviderGoogle)
	if cause == nil {
		return base
	}

	return fmt.Errorf("%w: %v", base, cause)
}
