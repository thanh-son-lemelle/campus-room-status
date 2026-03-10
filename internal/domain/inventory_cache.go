package domain

import (
	"context"
	"errors"
	"sync"
	"time"
)

type InventorySnapshot struct {
	Buildings []Building
	Rooms     []Room
}

type InventorySource interface {
	LoadInventory(ctx context.Context) (InventorySnapshot, error)
}

type CacheClock interface {
	Now() time.Time
}

type InventoryCacheMetadata struct {
	ExpiresAt   time.Time
	LastRefresh time.Time
}

type InventoryCache struct {
	source InventorySource
	ttl    time.Duration
	clock  CacheClock

	mu          sync.RWMutex
	snapshot    InventorySnapshot
	expiresAt   time.Time
	lastRefresh time.Time
	hasData     bool
}

func NewInventoryCache(ctx context.Context, source InventorySource, ttl time.Duration, clock CacheClock) (*InventoryCache, error) {
	if source == nil {
		return nil, errors.New("inventory source is required")
	}
	if ttl <= 0 {
		return nil, errors.New("ttl must be greater than zero")
	}
	if clock == nil {
		clock = cacheSystemClock{}
	}

	cache := &InventoryCache{
		source: source,
		ttl:    ttl,
		clock:  clock,
	}

	if err := cache.warmup(ctx); err != nil {
		return nil, err
	}

	return cache, nil
}

func (c *InventoryCache) GetInventory(ctx context.Context) (InventorySnapshot, error) {
	now := c.clock.Now()

	c.mu.RLock()
	if c.hasData && now.Before(c.expiresAt) {
		snapshot := cloneSnapshot(c.snapshot)
		c.mu.RUnlock()
		return snapshot, nil
	}
	c.mu.RUnlock()

	return c.refresh(ctx)
}

func (c *InventoryCache) Metadata() InventoryCacheMetadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return InventoryCacheMetadata{
		ExpiresAt:   c.expiresAt,
		LastRefresh: c.lastRefresh,
	}
}

func (c *InventoryCache) warmup(ctx context.Context) error {
	snapshot, err := c.source.LoadInventory(ctx)
	if err != nil {
		return err
	}

	now := c.clock.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshot = cloneSnapshot(snapshot)
	c.lastRefresh = now
	c.expiresAt = now.Add(c.ttl)
	c.hasData = true

	return nil
}

func (c *InventoryCache) refresh(ctx context.Context) (InventorySnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock.Now()
	if c.hasData && now.Before(c.expiresAt) {
		return cloneSnapshot(c.snapshot), nil
	}

	snapshot, err := c.source.LoadInventory(ctx)
	if err != nil {
		if c.hasData {
			return cloneSnapshot(c.snapshot), nil
		}
		return InventorySnapshot{}, err
	}

	c.snapshot = cloneSnapshot(snapshot)
	c.lastRefresh = now
	c.expiresAt = now.Add(c.ttl)
	c.hasData = true

	return cloneSnapshot(c.snapshot), nil
}

func cloneSnapshot(src InventorySnapshot) InventorySnapshot {
	var out InventorySnapshot

	if src.Buildings != nil {
		out.Buildings = make([]Building, len(src.Buildings))
		for i := range src.Buildings {
			out.Buildings[i] = src.Buildings[i]
			if src.Buildings[i].Floors != nil {
				out.Buildings[i].Floors = append([]int(nil), src.Buildings[i].Floors...)
			}
		}
	}

	if src.Rooms != nil {
		out.Rooms = make([]Room, len(src.Rooms))
		for i := range src.Rooms {
			out.Rooms[i] = src.Rooms[i]
			if src.Rooms[i].CurrentEvent != nil {
				currentEvent := *src.Rooms[i].CurrentEvent
				out.Rooms[i].CurrentEvent = &currentEvent
			}
			if src.Rooms[i].NextEvent != nil {
				nextEvent := *src.Rooms[i].NextEvent
				out.Rooms[i].NextEvent = &nextEvent
			}
		}
	}

	return out
}

type cacheSystemClock struct{}

func (cacheSystemClock) Now() time.Time {
	return time.Now().UTC()
}
