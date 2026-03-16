package mockdata

type InventorySnapshot struct {
	Buildings []Building
	Rooms     []Room
}

// InventorySnapshotFixture inventories snapshot fixture.
//
// Summary:
// - Inventories snapshot fixture.
//
// Attributes:
// - buildingID (string): Input parameter.
// - roomCode (string): Input parameter.
//
// Returns:
// - value1 (InventorySnapshot): Returned value.
func InventorySnapshotFixture(buildingID string, roomCode string) InventorySnapshot {
	return InventorySnapshot{
		Buildings: []Building{BuildingFromID(buildingID)},
		Rooms:     []Room{RoomFromBuildingAndCode(buildingID, roomCode)},
	}
}
