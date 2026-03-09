package buildings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHandler_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/buildings", Handler)

	req := httptest.NewRequest(http.MethodGet, "/buildings", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_ReturnsExactContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/buildings", Handler)

	req := httptest.NewRequest(http.MethodGet, "/buildings", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	if len(payload) != 2 {
		t.Fatalf("expected only 2 top-level fields (timestamp, buildings), got %d", len(payload))
	}

	timestamp, ok := payload["timestamp"]
	if !ok {
		t.Fatalf("expected timestamp field")
	}

	timestampString, ok := timestamp.(string)
	if !ok {
		t.Fatalf("expected timestamp to be a string, got %T", timestamp)
	}

	if _, err := time.Parse(time.RFC3339, timestampString); err != nil {
		t.Fatalf("expected timestamp to be RFC3339, got %q: %v", timestampString, err)
	}

	buildings, ok := payload["buildings"]
	if !ok {
		t.Fatalf("expected buildings field")
	}

	buildingList, ok := buildings.([]any)
	if !ok {
		t.Fatalf("expected buildings to be an array, got %T", buildings)
	}

	if len(buildingList) == 0 {
		t.Fatalf("expected at least one building in fixture")
	}

	for i, item := range buildingList {
		building, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected buildings[%d] to be an object, got %T", i, item)
		}

		if len(building) != 4 {
			t.Fatalf("expected buildings[%d] to have exactly 4 fields, got %d", i, len(building))
		}

		if _, ok := building["id"].(string); !ok {
			t.Fatalf("expected buildings[%d].id to be a string", i)
		}

		if _, ok := building["name"].(string); !ok {
			t.Fatalf("expected buildings[%d].name to be a string", i)
		}

		if _, ok := building["address"].(string); !ok {
			t.Fatalf("expected buildings[%d].address to be a string", i)
		}

		floors, ok := building["floors"].([]any)
		if !ok {
			t.Fatalf("expected buildings[%d].floors to be an array, got %T", i, building["floors"])
		}

		for j, floor := range floors {
			if _, ok := floor.(float64); !ok {
				t.Fatalf("expected buildings[%d].floors[%d] to be numeric, got %T", i, j, floor)
			}
		}
	}
}
