package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewRouter_ExposesHealthAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/health to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_ExposesBuildingsAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/buildings", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/buildings to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_ExposesRoomDetailAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/AMPHI-A", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/rooms/AMPHI-A to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_ExposesRoomsAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/rooms to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_ExposesRoomScheduleAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/AMPHI-A/schedule?start=2026-03-09T08:00:00Z&end=2026-03-09T18:00:00Z",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/rooms/AMPHI-A/schedule to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_HealthResponseMatchesContract(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/health to return %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	requiredKeys := []string{
		"status",
		"version",
		"google_admin_api_connected",
		"google_calendar_api_connected",
		"last_sync",
		"response_time_ms",
	}

	for _, key := range requiredKeys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in /api/v1/health response", key)
		}
	}
}

func TestNewRouter_Error400UsesStandardFormat(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms?capacity_min=not-a-number", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "INVALID_PARAMETERS" {
		t.Fatalf("expected error.code %q, got %q", "INVALID_PARAMETERS", code)
	}
}

func TestNewRouter_Error404UsesStandardFormat(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown-endpoint", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	assertStandardErrorResponse(t, w.Body.Bytes())
}

func TestNewRouter_Error503UsesStandardFormat(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/SVC-UNAVAILABLE/schedule?start=2026-03-09T08:00:00Z&end=2026-03-09T18:00:00Z",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "GOOGLE_SERVICE_UNAVAILABLE" {
		t.Fatalf("expected error.code %q, got %q", "GOOGLE_SERVICE_UNAVAILABLE", code)
	}
}

func TestNewRouter_ErrorResponseIncludesTimestamp(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms?capacity_max=invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())
	if _, ok := errObj["timestamp"]; !ok {
		t.Fatalf("expected error.timestamp to be present")
	}
}

func TestNewRouter_RoomNotFoundReturnsRoomNotFoundCode(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/NOPE-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "ROOM_NOT_FOUND" {
		t.Fatalf("expected error.code %q, got %q", "ROOM_NOT_FOUND", code)
	}

	message, ok := errObj["message"].(string)
	if !ok {
		t.Fatalf("expected error.message to be a string")
	}
	if !strings.Contains(message, "NOPE-404") {
		t.Fatalf("expected error.message to contain missing room code, got %q", message)
	}
}

func TestNewRouter_RecoveryUsesStandardErrorFormat(t *testing.T) {
	r := NewRouter()
	r.GET("/api/v1/panic", func(_ *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("expected error.code %q, got %q", "INTERNAL_SERVER_ERROR", code)
	}

	message, ok := errObj["message"].(string)
	if !ok {
		t.Fatalf("expected error.message to be a string")
	}
	if strings.Contains(strings.ToLower(message), "panic") {
		t.Fatalf("expected panic details to stay hidden, got %q", message)
	}
}

func assertStandardErrorResponse(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	if len(payload) != 1 {
		t.Fatalf("expected only error envelope at top level, got %d fields", len(payload))
	}

	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected payload.error to be an object, got %T", payload["error"])
	}

	if len(errObj) != 3 {
		t.Fatalf("expected payload.error to contain exactly 3 fields, got %d", len(errObj))
	}

	if _, ok := errObj["code"].(string); !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if _, ok := errObj["message"].(string); !ok {
		t.Fatalf("expected error.message to be a string")
	}

	timestamp, ok := errObj["timestamp"].(string)
	if !ok {
		t.Fatalf("expected error.timestamp to be a string")
	}
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		t.Fatalf("expected error.timestamp RFC3339, got %q: %v", timestamp, err)
	}

	return errObj
}
