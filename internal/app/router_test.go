package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
