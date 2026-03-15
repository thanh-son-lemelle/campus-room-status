package mockdata

type DirectoryRoom struct {
	ResourceName     string
	ResourceEmail    string
	Capacity         int
	ResourceType     string
	ResourceCategory string
}

func DirectoryRoomAmphiA() DirectoryRoom {
	return DirectoryRoom{
		ResourceName:  "AMPHI-A",
		ResourceEmail: "amphi-a@example.org",
	}
}

func DirectoryRoomAmphiAMaintenance() DirectoryRoom {
	room := DirectoryRoomAmphiA()
	room.ResourceCategory = "maintenance"
	return room
}
