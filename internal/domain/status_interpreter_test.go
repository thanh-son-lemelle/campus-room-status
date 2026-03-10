package domain

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testClock struct {
	now time.Time
}

func (c testClock) Now() time.Time {
	return c.now
}

type fakeUnavailabilitySource struct {
	unavailable bool
	err         error
}

func (s fakeUnavailabilitySource) IsRoomUnavailable(ctx context.Context, roomCode string, at time.Time) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.unavailable, nil
}

func TestStatusInterpreter_ReturnsOccupiedWhenEventIsInProgress(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(testClock{now: now}, nil)

	room := DirectoryRoom{
		ResourceName:  "AMPHI-A",
		ResourceEmail: "amphi-a@example.org",
	}
	events := []Event{
		{
			Title: "Algorithms",
			Start: now.Add(-10 * time.Minute),
			End:   now.Add(20 * time.Minute),
		},
	}

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusOccupied {
		t.Fatalf("expected status %q, got %q", RoomStatusOccupied, status)
	}
}

func TestStatusInterpreter_ReturnsUpcomingWhenNextEventStartsInLessThanThirtyMinutes(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(testClock{now: now}, nil)

	room := DirectoryRoom{
		ResourceName:  "AMPHI-A",
		ResourceEmail: "amphi-a@example.org",
	}
	events := []Event{
		{
			Title: "Distributed Systems",
			Start: now.Add(29*time.Minute + 30*time.Second),
			End:   now.Add(89 * time.Minute),
		},
	}

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusUpcoming {
		t.Fatalf("expected status %q, got %q", RoomStatusUpcoming, status)
	}
}

func TestStatusInterpreter_ReturnsAvailableWhenNoCurrentOrImminentEvent(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(testClock{now: now}, nil)

	room := DirectoryRoom{
		ResourceName:  "AMPHI-A",
		ResourceEmail: "amphi-a@example.org",
	}
	events := []Event{
		{
			Title: "Future Session",
			Start: now.Add(45 * time.Minute),
			End:   now.Add(2 * time.Hour),
		},
	}

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusAvailable {
		t.Fatalf("expected status %q, got %q", RoomStatusAvailable, status)
	}
}

func TestStatusInterpreter_ReturnsMaintenanceWhenExternalUnavailabilitySourceSaysUnavailable(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(
		testClock{now: now},
		fakeUnavailabilitySource{unavailable: true},
	)

	room := DirectoryRoom{
		ResourceName:  "AMPHI-A",
		ResourceEmail: "amphi-a@example.org",
	}
	events := []Event{
		{
			Title: "Current Event",
			Start: now.Add(-5 * time.Minute),
			End:   now.Add(5 * time.Minute),
		},
	}

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusMaintenance {
		t.Fatalf("expected status %q, got %q", RoomStatusMaintenance, status)
	}
}

func TestStatusInterpreter_DoesNotInventMaintenanceWhenNoReliableSourceExists(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(testClock{now: now}, nil)

	room := DirectoryRoom{
		ResourceName:     "AMPHI-A",
		ResourceEmail:    "amphi-a@example.org",
		ResourceCategory: "maintenance",
	}
	events := []Event{
		{
			Title: "Noon Session",
			Start: now.Add(3 * time.Hour),
			End:   now.Add(4 * time.Hour),
		},
	}

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusAvailable {
		t.Fatalf("expected status %q without reliable maintenance source, got %q", RoomStatusAvailable, status)
	}
}

func TestStatusInterpreter_IgnoresMaintenanceWhenSourceFailsAndFallsBackToCalendarRules(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(
		testClock{now: now},
		fakeUnavailabilitySource{err: errors.New("source temporarily unavailable")},
	)

	room := DirectoryRoom{
		ResourceName:  "AMPHI-A",
		ResourceEmail: "amphi-a@example.org",
	}
	events := []Event{
		{
			Title: "Current Event",
			Start: now.Add(-5 * time.Minute),
			End:   now.Add(5 * time.Minute),
		},
	}

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusOccupied {
		t.Fatalf("expected fallback status %q when maintenance source fails, got %q", RoomStatusOccupied, status)
	}
}
