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

func newRuntimeServices() (runtimeServices, error) {
	return newRuntimeServicesWithConfig(loadRuntimeConfigFromEnv())
}

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

func newUnavailableRuntimeServices(cause error) runtimeServices {
	return newUnavailableRuntimeServicesWithConfig(cause, loadRuntimeConfigFromEnv())
}

func newUnavailableRuntimeServicesWithConfig(cause error, cfg runtimeConfig) runtimeServices {
	serviceErr := runtimeServiceUnavailableError(cause)
	return runtimeServices{
		buildingService: unavailableBuildingService{err: serviceErr},
		roomService:     unavailableRoomService{err: serviceErr},
		// Keep /health available to expose degraded startup state.
		healthService: health.NewService(nil, nil, nil, cfg.version),
	}
}

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

func (s unavailableBuildingService) ListBuildings(context.Context) ([]domain.Building, error) {
	return nil, s.err
}

type unavailableRoomService struct {
	err error
}

func (s unavailableRoomService) ListRooms(context.Context, domain.RoomFilters) ([]domain.Room, error) {
	return nil, s.err
}

func (s unavailableRoomService) GetRoomDetail(context.Context, string) (domain.Room, []domain.Event, error) {
	return domain.Room{}, nil, s.err
}

func (s unavailableRoomService) GetRoomSchedule(context.Context, string, time.Time, time.Time) ([]domain.Event, error) {
	return nil, s.err
}

func runtimeServiceUnavailableError(cause error) error {
	base := domain.NewServiceUnavailableError(domain.UnavailableProviderGoogle)
	if cause == nil {
		return base
	}

	return fmt.Errorf("%w: %v", base, cause)
}
