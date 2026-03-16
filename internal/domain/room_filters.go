package domain

import (
	"sort"
	"strings"
)

// FilterAndSortRooms filters and sort rooms.
//
// Summary:
// - Filters and sort rooms.
//
// Attributes:
// - rooms ([]Room): Input parameter.
// - filters (RoomFilters): Input parameter.
//
// Returns:
// - value1 ([]Room): Returned value.
// - value2 (error): Returned value.
func FilterAndSortRooms(rooms []Room, filters RoomFilters) ([]Room, error) {
	if err := ValidateRoomFilters(filters); err != nil {
		return nil, err
	}

	filtered := filterRooms(rooms, filters, true)

	sortField := normalizedStringPointer(filters.Sort)
	if sortField != "" {
		order := normalizedStringPointer(filters.Order)
		if order == "" {
			order = "asc"
		}

		sort.SliceStable(filtered, func(i, j int) bool {
			switch sortField {
			case "name":
				left := strings.ToLower(filtered[i].Name)
				right := strings.ToLower(filtered[j].Name)
				if left == right {
					return false
				}
				if order == "desc" {
					return left > right
				}
				return left < right
			case "capacity":
				if filtered[i].Capacity == filtered[j].Capacity {
					return false
				}
				if order == "desc" {
					return filtered[i].Capacity > filtered[j].Capacity
				}
				return filtered[i].Capacity < filtered[j].Capacity
			case "status":
				left := strings.ToLower(filtered[i].Status)
				right := strings.ToLower(filtered[j].Status)
				if left == right {
					return false
				}
				if order == "desc" {
					return left > right
				}
				return left < right
			default:
				return false
			}
		})
	}

	return filtered, nil
}

// PrefilterRooms applies non-status filters only and keeps input ordering.
// It is useful before expensive status enrichment.
func PrefilterRooms(rooms []Room, filters RoomFilters) ([]Room, error) {
	if err := ValidateRoomFilters(filters); err != nil {
		return nil, err
	}

	return filterRooms(rooms, filters, false), nil
}

// ValidateRoomFilters validates room filters.
//
// Summary:
// - Validates room filters.
//
// Attributes:
// - filters (RoomFilters): Input parameter.
//
// Returns:
// - value1 (error): Returned value.
func ValidateRoomFilters(filters RoomFilters) error {
	status := normalizedStringPointer(filters.Status)
	if status != "" &&
		status != RoomStatusAvailable &&
		status != RoomStatusOccupied &&
		status != RoomStatusUpcoming &&
		status != RoomStatusMaintenance {
		return &InvalidParameterError{
			Parameter: "status",
		}
	}

	if filters.CapacityMin != nil && filters.CapacityMax != nil && *filters.CapacityMin > *filters.CapacityMax {
		return &InvalidParameterError{
			Parameter: "capacity_min",
		}
	}

	sortField := normalizedStringPointer(filters.Sort)
	if sortField != "" {
		if sortField != "name" && sortField != "capacity" && sortField != "status" {
			return &InvalidParameterError{
				Parameter: "sort",
			}
		}
	}

	order := normalizedStringPointer(filters.Order)
	if order != "" && order != "asc" && order != "desc" {
		return &InvalidParameterError{
			Parameter: "order",
		}
	}

	return nil
}

// normalizedStringPointer normalizeds string pointer.
//
// Summary:
// - Normalizeds string pointer.
//
// Attributes:
// - value (*string): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
func normalizedStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*value))
}

// filterRooms filters rooms.
//
// Summary:
// - Filters rooms.
//
// Attributes:
// - rooms ([]Room): Input parameter.
// - filters (RoomFilters): Input parameter.
// - includeStatus (bool): Input parameter.
//
// Returns:
// - value1 ([]Room): Returned value.
func filterRooms(rooms []Room, filters RoomFilters, includeStatus bool) []Room {
	filtered := make([]Room, 0, len(rooms))

	for _, room := range rooms {
		if filters.Building != nil && room.Building != *filters.Building {
			continue
		}
		if filters.Floor != nil && room.Floor != *filters.Floor {
			continue
		}
		if filters.Type != nil && room.Type != *filters.Type {
			continue
		}
		if includeStatus && filters.Status != nil && room.Status != *filters.Status {
			continue
		}
		if filters.CapacityMin != nil && room.Capacity < *filters.CapacityMin {
			continue
		}
		if filters.CapacityMax != nil && room.Capacity > *filters.CapacityMax {
			continue
		}

		filtered = append(filtered, cloneRoom(room))
	}

	return filtered
}

// cloneRoom clones room.
//
// Summary:
// - Clones room.
//
// Attributes:
// - room (Room): Input parameter.
//
// Returns:
// - value1 (Room): Returned value.
func cloneRoom(room Room) Room {
	cloned := room

	if room.CurrentEvent != nil {
		current := *room.CurrentEvent
		cloned.CurrentEvent = &current
	}
	if room.NextEvent != nil {
		next := *room.NextEvent
		cloned.NextEvent = &next
	}

	return cloned
}
