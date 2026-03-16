package adminsdk

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"campus-room-status/internal/domain"
	admin "google.golang.org/api/admin/directory/v1"
)

// listBuildings lists buildings.
//
// Summary:
// - Lists buildings.
//
// Attributes:
// - ctx (context.Context): Input parameter.
//
// Returns:
// - value1 ([]domain.Building): Returned value.
// - value2 (error): Returned value.
func (s *InventorySource) listBuildings(ctx context.Context) ([]domain.Building, error) {
	pageToken := ""
	visitedTokens := make(map[string]struct{})
	buildingsByID := make(map[string]domain.Building)

	for {
		call := s.service.Resources.Buildings.List(s.customer).MaxResults(s.pageSize)
		if pageToken != "" {
			if _, exists := visitedTokens[pageToken]; exists {
				return nil, fmt.Errorf("detected repeated page token %q on buildings endpoint", pageToken)
			}
			visitedTokens[pageToken] = struct{}{}
			call = call.PageToken(pageToken)
		}

		response, err := call.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("buildings list request failed: %w", err)
		}
		if response == nil {
			break
		}

		for _, building := range response.Buildings {
			mapped, ok := mapBuilding(building)
			if !ok {
				continue
			}
			buildingsByID[mapped.ID] = mapped
		}

		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	buildings := make([]domain.Building, 0, len(buildingsByID))
	for _, building := range buildingsByID {
		buildings = append(buildings, building)
	}
	sort.SliceStable(buildings, func(i, j int) bool {
		return buildings[i].ID < buildings[j].ID
	})

	return buildings, nil
}

// mapBuilding maps building.
//
// Summary:
// - Maps building.
//
// Attributes:
// - building (*admin.Building): Input parameter.
//
// Returns:
// - value1 (domain.Building): Returned value.
// - value2 (bool): Returned value.
func mapBuilding(building *admin.Building) (domain.Building, bool) {
	if building == nil {
		return domain.Building{}, false
	}

	id := strings.TrimSpace(building.BuildingId)
	if id == "" {
		return domain.Building{}, false
	}

	name := firstNonEmpty(building.BuildingName, id)
	floors := normalizeFloorNames(building.FloorNames)
	address := mapBuildingAddress(building.Address)

	return domain.Building{
		ID:      id,
		Name:    name,
		Address: address,
		Floors:  floors,
	}, true
}

// mapBuildingAddress maps building address.
//
// Summary:
// - Maps building address.
//
// Attributes:
// - address (*admin.BuildingAddress): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
func mapBuildingAddress(address *admin.BuildingAddress) string {
	if address == nil {
		return ""
	}

	parts := make([]string, 0, 8)
	parts = append(parts, trimNonEmpty(address.AddressLines)...)
	parts = appendIfNotEmpty(parts, address.Locality)
	parts = appendIfNotEmpty(parts, address.AdministrativeArea)
	parts = appendIfNotEmpty(parts, address.PostalCode)
	parts = appendIfNotEmpty(parts, address.RegionCode)

	return strings.Join(parts, ", ")
}

// deriveBuildingsFromRooms derives buildings from rooms.
//
// Summary:
// - Derives buildings from rooms.
//
// Attributes:
// - rooms ([]domain.Room): Input parameter.
//
// Returns:
// - value1 ([]domain.Building): Returned value.
func deriveBuildingsFromRooms(rooms []domain.Room) []domain.Building {
	byID := make(map[string]*domain.Building)

	for _, room := range rooms {
		if strings.TrimSpace(room.Building) == "" {
			continue
		}

		building, exists := byID[room.Building]
		if !exists {
			building = &domain.Building{
				ID:   room.Building,
				Name: room.Building,
			}
			byID[room.Building] = building
		}

		if room.Floor != 0 {
			floor := strconv.Itoa(room.Floor)
			if !slicesContainsString(building.Floors, floor) {
				building.Floors = append(building.Floors, floor)
			}
		}
	}

	buildings := make([]domain.Building, 0, len(byID))
	for _, building := range byID {
		sort.SliceStable(building.Floors, func(i, j int) bool {
			left, leftErr := strconv.Atoi(building.Floors[i])
			right, rightErr := strconv.Atoi(building.Floors[j])
			if leftErr == nil && rightErr == nil {
				return left < right
			}
			if leftErr == nil {
				return true
			}
			if rightErr == nil {
				return false
			}
			return building.Floors[i] < building.Floors[j]
		})
		buildings = append(buildings, *building)
	}
	sort.SliceStable(buildings, func(i, j int) bool {
		return buildings[i].ID < buildings[j].ID
	})

	return buildings
}

// normalizeFloorNames normalizes floor names.
//
// Summary:
// - Normalizes floor names.
//
// Attributes:
// - floorNames ([]string): Input parameter.
//
// Returns:
// - value1 ([]string): Returned value.
func normalizeFloorNames(floorNames []string) []string {
	if len(floorNames) == 0 {
		return nil
	}

	floors := make([]string, 0, len(floorNames))
	for _, floorName := range floorNames {
		value := strings.TrimSpace(floorName)
		if value == "" {
			continue
		}
		if !slicesContainsString(floors, value) {
			floors = append(floors, value)
		}
	}

	return floors
}
