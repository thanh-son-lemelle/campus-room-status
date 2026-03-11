package domain

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRoomEventsCache_CacheHitOnAvailability(t *testing.T) {
	now := time.Date(2026, time.March, 10, 14, 0, 0, 0, time.UTC)
	ttl := 2 * time.Minute
	clock := &eventsCacheFakeClock{now: now}
	calendar := &eventsFakeCalendarClient{
		responses: [][]Event{
			eventsFixture("Algorithms"),
		},
	}

	cache, err := NewRoomEventsCache(calendar, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	key := RoomEventsKey{
		RoomEmail: "amphi-a@example.org",
		Start:     now,
		End:       now.Add(time.Hour),
	}

	first, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("expected first load to succeed, got error: %v", err)
	}

	clock.Advance(time.Minute)
	second, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("expected cache hit to succeed, got error: %v", err)
	}

	if calendar.Calls() != 1 {
		t.Fatalf("expected only one calendar call before expiration, got %d", calendar.Calls())
	}

	if len(first) != 1 || first[0].Title != "Algorithms" {
		t.Fatalf("expected first events payload to be Algorithms, got %+v", first)
	}
	if len(second) != 1 || second[0].Title != "Algorithms" {
		t.Fatalf("expected cache hit payload to be Algorithms, got %+v", second)
	}
}

func TestRoomEventsCache_RefreshesAfterExpiration(t *testing.T) {
	now := time.Date(2026, time.March, 10, 15, 0, 0, 0, time.UTC)
	ttl := 90 * time.Second
	clock := &eventsCacheFakeClock{now: now}
	calendar := &eventsFakeCalendarClient{
		responses: [][]Event{
			eventsFixture("Algorithms"),
			eventsFixture("Distributed Systems"),
		},
	}

	cache, err := NewRoomEventsCache(calendar, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	key := RoomEventsKey{
		RoomEmail: "amphi-a@example.org",
		Start:     now,
		End:       now.Add(time.Hour),
	}

	if _, err := cache.Get(context.Background(), key); err != nil {
		t.Fatalf("expected first load to succeed, got error: %v", err)
	}

	clock.Advance(ttl + time.Second)
	refreshed, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("expected refresh to succeed, got error: %v", err)
	}

	if calendar.Calls() != 2 {
		t.Fatalf("expected two calendar calls (initial + refresh), got %d", calendar.Calls())
	}

	if len(refreshed) != 1 || refreshed[0].Title != "Distributed Systems" {
		t.Fatalf("expected refreshed events payload to be Distributed Systems, got %+v", refreshed)
	}
}

func TestRoomEventsCache_FallsBackToStaleWhenCalendarUnavailable(t *testing.T) {
	now := time.Date(2026, time.March, 10, 16, 0, 0, 0, time.UTC)
	ttl := time.Minute
	clock := &eventsCacheFakeClock{now: now}
	calendar := &eventsFakeCalendarClient{
		responses: [][]Event{
			eventsFixture("Algorithms"),
		},
	}

	cache, err := NewRoomEventsCache(calendar, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	key := RoomEventsKey{
		RoomEmail: "amphi-a@example.org",
		Start:     now,
		End:       now.Add(time.Hour),
	}

	initial, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("expected first load to succeed, got error: %v", err)
	}
	if len(initial) != 1 || initial[0].Title != "Algorithms" {
		t.Fatalf("expected initial events payload to be Algorithms, got %+v", initial)
	}

	calendar.SetError(errors.New("calendar unavailable"))
	clock.Advance(ttl + time.Second)

	stale, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("expected stale fallback without error, got %v", err)
	}

	if calendar.Calls() != 2 {
		t.Fatalf("expected refresh attempt after expiration, got %d calls", calendar.Calls())
	}
	if len(stale) != 1 || stale[0].Title != "Algorithms" {
		t.Fatalf("expected stale events payload to stay Algorithms, got %+v", stale)
	}

	meta := cache.Metadata(key)
	if !meta.Stale {
		t.Fatalf("expected cache entry to be marked stale after failed refresh")
	}
}

