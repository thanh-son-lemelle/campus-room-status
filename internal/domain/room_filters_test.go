package domain

import (
	"errors"
	"testing"

	mockdata "campus-room-status/internal/mockData"
)

func TestFilterAndSortRooms_FilterByBuilding(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		Building: stringRef("B1"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 rooms for building B1, got %d", len(result))
	}
	for _, room := range result {
		if room.Building != "B1" {
			t.Fatalf("expected room building B1, got %q", room.Building)
		}
	}
}

func TestFilterAndSortRooms_FilterByType(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		Type: stringRef("lab"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 lab rooms, got %d", len(result))
	}
	for _, room := range result {
		if room.Type != "lab" {
			t.Fatalf("expected room type lab, got %q", room.Type)
		}
	}
}

func TestFilterAndSortRooms_FilterByStatus(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		Status: stringRef("available"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 available rooms, got %d", len(result))
	}
	for _, room := range result {
		if room.Status != "available" {
			t.Fatalf("expected status available, got %q", room.Status)
		}
	}
}

func TestFilterAndSortRooms_FilterByCapacityMin(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		CapacityMin: intRef(100),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 rooms with capacity >= 100, got %d", len(result))
	}
	for _, room := range result {
		if room.Capacity < 100 {
			t.Fatalf("expected capacity >= 100, got %d", room.Capacity)
		}
	}
}

func TestFilterAndSortRooms_FilterByCapacityMax(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		CapacityMax: intRef(50),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 room with capacity <= 50, got %d", len(result))
	}
	if result[0].Code != "LAB-204" {
		t.Fatalf("expected LAB-204, got %q", result[0].Code)
	}
}

func TestFilterAndSortRooms_SortByName(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		Sort:  stringRef("name"),
		Order: stringRef("asc"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"Alpha Lab", "Amphitheater A", "Computer Lab 204"}
	assertRoomNamesOrder(t, result, expected)
}

func TestFilterAndSortRooms_SortByCapacity(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		Sort:  stringRef("capacity"),
		Order: stringRef("desc"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []int{220, 180, 30}
	assertRoomCapacitiesOrder(t, result, expected)
}

func TestFilterAndSortRooms_SortByStatus(t *testing.T) {
	rooms := testRoomsForFilterSort()

	result, err := FilterAndSortRooms(rooms, RoomFilters{
		Sort:  stringRef("status"),
		Order: stringRef("asc"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"available", "available", "occupied"}
	assertRoomStatusesOrder(t, result, expected)
}

func TestFilterAndSortRooms_OrderAscAndDesc(t *testing.T) {
	rooms := testRoomsForFilterSort()

	asc, err := FilterAndSortRooms(rooms, RoomFilters{
		Sort:  stringRef("capacity"),
		Order: stringRef("asc"),
	})
	if err != nil {
		t.Fatalf("expected no error for asc order, got %v", err)
	}

	desc, err := FilterAndSortRooms(rooms, RoomFilters{
		Sort:  stringRef("capacity"),
		Order: stringRef("desc"),
	})
	if err != nil {
		t.Fatalf("expected no error for desc order, got %v", err)
	}

	assertRoomCapacitiesOrder(t, asc, []int{30, 180, 220})
	assertRoomCapacitiesOrder(t, desc, []int{220, 180, 30})
}

func TestFilterAndSortRooms_ReturnsErrorWhenSortIsInvalid(t *testing.T) {
	_, err := FilterAndSortRooms(testRoomsForFilterSort(), RoomFilters{
		Sort: stringRef("invalid_sort"),
	})
	if err == nil {
		t.Fatalf("expected error for invalid sort")
	}

	var invalidParamErr *InvalidParameterError
	if !errors.As(err, &invalidParamErr) {
		t.Fatalf("expected InvalidParameterError, got %T", err)
	}
	if invalidParamErr.Parameter != "sort" {
		t.Fatalf("expected invalid parameter 'sort', got %q", invalidParamErr.Parameter)
	}
}

func TestFilterAndSortRooms_ReturnsErrorWhenCapacityMinGreaterThanCapacityMax(t *testing.T) {
	_, err := FilterAndSortRooms(testRoomsForFilterSort(), RoomFilters{
		CapacityMin: intRef(200),
		CapacityMax: intRef(100),
	})
	if err == nil {
		t.Fatalf("expected error when capacity_min > capacity_max")
	}

	var invalidParamErr *InvalidParameterError
	if !errors.As(err, &invalidParamErr) {
		t.Fatalf("expected InvalidParameterError, got %T", err)
	}
	if invalidParamErr.Parameter != "capacity_min" {
		t.Fatalf("expected invalid parameter 'capacity_min', got %q", invalidParamErr.Parameter)
	}
}

func TestFilterAndSortRooms_ReturnsErrorWhenStatusIsInvalid(t *testing.T) {
	_, err := FilterAndSortRooms(testRoomsForFilterSort(), RoomFilters{
		Status: stringRef("unavailble"),
	})
	if err == nil {
		t.Fatalf("expected error when status is invalid")
	}

	var invalidParamErr *InvalidParameterError
	if !errors.As(err, &invalidParamErr) {
		t.Fatalf("expected InvalidParameterError, got %T", err)
	}
	if invalidParamErr.Parameter != "status" {
		t.Fatalf("expected invalid parameter 'status', got %q", invalidParamErr.Parameter)
	}
}

func testRoomsForFilterSort() []Room {
	return domainRoomsFromMock(mockdata.RoomsForFilterAndSort())
}

func assertRoomNamesOrder(t *testing.T, rooms []Room, expected []string) {
	t.Helper()
	if len(rooms) != len(expected) {
		t.Fatalf("expected %d rooms, got %d", len(expected), len(rooms))
	}
	for i := range expected {
		if rooms[i].Name != expected[i] {
			t.Fatalf("expected rooms[%d].name %q, got %q", i, expected[i], rooms[i].Name)
		}
	}
}

func assertRoomCapacitiesOrder(t *testing.T, rooms []Room, expected []int) {
	t.Helper()
	if len(rooms) != len(expected) {
		t.Fatalf("expected %d rooms, got %d", len(expected), len(rooms))
	}
	for i := range expected {
		if rooms[i].Capacity != expected[i] {
			t.Fatalf("expected rooms[%d].capacity %d, got %d", i, expected[i], rooms[i].Capacity)
		}
	}
}

func assertRoomStatusesOrder(t *testing.T, rooms []Room, expected []string) {
	t.Helper()
	if len(rooms) != len(expected) {
		t.Fatalf("expected %d rooms, got %d", len(expected), len(rooms))
	}
	for i := range expected {
		if rooms[i].Status != expected[i] {
			t.Fatalf("expected rooms[%d].status %q, got %q", i, expected[i], rooms[i].Status)
		}
	}
}

func stringRef(v string) *string {
	return &v
}

func intRef(v int) *int {
	return &v
}
