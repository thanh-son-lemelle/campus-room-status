package buildings

import (
	"context"
	"testing"
	"time"

	"campus-room-status/internal/domain"
)

func TestService_ReturnsBuildingsFromInventoryCache(t *testing.T) {
	now := time.Date(2026, time.March, 10, 20, 0, 0, 0, time.UTC)
	clock := &serviceClock{now: now}
	source := &serviceInventorySource{
		snapshot: domain.InventorySnapshot{
			Buildings: []domain.Building{
				{
					ID:      "B1",
					Name:    "Building A",
					Address: "1 Campus Street",
					Floors:  []int{0, 1, 2},
				},
			},
		},
	}

	cache, err := domain.NewInventoryCache(context.Background(), source, time.Hour, clock)
	if err != nil {
		t.Fatalf("expected cache warmup to succeed, got error: %v", err)
	}

	service := NewService(cache)

	first, err := service.ListBuildings(context.Background())
	if err != nil {
		t.Fatalf("expected service to return buildings, got error: %v", err)
	}
	second, err := service.ListBuildings(context.Background())
	if err != nil {
		t.Fatalf("expected service to return buildings on second call, got error: %v", err)
	}

	if source.calls != 1 {
		t.Fatalf("expected only one source call thanks to cache, got %d", source.calls)
	}

	if len(first) != 1 || first[0].ID != "B1" {
		t.Fatalf("expected first response to contain building B1, got %+v", first)
	}
	if len(second) != 1 || second[0].ID != "B1" {
		t.Fatalf("expected second response to contain building B1, got %+v", second)
	}
}

type serviceClock struct {
	now time.Time
}

func (c *serviceClock) Now() time.Time {
	return c.now
}

type serviceInventorySource struct {
	snapshot domain.InventorySnapshot
	calls    int
}

func (s *serviceInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	s.calls++
	return s.snapshot, nil
}
