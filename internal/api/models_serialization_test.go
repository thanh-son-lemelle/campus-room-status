package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRoomResponse_CurrentEventCanBeNullInJSON(t *testing.T) {
	room := RoomResponse{
		Code:         "AMPHI-A",
		Name:         "Amphitheater A",
		Building:     "B1",
		Floor:        1,
		Capacity:     180,
		Type:         "amphitheater",
		Status:       "available",
		CurrentEvent: nil,
		NextEvent: &EventResponse{
			Title:     "Next Session",
			Start:     time.Date(2026, time.March, 10, 14, 0, 0, 0, time.UTC),
			End:       time.Date(2026, time.March, 10, 16, 0, 0, 0, time.UTC),
			Organizer: "Academic Office",
		},
	}

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
		Error: ErrorResponse{
			Code:      "ROOM_NOT_FOUND",
			Message:   "room AMPHI-X not found",
			Timestamp: time.Date(2026, time.March, 10, 11, 30, 0, 0, time.UTC),
		},
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
	payload := RoomsListResponse{
		Timestamp: time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
		Filters: map[string]any{
			"building": "B1",
		},
		Count: 1,
		Rooms: []RoomResponse{
			{
				Code:         "AMPHI-A",
				Name:         "Amphitheater A",
				Building:     "B1",
				Floor:        1,
				Capacity:     180,
				Type:         "amphitheater",
				Status:       "available",
				CurrentEvent: nil,
				NextEvent:    nil,
			},
		},
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
