package api

import (
	"encoding/json"
	"testing"
	"time"

	mockdata "campus-room-status/internal/mockData"
)

func TestRoomResponse_CurrentEventCanBeNullInJSON(t *testing.T) {
	room := roomResponseFromMock(mockdata.APIRoomWithNullCurrentEvent())

	raw, err := json.Marshal(room)
	if err != nil {
		t.Fatalf("expected room JSON marshal to succeed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("expected room JSON unmarshal to succeed: %v", err)
	}

	currentEvent, ok := payload["current_event"]
	if !ok {
		t.Fatalf("expected current_event key in room JSON")
	}
	if currentEvent != nil {
		t.Fatalf("expected current_event to be null, got %T", currentEvent)
	}
}

func TestErrorEnvelope_JSONFormatIsStandard(t *testing.T) {
	errPayload := ErrorEnvelope{
		Error: errorResponseFromMock(mockdata.APIErrorRoomNotFound()),
	}

	raw, err := json.Marshal(errPayload)
	if err != nil {
		t.Fatalf("expected error JSON marshal to succeed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("expected error JSON unmarshal to succeed: %v", err)
	}

	if len(payload) != 1 {
		t.Fatalf("expected top-level error envelope only, got %d fields", len(payload))
	}

	envelope, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected top-level error object, got %T", payload["error"])
	}

	if len(envelope) != 3 {
		t.Fatalf("expected error object with exactly 3 fields, got %d", len(envelope))
	}

	if _, ok := envelope["code"].(string); !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if _, ok := envelope["message"].(string); !ok {
		t.Fatalf("expected error.message to be a string")
	}

	timestamp, ok := envelope["timestamp"].(string)
	if !ok {
		t.Fatalf("expected error.timestamp to be a string")
	}
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		t.Fatalf("expected error.timestamp RFC3339, got %q: %v", timestamp, err)
	}
}

func TestRoomsListResponse_JSONShapeContainsTimestampFiltersCountAndRooms(t *testing.T) {
	mockRooms := mockdata.APIRoomsListSingleRoom()
	rooms := make([]RoomResponse, len(mockRooms))
	for i := range mockRooms {
		rooms[i] = roomResponseFromMock(mockRooms[i])
	}

	payload := RoomsListResponse{
		Timestamp: time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
		Filters: map[string]any{
			"building": "B1",
		},
		Count: 1,
		Rooms: rooms,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("expected rooms list JSON marshal to succeed: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("expected rooms list JSON unmarshal to succeed: %v", err)
	}

	if _, ok := data["timestamp"].(string); !ok {
		t.Fatalf("expected timestamp to be a string")
	}
	if _, ok := data["filters"].(map[string]any); !ok {
		t.Fatalf("expected filters to be an object")
	}
	if _, ok := data["count"].(float64); !ok {
		t.Fatalf("expected count to be numeric")
	}
	if _, ok := data["rooms"].([]any); !ok {
		t.Fatalf("expected rooms to be an array")
	}
}

func roomResponseFromMock(room mockdata.APIRoom) RoomResponse {
	return RoomResponse{
		Code:         room.Code,
		Name:         room.Name,
		Building:     room.Building,
		Floor:        room.Floor,
		Capacity:     room.Capacity,
		Type:         room.Type,
		Status:       room.Status,
		CurrentEvent: eventResponsePtrFromMock(room.CurrentEvent),
		NextEvent:    eventResponsePtrFromMock(room.NextEvent),
	}
}

func eventResponsePtrFromMock(event *mockdata.APIEvent) *EventResponse {
	if event == nil {
		return nil
	}
	return &EventResponse{
		Title:     event.Title,
		Start:     event.Start,
		End:       event.End,
		Organizer: event.Organizer,
	}
}

func errorResponseFromMock(err mockdata.APIError) ErrorResponse {
	return ErrorResponse{
		Code:      err.Code,
		Message:   err.Message,
		Timestamp: err.Timestamp,
	}
}
