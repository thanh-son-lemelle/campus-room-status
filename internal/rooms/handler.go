package rooms

import (
	"net/http"
	"strconv"
	"time"

	"campus-room-status/internal/api"
	"campus-room-status/internal/domain"
	"github.com/gin-gonic/gin"
)

var scheduleFixture = []api.EventResponse{
	{
		Title:     "Advanced Networks",
		Start:     time.Date(2026, time.March, 9, 9, 0, 0, 0, time.UTC),
		End:       time.Date(2026, time.March, 9, 11, 0, 0, 0, time.UTC),
		Organizer: "IT Department",
	},
	{
		Title:     "Distributed Systems",
		Start:     time.Date(2026, time.March, 9, 14, 0, 0, 0, time.UTC),
		End:       time.Date(2026, time.March, 9, 16, 0, 0, 0, time.UTC),
		Organizer: "Engineering Office",
	},
}

var nextEventFixture = &api.EventResponse{
	Title:     "Capstone Review",
	Start:     time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC),
	End:       time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
	Organizer: "Academic Board",
}

var currentEventFixture = &api.EventResponse{
	Title:     "OS Lab Session",
	Start:     time.Date(2026, time.March, 9, 10, 0, 0, 0, time.UTC),
	End:       time.Date(2026, time.March, 9, 12, 0, 0, 0, time.UTC),
	Organizer: "Systems Team",
}

var roomsFixture = []api.RoomResponse{
	{
		Code:         "AMPHI-A",
		Name:         "Amphitheater A",
		Building:     "B1",
		Floor:        1,
		Capacity:     180,
		Type:         "amphitheater",
		Status:       "available",
		CurrentEvent: nil,
		NextEvent:    nextEventFixture,
	},
	{
		Code:         "LAB-204",
		Name:         "Computer Lab 204",
		Building:     "B2",
		Floor:        2,
		Capacity:     30,
		Type:         "lab",
		Status:       "occupied",
		CurrentEvent: currentEventFixture,
		NextEvent:    nil,
	},
}

func ListHandler(c *gin.Context) {
	responseFilters := make(map[string]any)
	queryFilters := domain.RoomFilters{}

	if building := c.Query("building"); building != "" {
		queryFilters.Building = &building
		responseFilters["building"] = building
	}

	if roomType := c.Query("type"); roomType != "" {
		queryFilters.Type = &roomType
		responseFilters["type"] = roomType
	}

	if status := c.Query("status"); status != "" {
		queryFilters.Status = &status
		responseFilters["status"] = status
	}

	if raw := c.Query("capacity_min"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			api.WriteError(c, &domain.InvalidParameterError{
				Parameter: "capacity_min",
				Value:     raw,
			})
			return
		}

		queryFilters.CapacityMin = &parsed
		responseFilters["capacity_min"] = parsed
	}

	if raw := c.Query("capacity_max"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			api.WriteError(c, &domain.InvalidParameterError{
				Parameter: "capacity_max",
				Value:     raw,
			})
			return
		}

		queryFilters.CapacityMax = &parsed
		responseFilters["capacity_max"] = parsed
	}

	if sortField := c.Query("sort"); sortField != "" {
		queryFilters.Sort = &sortField
		responseFilters["sort"] = sortField
	}

	if order := c.Query("order"); order != "" {
		queryFilters.Order = &order
		responseFilters["order"] = order
	}

	rooms := make([]domain.Room, len(roomsFixture))
	for i := range roomsFixture {
		rooms[i] = apiRoomToDomainRoom(roomsFixture[i])
	}

	filteredDomainRooms, err := domain.FilterAndSortRooms(rooms, queryFilters)
	if err != nil {
		api.WriteError(c, err)
		return
	}

	filteredRooms := make([]api.RoomResponse, len(filteredDomainRooms))
	for i := range filteredDomainRooms {
		filteredRooms[i] = domainRoomToAPIRoom(filteredDomainRooms[i])
	}

	c.JSON(http.StatusOK, api.RoomsListResponse{
		Timestamp: time.Now().UTC(),
		Filters:   responseFilters,
		Count:     len(filteredRooms),
		Rooms:     filteredRooms,
	})
}

func DetailHandler(c *gin.Context) {
	room, err := roomByCode(c.Param("code"))
	if err != nil {
		api.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.RoomDetailResponse{
		Code:          room.Code,
		Name:          room.Name,
		Building:      room.Building,
		Floor:         room.Floor,
		Capacity:      room.Capacity,
		Type:          room.Type,
		Status:        room.Status,
		CurrentEvent:  room.CurrentEvent,
		NextEvent:     room.NextEvent,
		ScheduleToday: scheduleFixture,
	})
}

func ScheduleHandler(c *gin.Context) {
	roomCode := c.Param("code")
	if roomCode == "SVC-UNAVAILABLE" {
		api.WriteError(c, &domain.ServiceUnavailableError{Service: "google"})
		return
	}

	if _, err := roomByCode(roomCode); err != nil {
		api.WriteError(c, err)
		return
	}

	if raw := c.Query("start"); raw != "" {
		if _, err := time.Parse(time.RFC3339, raw); err != nil {
			api.WriteError(c, &domain.InvalidParameterError{
				Parameter: "start",
				Value:     raw,
			})
			return
		}
	}

	if raw := c.Query("end"); raw != "" {
		if _, err := time.Parse(time.RFC3339, raw); err != nil {
			api.WriteError(c, &domain.InvalidParameterError{
				Parameter: "end",
				Value:     raw,
			})
			return
		}
	}

	c.JSON(http.StatusOK, api.RoomScheduleResponse{
		RoomCode: roomCode,
		Period: api.PeriodResponse{
			Start: c.Query("start"),
			End:   c.Query("end"),
		},
		Events: scheduleFixture,
	})
}

func roomByCode(code string) (*api.RoomResponse, error) {
	for i := range roomsFixture {
		if roomsFixture[i].Code == code {
			return &roomsFixture[i], nil
		}
	}

	return nil, &domain.RoomNotFoundError{RoomCode: code}
}

func apiRoomToDomainRoom(room api.RoomResponse) domain.Room {
	return domain.Room{
		Code:         room.Code,
		Name:         room.Name,
		Building:     room.Building,
		Floor:        room.Floor,
		Capacity:     room.Capacity,
		Type:         room.Type,
		Status:       room.Status,
		CurrentEvent: apiEventToDomainEvent(room.CurrentEvent),
		NextEvent:    apiEventToDomainEvent(room.NextEvent),
	}
}

func domainRoomToAPIRoom(room domain.Room) api.RoomResponse {
	return api.RoomResponse{
		Code:         room.Code,
		Name:         room.Name,
		Building:     room.Building,
		Floor:        room.Floor,
		Capacity:     room.Capacity,
		Type:         room.Type,
		Status:       room.Status,
		CurrentEvent: domainEventToAPIEvent(room.CurrentEvent),
		NextEvent:    domainEventToAPIEvent(room.NextEvent),
	}
}

func apiEventToDomainEvent(event *api.EventResponse) *domain.Event {
	if event == nil {
		return nil
	}

	return &domain.Event{
		Title:     event.Title,
		Start:     event.Start,
		End:       event.End,
		Organizer: event.Organizer,
	}
}

func domainEventToAPIEvent(event *domain.Event) *api.EventResponse {
	if event == nil {
		return nil
	}

	return &api.EventResponse{
		Title:     event.Title,
		Start:     event.Start,
		End:       event.End,
		Organizer: event.Organizer,
	}
}
