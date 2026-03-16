package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	mockdata "campus-room-status/internal/mockData"
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

	room := domainDirectoryRoomFromMock(mockdata.DirectoryRoomAmphiA())
	events := domainEventsFromMock(mockdata.RoomServiceEventsByRoom(now)["AMPHI-A"][:1])

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusOccupied {
		t.Fatalf("expected status %q, got %q", RoomStatusOccupied, status)
	}
}

func TestStatusInterpreter_ReturnsUpcomingWhenNextEventStartsInLessThanThirtyMinutes(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(testClock{now: now}, nil)

	room := domainDirectoryRoomFromMock(mockdata.DirectoryRoomAmphiA())
	events := domainEventsFromMock(mockdata.UpcomingEvent(now))

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusUpcoming {
		t.Fatalf("expected status %q, got %q", RoomStatusUpcoming, status)
	}
}

func TestStatusInterpreter_ReturnsAvailableWhenNoCurrentOrImminentEvent(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(testClock{now: now}, nil)

	room := domainDirectoryRoomFromMock(mockdata.DirectoryRoomAmphiA())
	events := domainEventsFromMock(mockdata.FutureEvent(now))

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

	room := domainDirectoryRoomFromMock(mockdata.DirectoryRoomAmphiA())
	events := domainEventsFromMock(mockdata.OccupiedEvent(now))

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusMaintenance {
		t.Fatalf("expected status %q, got %q", RoomStatusMaintenance, status)
	}
}

func TestStatusInterpreter_DoesNotInventMaintenanceWhenNoReliableSourceExists(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	interpreter := NewStatusInterpreter(testClock{now: now}, nil)

	room := domainDirectoryRoomFromMock(mockdata.DirectoryRoomAmphiAMaintenance())
	events := domainEventsFromMock(mockdata.NoonSessionEvent(now))

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

	room := domainDirectoryRoomFromMock(mockdata.DirectoryRoomAmphiA())
	events := domainEventsFromMock(mockdata.OccupiedEvent(now))

	status := interpreter.Resolve(context.Background(), room, events)
	if status != RoomStatusOccupied {
		t.Fatalf("expected fallback status %q when maintenance source fails, got %q", RoomStatusOccupied, status)
	}
}
