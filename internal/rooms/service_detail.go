package rooms

import (
	"context"
	"errors"
	"time"

	"campus-room-status/internal/domain"
)

// GetRoomDetail gets room detail.
//
// Summary:
// - Gets room detail.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - code (string): Input parameter.
//
// Returns:
// - value1 (domain.Room): Returned value.
// - value2 ([]domain.Event): Returned value.
// - value3 (error): Returned value.
func (s *service) GetRoomDetail(ctx context.Context, code string) (domain.Room, []domain.Event, error) {
	if s.inventory == nil {
		return domain.Room{}, nil, errors.New("inventory cache is required")
	}
	if s.events == nil {
		return domain.Room{}, nil, errors.New("room events cache is required")
	}

	snapshot, err := s.inventory.GetInventory(ctx)
	if err != nil {
		return domain.Room{}, nil, err
	}

	found, err := findRoomByCode(snapshot.Rooms, code)
	if err != nil {
		return domain.Room{}, nil, err
	}

	now := s.clock.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	events, err := s.events.Get(ctx, domain.RoomEventsKey{
		RoomEmail: roomEventLookupKey(*found),
		Start:     startOfDay,
		End:       endOfDay,
	})
	if err != nil {
		return domain.Room{}, nil, domain.NewServiceUnavailableError(domain.UnavailableProviderGoogle)
	}

	scheduleToday := sortEventsByStart(events)
	currentEvent, nextEvent := currentAndNextEvent(scheduleToday, now)

	enriched := cloneDomainRoom(*found)
	enriched.CurrentEvent = currentEvent
	enriched.NextEvent = nextEvent
	enriched.Status = s.statusInterpreter.Resolve(
		ctx,
		directoryRoomFromDomainRoom(*found),
		scheduleToday,
	)

	return enriched, scheduleToday, nil
}

// GetRoomSchedule gets room schedule.
//
// Summary:
// - Gets room schedule.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - code (string): Input parameter.
// - start (time.Time): Input parameter.
// - end (time.Time): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
// - value2 (error): Returned value.
func (s *service) GetRoomSchedule(ctx context.Context, code string, start time.Time, end time.Time) ([]domain.Event, error) {
	if s.inventory == nil {
		return nil, errors.New("inventory cache is required")
	}
	if s.events == nil {
		return nil, errors.New("room events cache is required")
	}

	snapshot, err := s.inventory.GetInventory(ctx)
	if err != nil {
		return nil, err
	}

	room, err := findRoomByCode(snapshot.Rooms, code)
	if err != nil {
		return nil, err
	}

	events, err := s.events.Get(ctx, domain.RoomEventsKey{
		RoomEmail: roomEventLookupKey(*room),
		Start:     start,
		End:       end,
	})
	if err != nil {
		return nil, domain.NewServiceUnavailableError(domain.UnavailableProviderGoogle)
	}

	return sortEventsByStart(filterEventsInPeriod(events, start, end)), nil
}
