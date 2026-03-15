package buildings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"campus-room-status/internal/domain"
	mockdata "campus-room-status/internal/mockData"
	"github.com/gin-gonic/gin"
)

func TestHandler_ReturnsTimestampAndBuildings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC)
	service := &fakeBuildingService{
		buildings: []domain.Building{domainBuildingFromMock(mockdata.BuildingB1())},
	}

	r := gin.New()
	r.GET("/buildings", NewHandler(service, fixedClock{now: now}))

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

	if timestampString != now.Format(time.RFC3339) {
		t.Fatalf("expected timestamp %q, got %q", now.Format(time.RFC3339), timestampString)
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

func TestHandler_Returns200WhenSourceFailsButCacheHasData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, time.March, 10, 19, 0, 0, 0, time.UTC)
	ttl := time.Hour
	clock := &mutableClock{now: now}
	source := &fakeInventorySource{
		snapshot: domain.InventorySnapshot{
			Buildings: []domain.Building{domainBuildingFromMock(mockdata.BuildingB2())},
		},
	}

	cache, err := domain.NewInventoryCache(context.Background(), source, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache warmup to succeed, got error: %v", err)
	}

	service := NewService(cache)

	r := gin.New()
	r.GET("/buildings", NewHandler(service, clock))

	source.err = errors.New("google directory unavailable")
	clock.Advance(ttl + time.Second)

	req := httptest.NewRequest(http.MethodGet, "/buildings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d with stale cache fallback, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	buildings, ok := payload["buildings"].([]any)
	if !ok {
		t.Fatalf("expected buildings to be an array, got %T", payload["buildings"])
	}
	if len(buildings) == 0 {
		t.Fatalf("expected stale buildings to be returned")
	}
}

type fakeBuildingService struct {
	buildings []domain.Building
	err       error
}

func (s *fakeBuildingService) ListBuildings(context.Context) ([]domain.Building, error) {
	if s.err != nil {
		return nil, s.err
	}

	out := make([]domain.Building, len(s.buildings))
	copy(out, s.buildings)
	return out, nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type mutableClock struct {
	now time.Time
}

func (c *mutableClock) Now() time.Time {
	return c.now
}

func (c *mutableClock) Advance(delta time.Duration) {
	c.now = c.now.Add(delta)
}

type fakeInventorySource struct {
	snapshot domain.InventorySnapshot
	err      error
}

func (s *fakeInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	if s.err != nil {
		return domain.InventorySnapshot{}, s.err
	}
	return s.snapshot, nil
}
