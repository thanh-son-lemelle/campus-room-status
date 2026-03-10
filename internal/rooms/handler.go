package rooms

import (
	"net/http"
	"sort"
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
	filters := make(map[string]any)

	building := c.Query("building")
	if building != "" {
		filters["building"] = building
	}

	roomType := c.Query("type")
	if roomType != "" {
		filters["type"] = roomType
	}

	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}

	var capacityMin *int
	if raw := c.Query("capacity_min"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			api.WriteError(c, &domain.InvalidParameterError{
				Parameter: "capacity_min",
				Value:     raw,
			})
			return
		}

		capacityMin = &parsed
		filters["capacity_min"] = parsed
	}

	var capacityMax *int
	if raw := c.Query("capacity_max"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			api.WriteError(c, &domain.InvalidParameterError{
				Parameter: "capacity_max",
				Value:     raw,
			})
			return
		}

		capacityMax = &parsed
		filters["capacity_max"] = parsed
	}

	sortField := c.Query("sort")
	if sortField != "" {
		filters["sort"] = sortField
	}

	order := c.Query("order")
	if order != "" {
		filters["order"] = order
	}

	filteredRooms := make([]api.RoomResponse, 0, len(roomsFixture))
	for _, room := range roomsFixture {
		if building != "" && room.Building != building {
			continue
		}
		if roomType != "" && room.Type != roomType {
			continue
		}
		if status != "" && room.Status != status {
			continue
		}
		if capacityMin != nil && room.Capacity < *capacityMin {
			continue
		}
		if capacityMax != nil && room.Capacity > *capacityMax {
			continue
		}
		filteredRooms = append(filteredRooms, room)
	}

	if sortField == "capacity" {
		sort.Slice(filteredRooms, func(i, j int) bool {
			if order == "desc" {
				return filteredRooms[i].Capacity > filteredRooms[j].Capacity
			}
			return filteredRooms[i].Capacity < filteredRooms[j].Capacity
		})
	}

	c.JSON(http.StatusOK, api.RoomsListResponse{
		Timestamp: time.Now().UTC(),
		Filters:   filters,
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