func TestRoomEventsCache_ExposesDegradedStateForHealth(t *testing.T) {
	now := time.Date(2026, time.March, 10, 17, 0, 0, 0, time.UTC)
	ttl := time.Minute
	clock := &eventsCacheFakeClock{now: now}
	calendar := &eventsFakeCalendarClient{
		responses: [][]Event{
			eventsFixture("Algorithms"),
			eventsFixture("Operating Systems"),
		},
	}

	cache, err := NewRoomEventsCache(calendar, ttl, clock)
	if err != nil {
		t.Fatalf("expected cache creation to succeed, got error: %v", err)
	}

	key := RoomEventsKey{
		RoomEmail: "amphi-a@example.org",
		Start:     now,
		End:       now.Add(time.Hour),
	}

	if _, err := cache.Get(context.Background(), key); err != nil {
		t.Fatalf("expected first load to succeed, got error: %v", err)
	}

	initialHealth := cache.HealthState()
	if initialHealth.Degraded {
		t.Fatalf("expected healthy cache state after successful load")
	}
	if initialHealth.LastCalendarErrorAt != nil {
		t.Fatalf("expected no calendar error timestamp while healthy")
	}
	if initialHealth.LastSuccessfulRefreshAt == nil {
		t.Fatalf("expected successful refresh timestamp while healthy")
	}
	if !initialHealth.LastSuccessfulRefreshAt.Equal(now) {
		t.Fatalf("expected last successful refresh %s, got %s", now, initialHealth.LastSuccessfulRefreshAt)
	}

	calendar.SetError(errors.New("calendar unavailable"))
	clock.Advance(ttl + time.Second)

	if _, err := cache.Get(context.Background(), key); err != nil {
		t.Fatalf("expected stale fallback without error, got %v", err)
	}

	degradedHealth := cache.HealthState()
	if !degradedHealth.Degraded {
		t.Fatalf("expected degraded cache state when stale fallback is used")
	}
	if degradedHealth.LastCalendarErrorAt == nil {
		t.Fatalf("expected calendar error timestamp in degraded state")
	}
	if degradedHealth.LastSuccessfulRefreshAt == nil {
		t.Fatalf("expected successful refresh timestamp to stay available in degraded state")
	}
	if !degradedHealth.LastSuccessfulRefreshAt.Equal(now) {
		t.Fatalf("expected last successful refresh to stay %s, got %s", now, degradedHealth.LastSuccessfulRefreshAt)
	}

	calendar.SetError(nil)
	clock.Advance(time.Second)

	refreshed, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("expected refresh recovery to succeed, got %v", err)
	}
	if len(refreshed) != 1 || refreshed[0].Title != "Operating Systems" {
		t.Fatalf("expected refreshed payload after recovery, got %+v", refreshed)
	}

	recoveredHealth := cache.HealthState()
	if recoveredHealth.Degraded {
		t.Fatalf("expected degraded state to be cleared after successful refresh")
	}
	if recoveredHealth.LastCalendarErrorAt != nil {
		t.Fatalf("expected calendar error timestamp to be cleared after recovery")
	}
	if recoveredHealth.LastSuccessfulRefreshAt == nil {
		t.Fatalf("expected successful refresh timestamp after recovery")
	}
	if !recoveredHealth.LastSuccessfulRefreshAt.Equal(clock.Now()) {
		t.Fatalf("expected last successful refresh %s, got %s", clock.Now(), recoveredHealth.LastSuccessfulRefreshAt)
	}
}

type eventsCacheFakeClock struct {
	now time.Time
}

func (c *eventsCacheFakeClock) Now() time.Time {
	return c.now
}

func (c *eventsCacheFakeClock) Advance(delta time.Duration) {
	c.now = c.now.Add(delta)
}

type eventsFakeCalendarClient struct {
	mu        sync.Mutex
	responses [][]Event
	err       error
	calls     int
}

func (c *eventsFakeCalendarClient) ListRoomEvents(context.Context, string, time.Time, time.Time) ([]Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.calls++
	if c.err != nil {
		return nil, c.err
	}

	if len(c.responses) == 0 {
		return nil, nil
	}

	response := cloneEvents(c.responses[0])
	if len(c.responses) > 1 {
		c.responses = c.responses[1:]
	}

	return response, nil
}

func (c *eventsFakeCalendarClient) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func (c *eventsFakeCalendarClient) SetError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
}

func eventsFixture(title string) []Event {
	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	return []Event{
		{
			Title:     title,
			Start:     start,
			End:       start.Add(90 * time.Minute),
			Organizer: "Academic Office",
		},
	}
}
