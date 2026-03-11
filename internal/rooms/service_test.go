package rooms

import (
	"context"
	"testing"
	"time"

	"campus-room-status/internal/domain"
)

func TestService_ListRooms_ReturnsCompleteListWithoutFilters(t *testing.T) {
	svc := newTestRoomService(t, false)

	rooms, err := svc.ListRooms(context.Background(), domain.RoomFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rooms) != 3 {
		t.Fatalf("expected 3 rooms, got %d", len(rooms))
	}
}

func TestService_ListRooms_FiltersByBuilding(t *testing.T) {
	svc := newTestRoomService(t, false)

	rooms, err := svc.ListRooms(context.Background(), domain.RoomFilters{
		Building: strPtr("B1"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms for building B1, got %d", len(rooms))
	}
	for _, room := range rooms {
		if room.Building != "B1" {
			t.Fatalf("expected room building B1, got %q", room.Building)
		}
	}
}

func TestService_ListRooms_FiltersByStatus(t *testing.T) {
	svc := newTestRoomService(t, false)

	rooms, err := svc.ListRooms(context.Background(), domain.RoomFilters{
		Status: strPtr(domain.RoomStatusUpcoming),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rooms) != 1 {
		t.Fatalf("expected 1 upcoming room, got %d", len(rooms))
	}
	if rooms[0].Code != "LAB-204" {
		t.Fatalf("expected LAB-204, got %q", rooms[0].Code)
	}
}

func TestService_ListRooms_SortsAscAndDesc(t *testing.T) {
	svc := newTestRoomService(t, false)

	asc, err := svc.ListRooms(context.Background(), domain.RoomFilters{
		Sort:  strPtr("name"),
		Order: strPtr("asc"),
	})
	if err != nil {
		t.Fatalf("expected no error for asc sort, got %v", err)
	}
	if asc[0].Name != "Alpha Lab" {
		t.Fatalf("expected Alpha Lab first in asc sort, got %q", asc[0].Name)
	}

	desc, err := svc.ListRooms(context.Background(), domain.RoomFilters{
		Sort:  strPtr("name"),
		Order: strPtr("desc"),
	})
	if err != nil {
		t.Fatalf("expected no error for desc sort, got %v", err)
	}
	if desc[0].Name != "Computer Lab 204" {
		t.Fatalf("expected Computer Lab 204 first in desc sort, got %q", desc[0].Name)
	}
}

func TestService_ListRooms_CurrentEventIsNilWhenRoomIsFree(t *testing.T) {
	svc := newTestRoomService(t, false)

	rooms, err := svc.ListRooms(context.Background(), domain.RoomFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	room := findRoomByCodeInSlice(t, rooms, "LAB-101")
	if room.CurrentEvent != nil {
		t.Fatalf("expected LAB-101 current_event to be nil for free room")
	}
}

func TestService_ListRooms_NextEventIsComputedCorrectly(t *testing.T) {
	svc := newTestRoomService(t, false)

	rooms, err := svc.ListRooms(context.Background(), domain.RoomFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	room := findRoomByCodeInSlice(t, rooms, "AMPHI-A")
	if room.NextEvent == nil {
		t.Fatalf("expected AMPHI-A to expose next_event")
	}
	if room.NextEvent.Title != "Security" {
		t.Fatalf("expected AMPHI-A next_event title Security, got %q", room.NextEvent.Title)
	}
}

func TestService_ListRooms_DerivesStatusFromEvents(t *testing.T) {
	svc := newTestRoomService(t, false)

	rooms, err := svc.ListRooms(context.Background(), domain.RoomFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	amphi := findRoomByCodeInSlice(t, rooms, "AMPHI-A")
	if amphi.Status != domain.RoomStatusOccupied {
		t.Fatalf("expected AMPHI-A status occupied, got %q", amphi.Status)
	}

	lab204 := findRoomByCodeInSlice(t, rooms, "LAB-204")
	if lab204.Status != domain.RoomStatusUpcoming {
		t.Fatalf("expected LAB-204 status upcoming, got %q", lab204.Status)
	}

	lab101 := findRoomByCodeInSlice(t, rooms, "LAB-101")
	if lab101.Status != domain.RoomStatusAvailable {
		t.Fatalf("expected LAB-101 status available, got %q", lab101.Status)
	}
}

func TestService_ListRooms_CanReturnMaintenanceWhenReliableUnavailabilityExists(t *testing.T) {
	svc := newTestRoomService(t, true)

	rooms, err := svc.ListRooms(context.Background(), domain.RoomFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, room := range rooms {
		if room.Status != domain.RoomStatusMaintenance {
			t.Fatalf("expected maintenance status when unavailability source is reliable, got %q", room.Status)
		}
	}
}

func TestService_GetRoomDetail_ReturnsKnownRoom(t *testing.T) {
	svc := newTestRoomService(t, false)

	detail, err := svc.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if detail.Code != "AMPHI-A" {
		t.Fatalf("expected room code AMPHI-A, got %q", detail.Code)
	}
}

func TestService_GetRoomDetail_ReturnsRoomNotFoundWhenUnknownCode(t *testing.T) {
	svc := newTestRoomService(t, false)

	_, err := svc.GetRoomDetail(context.Background(), "UNKNOWN")
	if err == nil {
		t.Fatalf("expected room not found error")
	}

	notFoundErr, ok := err.(*domain.RoomNotFoundError)
	if !ok {
		t.Fatalf("expected RoomNotFoundError, got %T", err)
	}
	if notFoundErr.RoomCode != "UNKNOWN" {
		t.Fatalf("expected missing code UNKNOWN, got %q", notFoundErr.RoomCode)
	}
}

func TestService_GetRoomDetail_ScheduleTodayIsOrdered(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	clock := roomServiceTestClock{now: now}

	inventory := fakeInventoryReader{
		snapshot: domain.InventorySnapshot{
			Rooms: []domain.Room{
				{Code: "AMPHI-A", Name: "Amphitheater A", Building: "B1", Capacity: 180, Type: "amphitheater"},
			},
		},
	}
	eventsReader := mapRoomEventsReader{
		eventsByRoom: map[string][]domain.Event{
			"AMPHI-A": {
				{Title: "Third", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
				{Title: "First", Start: now.Add(-30 * time.Minute), End: now.Add(15 * time.Minute)},
				{Title: "Second", Start: now.Add(time.Hour), End: now.Add(90 * time.Minute)},
			},
		},
	}
	interpreter := domain.NewStatusInterpreter(clock, nil)
	svc := NewService(inventory, eventsReader, interpreter, clock)

	detail, err := svc.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(detail.ScheduleToday) != 3 {
		t.Fatalf("expected 3 events in schedule_today, got %d", len(detail.ScheduleToday))
	}
	if detail.ScheduleToday[0].Title != "First" || detail.ScheduleToday[1].Title != "Second" || detail.ScheduleToday[2].Title != "Third" {
		t.Fatalf("expected ordered schedule First->Second->Third, got %+v", detail.ScheduleToday)
	}
}

func TestService_GetRoomDetail_StatusIsCoherentWithTodayEvents(t *testing.T) {
	svc := newTestRoomService(t, false)

	detail, err := svc.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if detail.Status != domain.RoomStatusOccupied {
		t.Fatalf("expected occupied status, got %q", detail.Status)
	}
	if detail.CurrentEvent == nil {
		t.Fatalf("expected current_event for occupied room")
	}
}

func TestService_GetRoomDetail_MaintenanceOnlyWhenReliableDataExists(t *testing.T) {
	withoutMaintenance := newTestRoomService(t, false)
	detailWithout, err := withoutMaintenance.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detailWithout.Status == domain.RoomStatusMaintenance {
		t.Fatalf("did not expect maintenance without reliable source")
	}

	withMaintenance := newTestRoomService(t, true)
	detailWith, err := withMaintenance.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detailWith.Status != domain.RoomStatusMaintenance {
		t.Fatalf("expected maintenance with reliable source, got %q", detailWith.Status)
	}
}

type roomServiceTestClock struct {
	now time.Time
}

func (c roomServiceTestClock) Now() time.Time {
	return c.now
}

type fakeInventoryReader struct {
	snapshot domain.InventorySnapshot
}

func (f fakeInventoryReader) GetInventory(context.Context) (domain.InventorySnapshot, error) {
	return f.snapshot, nil
}

type mapRoomEventsReader struct {
	eventsByRoom map[string][]domain.Event
}

func (m mapRoomEventsReader) Get(_ context.Context, key domain.RoomEventsKey) ([]domain.Event, error) {
	events := m.eventsByRoom[key.RoomEmail]
	out := make([]domain.Event, len(events))
	copy(out, events)
	return out, nil
}

type allUnavailableSource struct {
	unavailable bool
}

func (s allUnavailableSource) IsRoomUnavailable(context.Context, string, time.Time) (bool, error) {
	return s.unavailable, nil
}

func newTestRoomService(t *testing.T, forceMaintenance bool) domain.RoomService {
	t.Helper()

	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	clock := roomServiceTestClock{now: now}

	inventory := fakeInventoryReader{
		snapshot: domain.InventorySnapshot{
			Rooms: []domain.Room{
				{Code: "AMPHI-A", Name: "Amphitheater A", Building: "B1", Capacity: 180, Type: "amphitheater"},
				{Code: "LAB-204", Name: "Computer Lab 204", Building: "B2", Capacity: 30, Type: "lab"},
				{Code: "LAB-101", Name: "Alpha Lab", Building: "B1", Capacity: 120, Type: "lab"},
			},
		},
	}

	eventsReader := mapRoomEventsReader{
		eventsByRoom: map[string][]domain.Event{
			"AMPHI-A": {
				{
					Title: "Algorithms",
					Start: now.Add(-10 * time.Minute),
					End:   now.Add(20 * time.Minute),
				},
				{
					Title: "Security",
					Start: now.Add(40 * time.Minute),
					End:   now.Add(100 * time.Minute),
				},
			},
			"LAB-204": {
				{
					Title: "OS Lab",
					Start: now.Add(20 * time.Minute),
					End:   now.Add(80 * time.Minute),
				},
			},
			"LAB-101": {},
		},
	}

	var unavailability domain.UnavailabilitySource
	if forceMaintenance {
		unavailability = allUnavailableSource{unavailable: true}
	}

	interpreter := domain.NewStatusInterpreter(clock, unavailability)
	return NewService(inventory, eventsReader, interpreter, clock)
}

func findRoomByCodeInSlice(t *testing.T, rooms []domain.Room, code string) domain.Room {
	t.Helper()

	for _, room := range rooms {
		if room.Code == code {
			return room
		}
	}

	t.Fatalf("expected room %s in result set", code)
	return domain.Room{}
}

func strPtr(v string) *string {
	return &v
}
