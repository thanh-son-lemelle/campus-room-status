package rooms

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"campus-room-status/internal/domain"
	mockdata "campus-room-status/internal/mockData"
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

	detail, _, err := svc.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if detail.Code != "AMPHI-A" {
		t.Fatalf("expected room code AMPHI-A, got %q", detail.Code)
	}
}

func TestService_GetRoomDetail_ReturnsRoomNotFoundWhenUnknownCode(t *testing.T) {
	svc := newTestRoomService(t, false)

	_, _, err := svc.GetRoomDetail(context.Background(), "UNKNOWN")
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
			Rooms: domainRoomsFromMock([]mockdata.Room{mockdata.RoomAmphiA()}),
		},
	}
	eventsReader := mapRoomEventsReader{
		eventsByRoom: map[string][]domain.Event{
			"AMPHI-A": domainEventsFromMock(mockdata.DetailOrderedEvents(now)),
		},
	}
	interpreter := domain.NewStatusInterpreter(clock, nil)
	svc := NewService(inventory, eventsReader, interpreter, clock)

	detail, scheduleToday, err := svc.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(scheduleToday) != 3 {
		t.Fatalf("expected 3 events in schedule_today, got %d", len(scheduleToday))
	}
	if scheduleToday[0].Title != "First" || scheduleToday[1].Title != "Second" || scheduleToday[2].Title != "Third" {
		t.Fatalf("expected ordered schedule First->Second->Third, got %+v", scheduleToday)
	}
	_ = detail
}

func TestService_GetRoomDetail_StatusIsCoherentWithTodayEvents(t *testing.T) {
	svc := newTestRoomService(t, false)

	detail, _, err := svc.GetRoomDetail(context.Background(), "AMPHI-A")
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
	detailWithout, _, err := withoutMaintenance.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detailWithout.Status == domain.RoomStatusMaintenance {
		t.Fatalf("did not expect maintenance without reliable source")
	}

	withMaintenance := newTestRoomService(t, true)
	detailWith, _, err := withMaintenance.GetRoomDetail(context.Background(), "AMPHI-A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detailWith.Status != domain.RoomStatusMaintenance {
		t.Fatalf("expected maintenance with reliable source, got %q", detailWith.Status)
	}
}

func TestService_GetRoomSchedule_ReturnsEventsSortedByStartDate(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	clock := roomServiceTestClock{now: now}

	inventory := fakeInventoryReader{
		snapshot: domain.InventorySnapshot{
			Rooms: domainRoomsFromMock([]mockdata.Room{mockdata.RoomAmphiA()}),
		},
	}

	eventsReader := mapRoomEventsReader{
		eventsByRoom: map[string][]domain.Event{
			"AMPHI-A": domainEventsFromMock(mockdata.ScheduleOrderedEvents(now)),
		},
	}

	interpreter := domain.NewStatusInterpreter(clock, nil)
	svc := NewService(inventory, eventsReader, interpreter, clock)

	events, err := svc.GetRoomSchedule(context.Background(), "AMPHI-A", now, now.Add(8*time.Hour))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 schedule events, got %d", len(events))
	}
	if events[0].Title != "First" || events[1].Title != "Second" || events[2].Title != "Third" {
		t.Fatalf("expected sorted schedule First->Second->Third, got %+v", events)
	}
}

func TestService_GetRoomSchedule_ReturnsRoomNotFoundForUnknownCode(t *testing.T) {
	svc := newTestRoomService(t, false)

	_, err := svc.GetRoomSchedule(
		context.Background(),
		"UNKNOWN-ROOM",
		time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected room not found error")
	}

	if _, ok := err.(*domain.RoomNotFoundError); !ok {
		t.Fatalf("expected RoomNotFoundError, got %T", err)
	}
}

