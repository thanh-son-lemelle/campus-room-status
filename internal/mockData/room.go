package mockdata

type Room struct {
	Code          string
	ResourceEmail string
	Name          string
	Building      string
	Floor         int
	Capacity      int
	Type          string
	Status        string
}

func RoomAmphiA() Room {
	return Room{
		Code:     "AMPHI-A",
		Name:     "Amphitheater A",
		Building: "B1",
		Floor:    1,
		Capacity: 180,
		Type:     "amphitheater",
		Status:   "available",
	}
}

func RoomLab204() Room {
	return Room{
		Code:     "LAB-204",
		Name:     "Computer Lab 204",
		Building: "B2",
		Floor:    2,
		Capacity: 30,
		Type:     "lab",
		Status:   "occupied",
	}
}

func RoomLab101() Room {
	return Room{
		Code:     "LAB-101",
		Name:     "Alpha Lab",
		Building: "B1",
		Floor:    1,
		Capacity: 220,
		Type:     "lab",
		Status:   "available",
	}
}

func RoomR1() Room {
	return Room{
		Code:          "R1",
		ResourceEmail: "r1@example.org",
		Building:      "B1",
		Type:          "lab",
		Capacity:      20,
	}
}

func RoomR2() Room {
	return Room{
		Code:          "R2",
		ResourceEmail: "r2@example.org",
		Building:      "B1",
		Type:          "lab",
		Capacity:      40,
	}
}

func RoomR3() Room {
	return Room{
		Code:          "R3",
		ResourceEmail: "r3@example.org",
		Building:      "B2",
		Type:          "lab",
		Capacity:      50,
	}
}

func RoomWithResourceEmail() Room {
	room := RoomAmphiA()
	room.ResourceEmail = "amphi-a@example.org"
	return room
}

func RoomFromBuildingAndCode(buildingID string, roomCode string) Room {
	return Room{
		Code:     roomCode,
		Name:     "Room " + roomCode,
		Building: buildingID,
		Floor:    1,
		Capacity: 30,
		Type:     "lab",
		Status:   "available",
	}
}

func RoomsForRoomService() []Room {
	return []Room{RoomAmphiA(), RoomLab204(), RoomLab101()}
}

func RoomsForFilterAndSort() []Room {
	return []Room{RoomAmphiA(), RoomLab204(), RoomLab101()}
}

func RoomsForPrefilter() []Room {
	return []Room{RoomR1(), RoomR2(), RoomR3()}
}
