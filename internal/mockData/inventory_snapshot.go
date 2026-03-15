package mockdata

type InventorySnapshot struct {
	Buildings []Building
	Rooms     []Room
}

func InventorySnapshotFixture(buildingID string, roomCode string) InventorySnapshot {
	return InventorySnapshot{
		Buildings: []Building{BuildingFromID(buildingID)},
		Rooms:     []Room{RoomFromBuildingAndCode(buildingID, roomCode)},
	}
}
