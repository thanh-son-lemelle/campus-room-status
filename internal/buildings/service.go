package buildings

import (
	"context"
	"errors"

	"campus-room-status/internal/domain"
)

type inventoryReader interface {
	GetInventory(ctx context.Context) (domain.InventorySnapshot, error)
}

type service struct {
	inventory inventoryReader
}

var _ domain.BuildingService = (*service)(nil)

// NewService creates a new service.
//
// Summary:
// - Creates a new service.
//
// Attributes:
// - inventory (inventoryReader): Input parameter.
//
// Returns:
// - value1 (domain.BuildingService): Returned value.
func NewService(inventory inventoryReader) domain.BuildingService {
	return &service{inventory: inventory}
}

// ListBuildings lists buildings.
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
func (s *service) ListBuildings(ctx context.Context) ([]domain.Building, error) {
	if s.inventory == nil {
		return nil, errors.New("inventory cache is required")
	}

	snapshot, err := s.inventory.GetInventory(ctx)
	if err != nil {
		return nil, err
	}

	return cloneBuildings(snapshot.Buildings), nil
}

// cloneBuildings clones buildings.
//
// Summary:
// - Clones buildings.
//
// Attributes:
// - buildings ([]domain.Building): Input parameter.
//
// Returns:
// - value1 ([]domain.Building): Returned value.
func cloneBuildings(buildings []domain.Building) []domain.Building {
	if buildings == nil {
		return nil
	}

	out := make([]domain.Building, len(buildings))
	for i := range buildings {
		out[i] = domain.Building{
			ID:      buildings[i].ID,
			Name:    buildings[i].Name,
			Address: buildings[i].Address,
			Floors:  append([]string(nil), buildings[i].Floors...),
		}
	}

	return out
}
