package rooms

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestDetailHandler_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/rooms/:code", DetailHandler)

	req := httptest.NewRequest(http.MethodGet, "/rooms/AMPHI-A", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDetailHandler_ReturnsExpectedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/rooms/:code", DetailHandler)

	req := httptest.NewRequest(http.MethodGet, "/rooms/AMPHI-A", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	if len(payload) != 10 {
		t.Fatalf("expected exact room detail contract with 10 fields, got %d", len(payload))
	}

	if value, ok := payload["code"].(string); !ok || value == "" {
		t.Fatalf("expected code to be a non-empty string")
	}
	if _, ok := payload["name"].(string); !ok {
		t.Fatalf("expected name to be a string")
	}
	if _, ok := payload["building"].(string); !ok {
		t.Fatalf("expected building to be a string")
	}
	if _, ok := payload["floor"].(float64); !ok {
		t.Fatalf("expected floor to be numeric, got %T", payload["floor"])
	}
	if _, ok := payload["capacity"].(float64); !ok {
		t.Fatalf("expected capacity to be numeric, got %T", payload["capacity"])
	}
	if _, ok := payload["type"].(string); !ok {
		t.Fatalf("expected type to be a string")
	}
	if _, ok := payload["status"].(string); !ok {
		t.Fatalf("expected status to be a string")
	}

	currentEvent := payload["current_event"]
	if currentEvent != nil {
		event, ok := currentEvent.(map[string]any)
		if !ok {
			t.Fatalf("expected current_event to be an object or null, got %T", currentEvent)
		}
		assertEventContract(t, event, "current_event")
	}

	nextEvent := payload["next_event"]
	if nextEvent != nil {
		event, ok := nextEvent.(map[string]any)
		if !ok {
			t.Fatalf("expected next_event to be an object or null, got %T", nextEvent)
		}
		assertEventContract(t, event, "next_event")
	}

	scheduleToday, ok := payload["schedule_today"].([]any)
	if !ok {
		t.Fatalf("expected schedule_today to be an array, got %T", payload["schedule_today"])
	}

	for i, item := range scheduleToday {
		event, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected schedule_today[%d] to be an object, got %T", i, item)
		}
		assertEventContract(t, event, "schedule_today")
	}
}

func TestScheduleHandler_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/rooms/:code/schedule", ScheduleHandler)

	req := httptest.NewRequest(
		http.MethodGet,
		"/rooms/AMPHI-A/schedule?start=2026-03-09T08:00:00Z&end=2026-03-09T18:00:00Z",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScheduleHandler_ReturnsExpectedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/rooms/:code/schedule", ScheduleHandler)

	start := "2026-03-09T08:00:00Z"
	end := "2026-03-09T18:00:00Z"
	req := httptest.NewRequest(
		http.MethodGet,
		"/rooms/AMPHI-A/schedule?start="+start+"&end="+end,
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	if len(payload) != 3 {
		t.Fatalf("expected exact room schedule contract with 3 fields, got %d", len(payload))
	}

	if value, ok := payload["room_code"].(string); !ok || value == "" {
		t.Fatalf("expected room_code to be a non-empty string")
	}

	period, ok := payload["period"].(map[string]any)
	if !ok {
		t.Fatalf("expected period to be an object, got %T", payload["period"])
	}
	if len(period) != 2 {
		t.Fatalf("expected period to contain only start/end, got %d fields", len(period))
	}

	startValue, ok := period["start"].(string)
	if !ok {
		t.Fatalf("expected period.start to be a string")
	}
	if startValue != start {
		t.Fatalf("expected period.start %q, got %q", start, startValue)
	}
	if _, err := time.Parse(time.RFC3339, startValue); err != nil {
		t.Fatalf("expected period.start to be RFC3339, got %q: %v", startValue, err)
	}

	endValue, ok := period["end"].(string)
	if !ok {
		t.Fatalf("expected period.end to be a string")
	}
	if endValue != end {
		t.Fatalf("expected period.end %q, got %q", end, endValue)
	}
	if _, err := time.Parse(time.RFC3339, endValue); err != nil {
		t.Fatalf("expected period.end to be RFC3339, got %q: %v", endValue, err)
	}

	events, ok := payload["events"].([]any)
	if !ok {
		t.Fatalf("expected events to be an array, got %T", payload["events"])
	}

	for i, item := range events {
		event, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected events[%d] to be an object, got %T", i, item)
		}
		assertEventContract(t, event, "events")
	}
}

func assertEventContract(t *testing.T, event map[string]any, fieldName string) {
	t.Helper()

	if len(event) != 4 {
		t.Fatalf("expected %s event object to have exactly 4 fields, got %d", fieldName, len(event))
	}

	if _, ok := event["title"].(string); !ok {
		t.Fatalf("expected %s.title to be a string", fieldName)
	}

	start, ok := event["start"].(string)
	if !ok {
		t.Fatalf("expected %s.start to be a string", fieldName)
	}
	if _, err := time.Parse(time.RFC3339, start); err != nil {
		t.Fatalf("expected %s.start to be RFC3339, got %q: %v", fieldName, start, err)
	}

	end, ok := event["end"].(string)
	if !ok {
		t.Fatalf("expected %s.end to be a string", fieldName)
	}
	if _, err := time.Parse(time.RFC3339, end); err != nil {
		t.Fatalf("expected %s.end to be RFC3339, got %q: %v", fieldName, end, err)
	}

	if _, ok := event["organizer"].(string); !ok {
		t.Fatalf("expected %s.organizer to be a string", fieldName)
	}
}
