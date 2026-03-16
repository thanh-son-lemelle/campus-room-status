package mockdata

type Building struct {
	ID      string
	Name    string
	Address string
	Floors  []string
}

// BuildingB1 buildings b 1.
//
// Summary:
// - Buildings b 1.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Building): Returned value.
func BuildingB1() Building {
	return Building{
		ID:      "B1",
		Name:    "Building A",
		Address: "1 Campus Street",
		Floors:  []string{"0", "1", "2"},
	}
}

// BuildingB2 buildings b 2.
//
// Summary:
// - Buildings b 2.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Building): Returned value.
func BuildingB2() Building {
	return Building{
		ID:      "B2",
		Name:    "Building B",
		Address: "2 Campus Street",
		Floors:  []string{"0", "1", "2", "3"},
	}
}

// BuildingFromID buildings from id.
//
// Summary:
// - Buildings from id.
//
// Attributes:
// - buildingID (string): Input parameter.
//
// Returns:
// - value1 (Building): Returned value.
func BuildingFromID(buildingID string) Building {
	return Building{
		ID:      buildingID,
		Name:    "Building " + buildingID,
		Address: "1 Campus Street",
		Floors:  []string{"0", "1", "2"},
	}
}
