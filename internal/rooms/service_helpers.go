package rooms

import (
	"sort"
	"time"

	"campus-room-status/internal/domain"
)

func roomEventLookupKey(room domain.Room) string {
	if room.ResourceEmail != "" {
		return room.ResourceEmail
	}
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
