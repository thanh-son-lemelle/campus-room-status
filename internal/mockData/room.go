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

// RoomAmphiA rooms amphi a.
//
// Summary:
// - Rooms amphi a.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Room): Returned value.
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

// RoomLab204 rooms lab 204.
//
// Summary:
// - Rooms lab 204.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Room): Returned value.
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

// RoomLab101 rooms lab 101.
//
// Summary:
// - Rooms lab 101.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Room): Returned value.
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

// RoomR1 rooms r 1.
//
// Summary:
// - Rooms r 1.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Room): Returned value.
func RoomR1() Room {
	return Room{
		Code:          "R1",
		ResourceEmail: "r1@example.org",
		Building:      "B1",
		Type:          "lab",
		Capacity:      20,
	}
}

// RoomR2 rooms r 2.
//
// Summary:
// - Rooms r 2.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Room): Returned value.
func RoomR2() Room {
	return Room{
		Code:          "R2",
		ResourceEmail: "r2@example.org",
		Building:      "B1",
		Type:          "lab",
		Capacity:      40,
	}
}

// RoomR3 rooms r 3.
//
// Summary:
// - Rooms r 3.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Room): Returned value.
func RoomR3() Room {
	return Room{
		Code:          "R3",
		ResourceEmail: "r3@example.org",
		Building:      "B2",
		Type:          "lab",
		Capacity:      50,
	}
}

// RoomWithResourceEmail rooms with resource email.
//
// Summary:
// - Rooms with resource email.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (Room): Returned value.
func RoomWithResourceEmail() Room {
	room := RoomAmphiA()
	room.ResourceEmail = "amphi-a@example.org"
	return room
}

// RoomFromBuildingAndCode rooms from building and code.
//
// Summary:
// - Rooms from building and code.
//
// Attributes:
// - buildingID (string): Input parameter.
// - roomCode (string): Input parameter.
//
// Returns:
// - value1 (Room): Returned value.
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

// RoomsForRoomService roomses for room service.
//
// Summary:
// - Roomses for room service.
//
// Attributes:
// - None.
//
// Returns:
// - value1 ([]Room): Returned value.
func RoomsForRoomService() []Room {
	return []Room{RoomAmphiA(), RoomLab204(), RoomLab101()}
}

// RoomsForFilterAndSort roomses for filter and sort.
//
// Summary:
// - Roomses for filter and sort.
//
// Attributes:
// - None.
//
// Returns:
// - value1 ([]Room): Returned value.
func RoomsForFilterAndSort() []Room {
	return []Room{RoomAmphiA(), RoomLab204(), RoomLab101()}
}

// RoomsForPrefilter roomses for prefilter.
//
// Summary:
// - Roomses for prefilter.
//
// Attributes:
// - None.
//
// Returns:
// - value1 ([]Room): Returned value.
func RoomsForPrefilter() []Room {
	return []Room{RoomR1(), RoomR2(), RoomR3()}
}
