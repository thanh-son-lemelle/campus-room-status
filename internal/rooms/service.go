package rooms

import (
	"context"
	"errors"
	"time"

	"campus-room-status/internal/domain"
)

type inventoryReader interface {
	GetInventory(ctx context.Context) (domain.InventorySnapshot, error)
}

type eventsReader interface {
	Get(ctx context.Context, key domain.RoomEventsKey) ([]domain.Event, error)
}

type service struct {
	inventory         inventoryReader
	events            eventsReader
	statusInterpreter domain.StatusInterpreter
	clock             domain.Clock
	eventsWindow      time.Duration
}

var _ domain.RoomService = (*service)(nil)

func NewService(
	inventory inventoryReader,
	events eventsReader,
	statusInterpreter domain.StatusInterpreter,
	clock domain.Clock,
) domain.RoomService {
	if clock == nil {
		clock = serviceClock{}
	}
	if statusInterpreter == nil {
		statusInterpreter = domain.NewStatusInterpreter(clock, nil)
	}

	return &service{
		inventory:         inventory,
		events:            events,
		statusInterpreter: statusInterpreter,
		clock:             clock,
		eventsWindow:      24 * time.Hour,
	}
}

func (s *service) ListRooms(ctx context.Context, filters domain.RoomFilters) ([]domain.Room, error) {
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

	now := s.clock.Now()
	start := now.Add(-time.Hour)
	end := now.Add(s.eventsWindow)

	enrichedRooms := make([]domain.Room, 0, len(snapshot.Rooms))
	for _, room := range snapshot.Rooms {
		events, err := s.events.Get(ctx, domain.RoomEventsKey{
			RoomEmail: roomEventLookupKey(room),
			Start:     start,
			End:       end,
		})
		if err != nil {
			return nil, &domain.ServiceUnavailableError{Service: "google"}
		}

		currentEvent, nextEvent := currentAndNextEvent(events, now)

		enriched := cloneDomainRoom(room)
		enriched.CurrentEvent = currentEvent
		enriched.NextEvent = nextEvent
		enriched.Status = s.statusInterpreter.Resolve(
			ctx,
			directoryRoomFromDomainRoom(room),
			events,
		)

		enrichedRooms = append(enrichedRooms, enriched)
	}

	return domain.FilterAndSortRooms(enrichedRooms, filters)
}

func roomEventLookupKey(room domain.Room) string {
	if room.Code != "" {
		return room.Code
	}
	return room.Name
}

func directoryRoomFromDomainRoom(room domain.Room) domain.DirectoryRoom {
	return domain.DirectoryRoom{
		ResourceName:  room.Code,
		ResourceEmail: roomEventLookupKey(room),
		Capacity:      room.Capacity,
		ResourceType:  room.Type,
	}
}

func currentAndNextEvent(events []domain.Event, now time.Time) (*domain.Event, *domain.Event) {
	var current *domain.Event
	var next *domain.Event

	for _, event := range events {
		if !now.Before(event.Start) && now.Before(event.End) {
			if current == nil || event.Start.After(current.Start) {
				e := event
				current = &e
			}
			continue
		}

		if event.Start.After(now) {
			if next == nil || event.Start.Before(next.Start) {
				e := event
				next = &e
			}
		}
	}

	return current, next
}

func cloneDomainRoom(room domain.Room) domain.Room {
	cloned := room
	if room.CurrentEvent != nil {
		current := *room.CurrentEvent
		cloned.CurrentEvent = &current
	}
	if room.NextEvent != nil {
		next := *room.NextEvent
		cloned.NextEvent = &next
	}
	return cloned
}

type serviceClock struct{}

func (serviceClock) Now() time.Time {
	return time.Now().UTC()
}
