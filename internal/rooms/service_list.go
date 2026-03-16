package rooms

import (
	"context"
	"errors"
	"sync"
	"time"

	"campus-room-status/internal/domain"
)

// ListRooms lists rooms.
//
// Summary:
// - Lists rooms.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - filters (domain.RoomFilters): Input parameter.
//
// Returns:
// - value1 ([]domain.Room): Returned value.
// - value2 (error): Returned value.
func (s *service) ListRooms(ctx context.Context, filters domain.RoomFilters) ([]domain.Room, error) {
	if s.inventory == nil {
		return nil, errors.New("inventory cache is required")
	}
	if s.events == nil {
		return nil, errors.New("room events cache is required")
	}
	if err := domain.ValidateRoomFilters(filters); err != nil {
		return nil, err
	}

	snapshot, err := s.inventory.GetInventory(ctx)
	if err != nil {
		return nil, err
	}

	filteredRooms, err := domain.PrefilterRooms(snapshot.Rooms, filters)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	start, end := listRoomsEventsWindow(now, s.eventsWindow, s.eventsBucket)

	enrichedRooms, err := s.enrichRoomsWithCalendar(ctx, filteredRooms, start, end, now)
	if err != nil {
		return nil, domain.NewServiceUnavailableError(domain.UnavailableProviderGoogle)
	}

	return domain.FilterAndSortRooms(enrichedRooms, filters)
}

// enrichRoomsWithCalendar enriches rooms with calendar.
//
// Summary:
// - Enriches rooms with calendar.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - rooms ([]domain.Room): Input parameter.
// - start (time.Time): Input parameter.
// - end (time.Time): Input parameter.
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 ([]domain.Room): Returned value.
// - value2 (error): Returned value.
func (s *service) enrichRoomsWithCalendar(
	ctx context.Context,
	rooms []domain.Room,
	start time.Time,
	end time.Time,
	now time.Time,
) ([]domain.Room, error) {
	if len(rooms) == 0 {
		return []domain.Room{}, nil
	}

	concurrency := s.maxParallelFetch
	if concurrency <= 0 {
		concurrency = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	enriched := make([]domain.Room, len(rooms))
	sem := make(chan struct{}, concurrency)

	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for i, room := range rooms {
		i := i
		room := room

		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			events, err := s.events.Get(ctx, domain.RoomEventsKey{
				RoomEmail: roomEventLookupKey(room),
				Start:     start,
				End:       end,
			})
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
					cancel()
				}
				errMu.Unlock()
				return
			}

			currentEvent, nextEvent := currentAndNextEvent(events, now)

			enrichedRoom := cloneDomainRoom(room)
			enrichedRoom.CurrentEvent = currentEvent
			enrichedRoom.NextEvent = nextEvent
			enrichedRoom.Status = s.statusInterpreter.Resolve(
				ctx,
				directoryRoomFromDomainRoom(room),
				events,
			)

			enriched[i] = enrichedRoom
		}()
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return enriched, nil
}

// listRoomsEventsWindow lists rooms events window.
//
// Summary:
// - Lists rooms events window.
//
// Attributes:
// - now (time.Time): Input parameter.
// - window (time.Duration): Input parameter.
// - bucket (time.Duration): Input parameter.
//
// Returns:
// - value1 (time.Time): Returned value.
// - value2 (time.Time): Returned value.
func listRoomsEventsWindow(now time.Time, window, bucket time.Duration) (time.Time, time.Time) {
	base := now.UTC()
	if bucket > 0 {
		base = base.Truncate(bucket)
	}

	start := base.Add(-time.Hour)
	end := base.Add(window)
	return start, end
}