func TestService_ListRooms_UsesResourceEmailForCalendarLookupWhenAvailable(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	clock := roomServiceTestClock{now: now}

	inventory := fakeInventoryReader{
		snapshot: domain.InventorySnapshot{
			Rooms: domainRoomsFromMock([]mockdata.Room{mockdata.RoomWithResourceEmail()}),
		},
	}

	eventsReader := &capturingRoomEventsReader{}
	interpreter := domain.NewStatusInterpreter(clock, nil)
	svc := NewService(inventory, eventsReader, interpreter, clock)

	_, err := svc.ListRooms(context.Background(), domain.RoomFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(eventsReader.requestedKeys) != 1 {
		t.Fatalf("expected one events lookup, got %d", len(eventsReader.requestedKeys))
	}
	if eventsReader.requestedKeys[0] != "amphi-a@example.org" {
		t.Fatalf("expected events lookup by resourceEmail, got %q", eventsReader.requestedKeys[0])
	}
}

func TestService_ListRooms_ReturnsInvalidParameterWhenStatusFilterIsInvalid(t *testing.T) {
	svc := NewService(
		panicInventoryReader{},
		panicEventsReader{},
		nil,
		nil,
	)

	_, err := svc.ListRooms(context.Background(), domain.RoomFilters{
		Status: strPtr("unavailble"),
	})
	if err == nil {
		t.Fatalf("expected invalid status error")
	}

	invalidParamErr, ok := err.(*domain.InvalidParameterError)
	if !ok {
		t.Fatalf("expected InvalidParameterError, got %T", err)
	}
	if invalidParamErr.Parameter != "status" {
		t.Fatalf("expected invalid parameter 'status', got %q", invalidParamErr.Parameter)
	}
}

func TestService_ListRooms_PrefiltersBeforeCalendarFetch(t *testing.T) {
	now := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	clock := roomServiceTestClock{now: now}

	inventory := fakeInventoryReader{
		snapshot: domain.InventorySnapshot{
			Rooms: domainRoomsFromMock(mockdata.RoomsForPrefilter()),
		},
	}
	eventsReader := &capturingRoomEventsReader{}
	svc := NewService(inventory, eventsReader, nil, clock)

	_, err := svc.ListRooms(context.Background(), domain.RoomFilters{
		Building: strPtr("B1"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(eventsReader.requestedKeys) != 2 {
		t.Fatalf("expected 2 calendar fetches after prefilter, got %d", len(eventsReader.requestedKeys))
	}
	if !slices.Contains(eventsReader.requestedKeys, "r1@example.org") {
		t.Fatalf("expected fetch for r1@example.org, got %v", eventsReader.requestedKeys)
	}
	if !slices.Contains(eventsReader.requestedKeys, "r2@example.org") {
		t.Fatalf("expected fetch for r2@example.org, got %v", eventsReader.requestedKeys)
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

type capturingRoomEventsReader struct {
	mu            sync.Mutex
	requestedKeys []string
}

func (r *capturingRoomEventsReader) Get(_ context.Context, key domain.RoomEventsKey) ([]domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requestedKeys = append(r.requestedKeys, key.RoomEmail)
	return nil, nil
}

type panicInventoryReader struct{}

func (panicInventoryReader) GetInventory(context.Context) (domain.InventorySnapshot, error) {
	panic("inventory should not be called")
}

type panicEventsReader struct{}

func (panicEventsReader) Get(context.Context, domain.RoomEventsKey) ([]domain.Event, error) {
	panic("events reader should not be called")
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
			Rooms: domainRoomsFromMock(mockdata.RoomsForRoomService()),
		},
	}

	eventsReader := mapRoomEventsReader{
		eventsByRoom: domainEventsMapFromMock(mockdata.RoomServiceEventsByRoom(now)),
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

func domainRoomsFromMock(rooms []mockdata.Room) []domain.Room {
	out := make([]domain.Room, len(rooms))
	for i := range rooms {
		out[i] = domain.Room{
			Code:          rooms[i].Code,
			ResourceEmail: rooms[i].ResourceEmail,
			Name:          rooms[i].Name,
			Building:      rooms[i].Building,
			Floor:         rooms[i].Floor,
			Capacity:      rooms[i].Capacity,
			Type:          rooms[i].Type,
			Status:        rooms[i].Status,
		}
	}
	return out
}

func domainEventsFromMock(events []mockdata.Event) []domain.Event {
	out := make([]domain.Event, len(events))
	for i := range events {
		out[i] = domain.Event{
			Title:     events[i].Title,
			Start:     events[i].Start,
			End:       events[i].End,
			Organizer: events[i].Organizer,
		}
	}
	return out
}

func domainEventsMapFromMock(eventsByRoom map[string][]mockdata.Event) map[string][]domain.Event {
	out := make(map[string][]domain.Event, len(eventsByRoom))
	for roomKey, events := range eventsByRoom {
		out[roomKey] = domainEventsFromMock(events)
	}
	return out
}
