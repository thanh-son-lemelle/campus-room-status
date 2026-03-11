package domain

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestInventoryCache_WarmupAtStartup(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	ttl := time.Hour
	clock := &fakeClock{now: now}
	source := &fakeInventorySource{
		snapshots: []InventorySnapshot{
			snapshotFixture("B1", "AMPHI-A"),
		},
	}

	cache, err := NewInventoryCache(context.Background(), source, ttl, clock)
	if err != nil {
		t.Fatalf("expected warmup to succeed, got error: %v", err)
	}

	if source.Calls() != 1 {
		t.Fatalf("expected exactly one source call during warmup, got %d", source.Calls())
	}

	meta := cache.Metadata()
	if !meta.LastRefresh.Equal(now) {
		t.Fatalf("expected lastRefresh %s, got %s", now, meta.LastRefresh)
	}
	if !meta.ExpiresAt.Equal(now.Add(ttl)) {
		t.Fatalf("expected expiresAt %s, got %s", now.Add(ttl), meta.ExpiresAt)
	}
}

func TestInventoryCache_HitBeforeExpiration(t *testing.T) {
	now := time.Date(2026, time.March, 10, 11, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: now}
	source := &fakeInventorySource{
		snapshots: []InventorySnapshot{
			snapshotFixture("B1", "AMPHI-A"),
		},
	}

	cache, err := NewInventoryCache(context.Background(), source, time.Hour, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	clock.Advance(30 * time.Minute)
	got, err := cache.GetInventory(context.Background())
	if err != nil {
		t.Fatalf("expected cache hit to succeed, got error: %v", err)
	}

	if source.Calls() != 1 {
		t.Fatalf("expected no additional source call before expiration, got %d", source.Calls())
	}

	if len(got.Rooms) != 1 || got.Rooms[0].Code != "AMPHI-A" {
		t.Fatalf("expected cached room AMPHI-A, got %+v", got.Rooms)
	}
}

func TestInventoryCache_StaleDataIsServedWhenSourceFailsAfterExpiration(t *testing.T) {
	now := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	ttl := time.Hour
	clock := &fakeClock{now: now}
	source := &fakeInventorySource{
		snapshots: []InventorySnapshot{
			snapshotFixture("B1", "AMPHI-A"),
		},
	}

	cache, err := NewInventoryCache(context.Background(), source, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	source.SetError(errors.New("upstream unavailable"))
	clock.Advance(2 * time.Hour)

	got, err := cache.GetInventory(context.Background())
	if err != nil {
		t.Fatalf("expected stale data to be returned when refresh fails, got error: %v", err)
	}

	if source.Calls() != 2 {
		t.Fatalf("expected refresh attempt after expiration, got %d source calls", source.Calls())
	}

	if len(got.Rooms) != 1 || got.Rooms[0].Code != "AMPHI-A" {
		t.Fatalf("expected stale room AMPHI-A, got %+v", got.Rooms)
	}

	meta := cache.Metadata()
	if !meta.LastRefresh.Equal(now) {
		t.Fatalf("expected lastRefresh to stay on successful refresh timestamp %s, got %s", now, meta.LastRefresh)
	}
	if !meta.ExpiresAt.Equal(now.Add(ttl)) {
		t.Fatalf("expected expiresAt to stay unchanged after failed refresh, got %s", meta.ExpiresAt)
	}
}

func TestInventoryCache_RefreshOnExpiration(t *testing.T) {
	now := time.Date(2026, time.March, 10, 13, 0, 0, 0, time.UTC)
	ttl := time.Hour
	clock := &fakeClock{now: now}
	source := &fakeInventorySource{
		snapshots: []InventorySnapshot{
			snapshotFixture("B1", "AMPHI-A"),
			snapshotFixture("B2", "LAB-204"),
		},
	}

	cache, err := NewInventoryCache(context.Background(), source, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	clock.Advance(ttl + time.Minute)

	got, err := cache.GetInventory(context.Background())
	if err != nil {
		t.Fatalf("expected refresh after expiration to succeed, got error: %v", err)
	}

	if source.Calls() != 2 {
		t.Fatalf("expected warmup + refresh calls, got %d", source.Calls())
	}

	if len(got.Rooms) != 1 || got.Rooms[0].Code != "LAB-204" {
		t.Fatalf("expected refreshed room LAB-204, got %+v", got.Rooms)
	}

	meta := cache.Metadata()
	if !meta.LastRefresh.Equal(clock.Now()) {
		t.Fatalf("expected lastRefresh %s, got %s", clock.Now(), meta.LastRefresh)
	}
	if !meta.ExpiresAt.Equal(clock.Now().Add(ttl)) {
		t.Fatalf("expected expiresAt %s, got %s", clock.Now().Add(ttl), meta.ExpiresAt)
	}
}

func TestInventoryCache_ExposesDegradedStateForHealth(t *testing.T) {
	now := time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC)
	ttl := time.Hour
	clock := &fakeClock{now: now}
	source := &fakeInventorySource{
		snapshots: []InventorySnapshot{
			snapshotFixture("B1", "AMPHI-A"),
			snapshotFixture("B2", "LAB-204"),
		},
	}

	cache, err := NewInventoryCache(context.Background(), source, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	initialHealth := cache.HealthState()
	if initialHealth.Degraded {
		t.Fatalf("expected healthy cache state after warmup")
	}
	if initialHealth.LastAdminError != nil {
		t.Fatalf("expected no admin error timestamp while healthy")
	}
	if !initialHealth.LastRefresh.Equal(now) {
		t.Fatalf("expected last refresh %s, got %s", now, initialHealth.LastRefresh)
	}

	source.SetError(errors.New("admin directory unavailable"))
	clock.Advance(ttl + time.Minute)

	if _, err := cache.GetInventory(context.Background()); err != nil {
		t.Fatalf("expected stale fallback without error, got %v", err)
	}

	degradedHealth := cache.HealthState()
	if !degradedHealth.Degraded {
		t.Fatalf("expected degraded cache state when stale fallback is used")
	}
	if degradedHealth.LastAdminError == nil {
		t.Fatalf("expected admin error timestamp in degraded state")
	}
	if !degradedHealth.LastRefresh.Equal(now) {
		t.Fatalf("expected last refresh to stay on previous successful refresh, got %s", degradedHealth.LastRefresh)
	}

	source.SetError(nil)
	clock.Advance(time.Minute)

	refreshed, err := cache.GetInventory(context.Background())
	if err != nil {
		t.Fatalf("expected refresh recovery to succeed, got %v", err)
	}
	if len(refreshed.Rooms) != 1 || refreshed.Rooms[0].Code != "LAB-204" {
		t.Fatalf("expected refreshed inventory payload after recovery, got %+v", refreshed.Rooms)
	}

	recoveredHealth := cache.HealthState()
	if recoveredHealth.Degraded {
		t.Fatalf("expected degraded state to be cleared after successful refresh")
	}
	if recoveredHealth.LastAdminError != nil {
		t.Fatalf("expected admin error timestamp to be cleared after recovery")
	}
	if !recoveredHealth.LastRefresh.Equal(clock.Now()) {
		t.Fatalf("expected last refresh %s, got %s", clock.Now(), recoveredHealth.LastRefresh)
	}
}

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

func (c *fakeClock) Advance(delta time.Duration) {
	c.now = c.now.Add(delta)
}

type fakeInventorySource struct {
	mu        sync.Mutex
	snapshots []InventorySnapshot
	err       error
	calls     int
}

func (s *fakeInventorySource) LoadInventory(context.Context) (InventorySnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++
	if s.err != nil {
		return InventorySnapshot{}, s.err
	}

	if len(s.snapshots) == 0 {
		return InventorySnapshot{}, nil
	}

	snapshot := s.snapshots[0]
	if len(s.snapshots) > 1 {
		s.snapshots = s.snapshots[1:]
	}

	return snapshot, nil
}

func (s *fakeInventorySource) Calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *fakeInventorySource) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}

func snapshotFixture(buildingID string, roomCode string) InventorySnapshot {
	return InventorySnapshot{
		Buildings: []Building{
			{
				ID:      buildingID,
				Name:    "Building " + buildingID,
				Address: "1 Campus Street",
				Floors:  []int{0, 1, 2},
			},
		},
		Rooms: []Room{
			{
				Code:     roomCode,
				Name:     "Room " + roomCode,
				Building: buildingID,
				Floor:    1,
				Capacity: 30,
				Type:     "lab",
				Status:   "available",
			},
		},
	}
}
