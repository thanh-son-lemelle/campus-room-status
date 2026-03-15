package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type RoomEventsKey struct {
	RoomEmail string
	Start     time.Time
	End       time.Time
}

type RoomEventsCacheMetadata struct {
	HasData     bool
	Stale       bool
	ExpiresAt   time.Time
	LastRefresh time.Time
}

type RoomEventsCacheHealthState struct {
	Degraded                bool
	LastCalendarErrorAt     *time.Time
	LastSuccessfulRefreshAt *time.Time
}

type RoomEventsCache struct {
	calendar CalendarClient
	ttl      time.Duration
	clock    Clock

	mu                      sync.RWMutex
	entries                 map[string]roomEventsCacheEntry
	degraded                bool
	lastCalendarErrorAt     *time.Time
	lastSuccessfulRefreshAt *time.Time
	inflight                singleflight.Group
}

type roomEventsCacheEntry struct {
	events      []Event
	expiresAt   time.Time
	lastRefresh time.Time
	hasData     bool
}

func NewRoomEventsCache(calendar CalendarClient, ttl time.Duration, clock Clock) (*RoomEventsCache, error) {
	if calendar == nil {
		return nil, errors.New("calendar client is required")
	}
	if ttl <= 0 {
		return nil, errors.New("ttl must be greater than zero")
	}
	if clock == nil {
		clock = systemClock{}
	}

	return &RoomEventsCache{
		calendar: calendar,
		ttl:      ttl,
		clock:    clock,
		entries:  make(map[string]roomEventsCacheEntry),
	}, nil
}

func (c *RoomEventsCache) Get(ctx context.Context, key RoomEventsKey) ([]Event, error) {
	mapKey, normalizedKey, err := mapKeyFromRoomEventsKey(key)
	if err != nil {
		return nil, err
	}

	now := c.clock.Now()

	c.mu.RLock()
	entry, ok := c.entries[mapKey]
	if ok && entry.hasData && now.Before(entry.expiresAt) {
		events := cloneEvents(entry.events)
		c.mu.RUnlock()
		return events, nil
	}
	c.mu.RUnlock()

	result, fetchErr, _ := c.inflight.Do(mapKey, func() (any, error) {
		return c.refresh(ctx, mapKey, normalizedKey)
	})
	if fetchErr != nil {
		return nil, fetchErr
	}

	events, ok := result.([]Event)
	if !ok {
		return nil, errors.New("unexpected room events cache result type")
	}

	return cloneEvents(events), nil
}

func (c *RoomEventsCache) Metadata(key RoomEventsKey) RoomEventsCacheMetadata {
	mapKey, _, err := mapKeyFromRoomEventsKey(key)
	if err != nil {
		return RoomEventsCacheMetadata{}
	}

	now := c.clock.Now()

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[mapKey]
	if !ok || !entry.hasData {
		return RoomEventsCacheMetadata{}
	}

	return RoomEventsCacheMetadata{
		HasData:     true,
		Stale:       !now.Before(entry.expiresAt),
		ExpiresAt:   entry.expiresAt,
		LastRefresh: entry.lastRefresh,
	}
}

func (c *RoomEventsCache) HealthState() RoomEventsCacheHealthState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state := RoomEventsCacheHealthState{
		Degraded: c.degraded,
	}
	if c.lastCalendarErrorAt != nil {
		lastError := *c.lastCalendarErrorAt
		state.LastCalendarErrorAt = &lastError
	}
	if c.lastSuccessfulRefreshAt != nil {
		lastSuccess := *c.lastSuccessfulRefreshAt
		state.LastSuccessfulRefreshAt = &lastSuccess
	}

	return state
}

func (c *RoomEventsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]roomEventsCacheEntry)
	c.degraded = false
	c.lastCalendarErrorAt = nil
	c.lastSuccessfulRefreshAt = nil
}

func (c *RoomEventsCache) refresh(ctx context.Context, mapKey string, key RoomEventsKey) ([]Event, error) {
	now := c.clock.Now()

	c.mu.RLock()
	entry, ok := c.entries[mapKey]
	if ok && entry.hasData && now.Before(entry.expiresAt) {
		events := cloneEvents(entry.events)
		c.mu.RUnlock()
		return events, nil
	}
	c.mu.RUnlock()

	events, fetchErr := c.calendar.ListRoomEvents(ctx, key.RoomEmail, key.Start, key.End)

	c.mu.Lock()
	defer c.mu.Unlock()

	now = c.clock.Now()
	entry, ok = c.entries[mapKey]
	if ok && entry.hasData && now.Before(entry.expiresAt) {
		return cloneEvents(entry.events), nil
	}

	if fetchErr != nil {
		failedAt := now
		c.degraded = true
		c.lastCalendarErrorAt = &failedAt

		if ok && entry.hasData {
			return cloneEvents(entry.events), nil
		}
		return nil, fetchErr
	}

	refreshed := roomEventsCacheEntry{
		events:      cloneEvents(events),
		hasData:     true,
		lastRefresh: now,
		expiresAt:   now.Add(c.ttl),
	}
	c.entries[mapKey] = refreshed

	c.degraded = false
	c.lastCalendarErrorAt = nil
	successAt := now
	c.lastSuccessfulRefreshAt = &successAt

	return cloneEvents(refreshed.events), nil
}

func mapKeyFromRoomEventsKey(key RoomEventsKey) (string, RoomEventsKey, error) {
	normalized, err := normalizeRoomEventsKey(key)
	if err != nil {
		return "", RoomEventsKey{}, err
	}

	cacheKey := fmt.Sprintf(
		"%s|%d|%d",
		normalized.RoomEmail,
		normalized.Start.UnixNano(),
		normalized.End.UnixNano(),
	)

	return cacheKey, normalized, nil
}

func normalizeRoomEventsKey(key RoomEventsKey) (RoomEventsKey, error) {
	roomEmail := strings.TrimSpace(key.RoomEmail)
	if roomEmail == "" {
		return RoomEventsKey{}, errors.New("room email is required")
	}

	start := key.Start.UTC()
	end := key.End.UTC()
	if !end.After(start) {
		return RoomEventsKey{}, errors.New("end must be after start")
	}

	return RoomEventsKey{
		RoomEmail: roomEmail,
		Start:     start,
		End:       end,
	}, nil
}

func cloneEvents(events []Event) []Event {
	if events == nil {
		return nil
	}

	cloned := make([]Event, len(events))
	copy(cloned, events)
	return cloned
}
