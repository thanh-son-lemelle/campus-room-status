package health

import (
	"context"
	"testing"
	"time"

	"campus-room-status/internal/domain"
)

func TestService_GetHealth_HealthyComplete(t *testing.T) {
	now := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	calendarRefresh := now.Add(-30 * time.Second)
	adminRefresh := now.Add(-time.Minute)

	svc := NewService(
		fakeInventoryHealthReader{
			state: domain.InventoryCacheHealthState{
				Degraded:    false,
				LastRefresh: adminRefresh,
			},
		},
		fakeCalendarHealthReader{
			state: domain.RoomEventsCacheHealthState{
				Degraded:                false,
				LastSuccessfulRefreshAt: &calendarRefresh,
			},
		},
		sequenceClock{times: []time.Time{now, now}},
		"dev",
	)

	health, err := svc.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if health.Status != "ok" {
		t.Fatalf("expected status ok, got %q", health.Status)
	}
	if !health.GoogleAdminAPIConnected {
		t.Fatalf("expected google_admin_api_connected=true")
	}
	if !health.GoogleCalendarAPIConnected {
		t.Fatalf("expected google_calendar_api_connected=true")
	}
	if health.LastSync == nil || !health.LastSync.Equal(calendarRefresh) {
		t.Fatalf("expected last_sync %v, got %v", calendarRefresh, health.LastSync)
	}
}

func TestService_GetHealth_AdminDownButStaleAvailable(t *testing.T) {
	now := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	adminRefresh := now.Add(-time.Hour)
	calendarRefresh := now.Add(-time.Minute)
	adminError := now.Add(-2 * time.Minute)

	svc := NewService(
		fakeInventoryHealthReader{
			state: domain.InventoryCacheHealthState{
				Degraded:       true,
				LastRefresh:    adminRefresh,
				LastAdminError: &adminError,
			},
		},
		fakeCalendarHealthReader{
			state: domain.RoomEventsCacheHealthState{
				Degraded:                false,
				LastSuccessfulRefreshAt: &calendarRefresh,
			},
		},
		sequenceClock{times: []time.Time{now, now}},
		"dev",
	)

	health, err := svc.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if health.Status != "degraded" {
		t.Fatalf("expected degraded status, got %q", health.Status)
	}
	if health.GoogleAdminAPIConnected {
		t.Fatalf("expected google_admin_api_connected=false when admin is degraded")
	}
	if !health.GoogleCalendarAPIConnected {
		t.Fatalf("expected google_calendar_api_connected=true")
	}
}

func TestService_GetHealth_CalendarDownButStaleAvailable(t *testing.T) {
	now := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	adminRefresh := now.Add(-time.Minute)
	calendarRefresh := now.Add(-time.Hour)
	calendarError := now.Add(-2 * time.Minute)

	svc := NewService(
		fakeInventoryHealthReader{
			state: domain.InventoryCacheHealthState{
				Degraded:    false,
				LastRefresh: adminRefresh,
			},
		},
		fakeCalendarHealthReader{
			state: domain.RoomEventsCacheHealthState{
				Degraded:                true,
				LastCalendarErrorAt:     &calendarError,
				LastSuccessfulRefreshAt: &calendarRefresh,
			},
		},
		sequenceClock{times: []time.Time{now, now}},
		"dev",
	)

	health, err := svc.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if health.Status != "degraded" {
		t.Fatalf("expected degraded status, got %q", health.Status)
	}
	if !health.GoogleAdminAPIConnected {
		t.Fatalf("expected google_admin_api_connected=true")
	}
	if health.GoogleCalendarAPIConnected {
		t.Fatalf("expected google_calendar_api_connected=false when calendar is degraded")
	}
}

func TestService_GetHealth_LastSyncIsMostRecentSuccessfulRefresh(t *testing.T) {
	now := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	adminRefresh := now.Add(-45 * time.Second)
	calendarRefresh := now.Add(-2 * time.Minute)

	svc := NewService(
		fakeInventoryHealthReader{
			state: domain.InventoryCacheHealthState{
				Degraded:    false,
				LastRefresh: adminRefresh,
			},
		},
		fakeCalendarHealthReader{
			state: domain.RoomEventsCacheHealthState{
				Degraded:                false,
				LastSuccessfulRefreshAt: &calendarRefresh,
			},
		},
		sequenceClock{times: []time.Time{now, now}},
		"dev",
	)

	health, err := svc.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if health.LastSync == nil || !health.LastSync.Equal(adminRefresh) {
		t.Fatalf("expected last_sync %v, got %v", adminRefresh, health.LastSync)
	}
}

func TestService_GetHealth_ResponseTimeIsSet(t *testing.T) {
	start := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	end := start.Add(12 * time.Millisecond)

	svc := NewService(
		fakeInventoryHealthReader{
			state: domain.InventoryCacheHealthState{
				Degraded:    false,
				LastRefresh: start,
			},
		},
		fakeCalendarHealthReader{
			state: domain.RoomEventsCacheHealthState{
				Degraded: false,
			},
		},
		sequenceClock{times: []time.Time{start, end}},
		"dev",
	)

	health, err := svc.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if health.ResponseTimeMS <= 0 {
		t.Fatalf("expected response_time_ms > 0, got %d", health.ResponseTimeMS)
	}
}

type fakeInventoryHealthReader struct {
	state domain.InventoryCacheHealthState
}

func (f fakeInventoryHealthReader) HealthState() domain.InventoryCacheHealthState {
	return f.state
}

type fakeCalendarHealthReader struct {
	state domain.RoomEventsCacheHealthState
}

func (f fakeCalendarHealthReader) HealthState() domain.RoomEventsCacheHealthState {
	return f.state
}

type sequenceClock struct {
	times []time.Time
	index int
}

func (c sequenceClock) Now() time.Time {
	if len(c.times) == 0 {
		return time.Now().UTC()
	}

	if c.index >= len(c.times) {
		return c.times[len(c.times)-1]
	}

	value := c.times[c.index]
	c.index++
	return value
}
