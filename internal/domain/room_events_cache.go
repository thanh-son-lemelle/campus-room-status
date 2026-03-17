package domain

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const defaultRoomEventsChunkWindow = 30 * 24 * time.Hour

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
	calendar    CalendarClient
	ttl         time.Duration
	clock       Clock
	chunkWindow time.Duration

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

// NewRoomEventsCache creates a new room events cache.
//
// Summary:
// - Creates a new room events cache.
//
// Attributes:
// - calendar (CalendarClient): Input parameter.
// - ttl (time.Duration): Input parameter.
// - clock (Clock): Input parameter.
//
// Returns:
// - value1 (*RoomEventsCache): Returned value.
// - value2 (error): Returned value.
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
		calendar:    calendar,
		ttl:         ttl,
		clock:       clock,
		chunkWindow: defaultRoomEventsChunkWindow,
		entries:     make(map[string]roomEventsCacheEntry),
	}, nil
}

// Get gets data.
//
// Summary:
// - Gets data.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - key (RoomEventsKey): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
// - value2 (error): Returned value.
func (c *RoomEventsCache) Get(ctx context.Context, key RoomEventsKey) ([]Event, error) {
	_, normalizedKey, err := mapKeyFromRoomEventsKey(key)
	if err != nil {
		return nil, err
	}

	chunks := splitRoomEventsKeyIntoChunks(normalizedKey, c.chunkWindow)
	events := make([]Event, 0)
	for _, chunkKey := range chunks {
		chunkEvents, fetchErr := c.getChunk(ctx, chunkKey)
		if fetchErr != nil {
			return nil, fetchErr
		}
		events = append(events, chunkEvents...)
	}

	return normalizeEventsForRange(events, normalizedKey.Start, normalizedKey.End), nil
}

// Metadata metadatas function behavior.
//
// Summary:
// - Metadatas function behavior.
//
// Attributes:
// - key (RoomEventsKey): Input parameter.
//
// Returns:
// - value1 (RoomEventsCacheMetadata): Returned value.
func (c *RoomEventsCache) Metadata(key RoomEventsKey) RoomEventsCacheMetadata {
	_, normalizedKey, err := mapKeyFromRoomEventsKey(key)
	if err != nil {
		return RoomEventsCacheMetadata{}
	}
	chunks := splitRoomEventsKeyIntoChunks(normalizedKey, c.chunkWindow)
	if len(chunks) == 0 {
		return RoomEventsCacheMetadata{}
	}

	now := c.clock.Now()

	c.mu.RLock()
	defer c.mu.RUnlock()

	metadata := RoomEventsCacheMetadata{HasData: true}
	for i, chunk := range chunks {
		mapKey, _, mapErr := mapKeyFromRoomEventsKey(chunk)
		if mapErr != nil {
			return RoomEventsCacheMetadata{}
		}

		entry, ok := c.entries[mapKey]
		if !ok || !entry.hasData {
			return RoomEventsCacheMetadata{}
		}

		if !now.Before(entry.expiresAt) {
			metadata.Stale = true
		}
		if i == 0 || entry.expiresAt.Before(metadata.ExpiresAt) {
			metadata.ExpiresAt = entry.expiresAt
		}
		if entry.lastRefresh.After(metadata.LastRefresh) {
			metadata.LastRefresh = entry.lastRefresh
		}
	}

	return metadata
}

// getChunk gets a chunked key from cache or provider.
func (c *RoomEventsCache) getChunk(ctx context.Context, key RoomEventsKey) ([]Event, error) {
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

// HealthState healths state.
//
// Summary:
// - Healths state.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (RoomEventsCacheHealthState): Returned value.
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

// Clear clears state.
//
// Summary:
// - Clears state.
//
// Attributes:
// - None.
//
// Returns:
// - None.
func (c *RoomEventsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]roomEventsCacheEntry)
	c.degraded = false
	c.lastCalendarErrorAt = nil
	c.lastSuccessfulRefreshAt = nil
}

// refresh refreshes function behavior.
//
// Summary:
// - Refreshes function behavior.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - mapKey (string): Input parameter.
// - key (RoomEventsKey): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
// - value2 (error): Returned value.
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

// mapKeyFromRoomEventsKey maps key from room events key.
//
// Summary:
// - Maps key from room events key.
//
// Attributes:
// - key (RoomEventsKey): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
// - value2 (RoomEventsKey): Returned value.
// - value3 (error): Returned value.
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

// splitRoomEventsKeyIntoChunks splits a key into globally aligned chunks.
func splitRoomEventsKeyIntoChunks(key RoomEventsKey, chunkWindow time.Duration) []RoomEventsKey {
	if !key.End.After(key.Start) {
		return nil
	}
	if chunkWindow <= 0 {
		return []RoomEventsKey{key}
	}

	chunkStart := windowStartForTime(key.Start, chunkWindow)
	chunks := make([]RoomEventsKey, 0, int(key.End.Sub(chunkStart)/chunkWindow)+1)

	for chunkStart.Before(key.End) {
		chunkEnd := chunkStart.Add(chunkWindow)
		if !chunkEnd.After(chunkStart) {
			break
		}

		chunks = append(chunks, RoomEventsKey{
			RoomEmail: key.RoomEmail,
			Start:     chunkStart,
			End:       chunkEnd,
		})

		chunkStart = chunkEnd
	}

	return chunks
}

// windowStartForTime returns the aligned window start for the given timestamp.
func windowStartForTime(value time.Time, window time.Duration) time.Time {
	if window <= 0 {
		return value.UTC()
	}

	utc := value.UTC()
	windowNS := int64(window)
	valueNS := utc.UnixNano()
	remainder := valueNS % windowNS
	if remainder < 0 {
		remainder += windowNS
	}

	return time.Unix(0, valueNS-remainder).UTC()
}

// normalizeRoomEventsKey normalizes room events key.
//
// Summary:
// - Normalizes room events key.
//
// Attributes:
// - key (RoomEventsKey): Input parameter.
//
// Returns:
// - value1 (RoomEventsKey): Returned value.
// - value2 (error): Returned value.
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

// normalizeEventsForRange filters, deduplicates and sorts events for a requested range.
func normalizeEventsForRange(events []Event, start, end time.Time) []Event {
	if len(events) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(events))
	out := make([]Event, 0, len(events))
	for _, event := range events {
		if !event.End.After(start) || !event.Start.Before(end) {
			continue
		}

		fingerprint := eventFingerprint(event)
		if _, exists := seen[fingerprint]; exists {
			continue
		}
		seen[fingerprint] = struct{}{}
		out = append(out, event)
	}

	if len(out) == 0 {
		return nil
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			if out[i].End.Equal(out[j].End) {
				if out[i].Title == out[j].Title {
					return out[i].Organizer < out[j].Organizer
				}
				return out[i].Title < out[j].Title
			}
			return out[i].End.Before(out[j].End)
		}
		return out[i].Start.Before(out[j].Start)
	})

	return out
}

func eventFingerprint(event Event) string {
	return fmt.Sprintf(
		"%s|%d|%d|%s",
		event.Title,
		event.Start.UTC().UnixNano(),
		event.End.UTC().UnixNano(),
		event.Organizer,
	)
}

// cloneEvents clones events.
//
// Summary:
// - Clones events.
//
// Attributes:
// - events ([]Event): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func cloneEvents(events []Event) []Event {
	if events == nil {
		return nil
	}

	cloned := make([]Event, len(events))
	copy(cloned, events)
	return cloned
}
