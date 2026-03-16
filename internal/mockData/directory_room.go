package mockdata

type DirectoryRoom struct {
	ResourceName     string
	ResourceEmail    string
	Capacity         int
	ResourceType     string
	ResourceCategory string
}

// DirectoryRoomAmphiA directories room amphi a.
//
// Summary:
// - Directories room amphi a.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (DirectoryRoom): Returned value.
func DirectoryRoomAmphiA() DirectoryRoom {
	return DirectoryRoom{
		ResourceName:  "AMPHI-A",
		ResourceEmail: "amphi-a@example.org",
	}
}

// DirectoryRoomAmphiAMaintenance directories room amphi a maintenance.
//
// Summary:
// - Directories room amphi a maintenance.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (DirectoryRoom): Returned value.
func DirectoryRoomAmphiAMaintenance() DirectoryRoom {
	room := DirectoryRoomAmphiA()
	room.ResourceCategory = "maintenance"
	return room
}
