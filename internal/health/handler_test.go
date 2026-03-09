package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", Handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHealthHandler_ReturnsExpectedContractFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", Handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
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
			t.Errorf("expected key %q in /health response", key)
		}
	}
}

func TestHealthHandler_ReturnsExpectedContractTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", Handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	if _, ok := payload["status"].(string); !ok {
		t.Fatalf("expected status to be a string, got %T", payload["status"])
	}

	if _, ok := payload["version"].(string); !ok {
		t.Fatalf("expected version to be a string, got %T", payload["version"])
	}

	if _, ok := payload["google_admin_api_connected"].(bool); !ok {
		t.Fatalf("expected google_admin_api_connected to be a boolean, got %T", payload["google_admin_api_connected"])
	}

	if _, ok := payload["google_calendar_api_connected"].(bool); !ok {
		t.Fatalf("expected google_calendar_api_connected to be a boolean, got %T", payload["google_calendar_api_connected"])
	}

	lastSync := payload["last_sync"]
	if lastSync != nil {
		if _, ok := lastSync.(string); !ok {
			t.Fatalf("expected last_sync to be null or string, got %T", lastSync)
		}
	}

	if _, ok := payload["response_time_ms"].(float64); !ok {
		t.Fatalf("expected response_time_ms to be a number, got %T", payload["response_time_ms"])
	}
}
