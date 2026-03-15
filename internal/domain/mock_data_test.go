package domain

import mockdata "campus-room-status/internal/mockData"

func domainBuildingFromMock(building mockdata.Building) Building {
	return Building{
		ID:      building.ID,
		Name:    building.Name,
		Address: building.Address,
		Floors:  append([]int(nil), building.Floors...),
	}
}

func domainRoomFromMock(room mockdata.Room) Room {
	return Room{
		Code:          room.Code,
		ResourceEmail: room.ResourceEmail,
		Name:          room.Name,
		Building:      room.Building,
		Floor:         room.Floor,
		Capacity:      room.Capacity,
		Type:          room.Type,
		Status:        room.Status,
	}
}

func domainRoomsFromMock(rooms []mockdata.Room) []Room {
	out := make([]Room, len(rooms))
	for i := range rooms {
		out[i] = domainRoomFromMock(rooms[i])
	}
	return out
}

func domainEventFromMock(event mockdata.Event) Event {
	return Event{
		Title:     event.Title,
		Start:     event.Start,
		End:       event.End,
		Organizer: event.Organizer,
	}
}

func domainEventsFromMock(events []mockdata.Event) []Event {
	out := make([]Event, len(events))
	for i := range events {
		out[i] = domainEventFromMock(events[i])
	}
	return out
}

func domainDirectoryRoomFromMock(room mockdata.DirectoryRoom) DirectoryRoom {
	return DirectoryRoom{
		ResourceName:     room.ResourceName,
		ResourceEmail:    room.ResourceEmail,
		Capacity:         room.Capacity,
		ResourceType:     room.ResourceType,
		ResourceCategory: room.ResourceCategory,
	}
}

func domainInventorySnapshotFromMock(snapshot mockdata.InventorySnapshot) InventorySnapshot {
	buildings := make([]Building, len(snapshot.Buildings))
	for i := range snapshot.Buildings {
		buildings[i] = domainBuildingFromMock(snapshot.Buildings[i])
	}

	return InventorySnapshot{
		Buildings: buildings,
		Rooms:     domainRoomsFromMock(snapshot.Rooms),
	}
}
