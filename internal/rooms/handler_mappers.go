package rooms

import (
	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
)

func domainRoomToAPIRoom(room domain.Room) api.RoomResponse {
	return api.RoomResponse{
		Code:          room.Code,
		ResourceEmail: room.ResourceEmail,
		Name:          room.Name,
		Building:      room.Building,
		Floor:         room.Floor,
		Capacity:      room.Capacity,
		Type:          room.Type,
		Status:        room.Status,
		CurrentEvent:  domainEventToAPIEvent(room.CurrentEvent),
		NextEvent:     domainEventToAPIEvent(room.NextEvent),
	}
}

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
