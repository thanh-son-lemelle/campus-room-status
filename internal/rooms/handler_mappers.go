package rooms

import (
	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
)

// domainRoomToAPIRoom domains room to api room.
//
// Summary:
// - Domains room to api room.
//
// Attributes:
// - room (domain.Room): Input parameter.
//
// Returns:
// - value1 (api.RoomResponse): Returned value.
func domainRoomToAPIRoom(room domain.Room) api.RoomResponse {
	return api.RoomResponse{
		Code:         room.Code,
		Name:         room.Name,
		Building:     room.Building,
		Floor:        room.Floor,
		Capacity:     room.Capacity,
		Type:         room.Type,
		Status:       room.Status,
		CurrentEvent: domainEventToAPIEvent(room.CurrentEvent),
		NextEvent:    domainEventToAPIEvent(room.NextEvent),
	}
}

// domainEventToAPIEvent domains event to api event.
//
// Summary:
// - Domains event to api event.
//
// Attributes:
// - event (*domain.Event): Input parameter.
//
// Returns:
// - value1 (*api.EventResponse): Returned value.
func domainEventToAPIEvent(event *domain.Event) *api.EventResponse {
	if event == nil {
		return nil
	}

	return &api.EventResponse{
		Title:     event.Title,
		Start:     event.Start,
		End:       event.End,
		Organizer: event.Organizer,
	}
}

// mapDomainEventsToAPIEvents maps domain events to api events.
//
// Summary:
// - Maps domain events to api events.
//
// Attributes:
// - events ([]domain.Event): Input parameter.
//
// Returns:
// - value1 ([]api.EventResponse): Returned value.
func mapDomainEventsToAPIEvents(events []domain.Event) []api.EventResponse {
	if events == nil {
		return nil
	}

	out := make([]api.EventResponse, len(events))
	for i := range events {
		out[i] = api.EventResponse{
			Title:     events[i].Title,
			Start:     events[i].Start,
			End:       events[i].End,
			Organizer: events[i].Organizer,
		}
	}

	return out
}
