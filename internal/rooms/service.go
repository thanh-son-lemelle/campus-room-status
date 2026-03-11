package rooms

import (
	"context"
	"errors"
	"sort"
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
		return domain.Room{}, nil, &domain.ServiceUnavailableError{Service: "google"}
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
		return nil, &domain.ServiceUnavailableError{Service: "google"}
	}

	return sortEventsByStart(filterEventsInPeriod(events, start, end)), nil
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

func sortEventsByStart(events []domain.Event) []domain.Event {
	cloned := cloneEvents(events)
	sort.SliceStable(cloned, func(i, j int) bool {
		return cloned[i].Start.Before(cloned[j].Start)
	})
	return cloned
}

func cloneEvents(events []domain.Event) []domain.Event {
	if events == nil {
		return nil
	}

	cloned := make([]domain.Event, len(events))
	copy(cloned, events)
	return cloned
}

func findRoomByCode(rooms []domain.Room, code string) (*domain.Room, error) {
	for i := range rooms {
		if rooms[i].Code == code {
			room := cloneDomainRoom(rooms[i])
			return &room, nil
		}
	}

	return nil, &domain.RoomNotFoundError{RoomCode: code}
}

func filterEventsInPeriod(events []domain.Event, start time.Time, end time.Time) []domain.Event {
	filtered := make([]domain.Event, 0, len(events))
	for _, event := range events {
		if event.End.After(start) && event.Start.Before(end) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

type serviceClock struct{}

func (serviceClock) Now() time.Time {
	return time.Now().UTC()
}
