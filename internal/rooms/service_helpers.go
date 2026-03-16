package rooms

import (
	"sort"
	"time"

	"campus-room-status/internal/domain"
)

// roomEventLookupKey rooms event lookup key.
//
// Summary:
// - Rooms event lookup key.
//
// Attributes:
// - room (domain.Room): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
func roomEventLookupKey(room domain.Room) string {
	if room.ResourceEmail != "" {
		return room.ResourceEmail
	}
	if room.Code != "" {
		return room.Code
	}
	return room.Name
}

// directoryRoomFromDomainRoom directories room from domain room.
//
// Summary:
// - Directories room from domain room.
//
// Attributes:
// - room (domain.Room): Input parameter.
//
// Returns:
// - value1 (domain.DirectoryRoom): Returned value.
func directoryRoomFromDomainRoom(room domain.Room) domain.DirectoryRoom {
	return domain.DirectoryRoom{
		ResourceName:  room.Code,
		ResourceEmail: roomEventLookupKey(room),
		Capacity:      room.Capacity,
		ResourceType:  room.Type,
	}
}

// currentAndNextEvent currents and next event.
//
// Summary:
// - Currents and next event.
//
// Attributes:
// - events ([]domain.Event): Input parameter.
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 (*domain.Event): Returned value.
// - value2 (*domain.Event): Returned value.
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

// cloneDomainRoom clones domain room.
//
// Summary:
// - Clones domain room.
//
// Attributes:
// - room (domain.Room): Input parameter.
//
// Returns:
// - value1 (domain.Room): Returned value.
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

// sortEventsByStart sorts events by start.
//
// Summary:
// - Sorts events by start.
//
// Attributes:
// - events ([]domain.Event): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
func sortEventsByStart(events []domain.Event) []domain.Event {
	cloned := cloneEvents(events)
	sort.SliceStable(cloned, func(i, j int) bool {
		return cloned[i].Start.Before(cloned[j].Start)
	})
	return cloned
}

// cloneEvents clones events.
//
// Summary:
// - Clones events.
//
// Attributes:
// - events ([]domain.Event): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
func cloneEvents(events []domain.Event) []domain.Event {
	if events == nil {
		return nil
	}

	cloned := make([]domain.Event, len(events))
	copy(cloned, events)
	return cloned
}

// findRoomByCode finds room by code.
//
// Summary:
// - Finds room by code.
//
// Attributes:
// - rooms ([]domain.Room): Input parameter.
// - code (string): Input parameter.
//
// Returns:
// - value1 (*domain.Room): Returned value.
// - value2 (error): Returned value.
func findRoomByCode(rooms []domain.Room, code string) (*domain.Room, error) {
	for i := range rooms {
		if rooms[i].Code == code {
			room := cloneDomainRoom(rooms[i])
			return &room, nil
		}
	}

	return nil, &domain.RoomNotFoundError{RoomCode: code}
}

// filterEventsInPeriod filters events in period.
//
// Summary:
// - Filters events in period.
//
// Attributes:
// - events ([]domain.Event): Input parameter.
// - start (time.Time): Input parameter.
// - end (time.Time): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
func filterEventsInPeriod(events []domain.Event, start time.Time, end time.Time) []domain.Event {
	filtered := make([]domain.Event, 0, len(events))
	for _, event := range events {
		if event.End.After(start) && event.Start.Before(end) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
