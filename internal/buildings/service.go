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

func NewService(inventory inventoryReader) domain.BuildingService {
	return &service{inventory: inventory}
}

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
			Floors:  append([]int(nil), buildings[i].Floors...),
		}
	}

	return out
}
