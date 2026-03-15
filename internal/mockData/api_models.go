package mockdata

import "time"

type APIEvent struct {
	Title     string
	Start     time.Time
	End       time.Time
	Organizer string
}

type APIRoom struct {
	Code         string
	Name         string
	Building     string
	Floor        int
	Capacity     int
	Type         string
	Status       string
	CurrentEvent *APIEvent
	NextEvent    *APIEvent
}

type APIError struct {
	Code      string
	Message   string
	Timestamp time.Time
}

func APIRoomWithNullCurrentEvent() APIRoom {
	return APIRoom{
		Code:         "AMPHI-A",
		Name:         "Amphitheater A",
		Building:     "B1",
		Floor:        1,
		Capacity:     180,
		Type:         "amphitheater",
		Status:       "available",
		CurrentEvent: nil,
		NextEvent: &APIEvent{
			Title:     "Next Session",
			Start:     time.Date(2026, time.March, 10, 14, 0, 0, 0, time.UTC),
			End:       time.Date(2026, time.March, 10, 16, 0, 0, 0, time.UTC),
			Organizer: "Academic Office",
		},
	}
}

func APIErrorRoomNotFound() APIError {
	return APIError{
		Code:      "ROOM_NOT_FOUND",
		Message:   "room AMPHI-X not found",
		Timestamp: time.Date(2026, time.March, 10, 11, 30, 0, 0, time.UTC),
	}
}

func APIRoomsListSingleRoom() []APIRoom {
	room := APIRoomWithNullCurrentEvent()
	room.NextEvent = nil
	return []APIRoom{room}
}
