package mockdata

type Building struct {
	ID      string
	Name    string
	Address string
	Floors  []int
}

func BuildingB1() Building {
	return Building{
		ID:      "B1",
		Name:    "Building A",
		Address: "1 Campus Street",
		Floors:  []int{0, 1, 2},
	}
}

func BuildingB2() Building {
	return Building{
		ID:      "B2",
		Name:    "Building B",
		Address: "2 Campus Street",
		Floors:  []int{0, 1, 2, 3},
	}
}

func BuildingFromID(buildingID string) Building {
	return Building{
		ID:      buildingID,
		Name:    "Building " + buildingID,
		Address: "1 Campus Street",
		Floors:  []int{0, 1, 2},
	}
}
