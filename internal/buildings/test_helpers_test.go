package buildings

import (
	"campus-room-status/internal/domain"
	mockdata "campus-room-status/internal/mockData"
)

func domainBuildingFromMock(building mockdata.Building) domain.Building {
	return domain.Building{
		ID:      building.ID,
		Name:    building.Name,
		Address: building.Address,
		Floors:  append([]string(nil), building.Floors...),
	}
}
