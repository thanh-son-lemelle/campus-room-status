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

// APIRoomWithNullCurrentEvent apis room with null current event.
//
// Summary:
// - Apis room with null current event.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (APIRoom): Returned value.
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

// APIErrorRoomNotFound apis error room not found.
//
// Summary:
// - Apis error room not found.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (APIError): Returned value.
func APIErrorRoomNotFound() APIError {
	return APIError{
		Code:      "ROOM_NOT_FOUND",
		Message:   "room AMPHI-X not found",
		Timestamp: time.Date(2026, time.March, 10, 11, 30, 0, 0, time.UTC),
	}
}

// APIRoomsListSingleRoom apis rooms list single room.
//
// Summary:
// - Apis rooms list single room.
//
// Attributes:
// - None.
//
// Returns:
// - value1 ([]APIRoom): Returned value.
func APIRoomsListSingleRoom() []APIRoom {
	room := APIRoomWithNullCurrentEvent()
	room.NextEvent = nil
	return []APIRoom{room}
}
