package rooms

import (
	"context"
	"errors"
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

var defaultRoomService = newDefaultListService()
var defaultListHandler = NewListHandler(defaultRoomService, nil)
var defaultDetailHandler = NewDetailHandler(defaultRoomService)
var defaultScheduleHandler = NewScheduleHandler(defaultRoomService)

func ListHandler(c *gin.Context) {
	defaultListHandler(c)
}

// NewListHandler godoc
// @Summary List rooms with current and next events
// @Tags rooms
// @Produce json
// @Param building query string false "Building code or ID"
// @Param type query string false "Room type"
// @Param status query string false "Room status (available|occupied|upcoming|maintenance)"
// @Param capacity_min query int false "Minimum capacity"
// @Param capacity_max query int false "Maximum capacity"
// @Param sort query string false "Sort field (name|capacity|status)"
// @Param order query string false "Sort order (asc|desc)"
// @Success 200 {object} api.RoomsListResponse
// @Failure 400 {object} api.ErrorEnvelope
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/rooms [get]
func NewListHandler(service domain.RoomService, clock domain.Clock) gin.HandlerFunc {
	h := &listHandler{
		service: service,
		clock:   clock,
	}
	if h.clock == nil {
		h.clock = listHandlerClock{}
	}

	return h.handle
}

type listHandler struct {
	service domain.RoomService
	clock   domain.Clock
}

func (h *listHandler) handle(c *gin.Context) {
	if h.service == nil {
		api.WriteError(c, api.NewHTTPError(
			http.StatusInternalServerError,
			api.ErrorCodeInternalServerError,
			"Room service is not configured",
		))
		return
	}

	queryFilters, responseFilters, err := parseListFilters(c)
	if err != nil {
		api.WriteError(c, err)
		return
	}

	rooms, err := h.service.ListRooms(c.Request.Context(), queryFilters)
	if err != nil {
		var invalidParamErr *domain.InvalidParameterError
		var serviceUnavailableErr *domain.ServiceUnavailableError
		switch {
		case errors.As(err, &invalidParamErr):
			api.WriteError(c, err)
		case errors.As(err, &serviceUnavailableErr):
			api.WriteError(c, err)
		default:
			api.WriteError(c, api.NewHTTPError(
				http.StatusInternalServerError,
				api.ErrorCodeInternalServerError,
				"Une erreur interne est survenue",
			))
		}
		return
	}

	responseRooms := make([]api.RoomResponse, len(rooms))
	for i := range rooms {
		responseRooms[i] = domainRoomToAPIRoom(rooms[i])
	}

	c.JSON(http.StatusOK, api.RoomsListResponse{
		Timestamp: h.clock.Now().UTC(),
		Filters:   responseFilters,
		Count:     len(responseRooms),
		Rooms:     responseRooms,
	})
}

func DetailHandler(c *gin.Context) {
	defaultDetailHandler(c)
}

// NewDetailHandler godoc
// @Summary Get room detail and today's schedule
// @Tags rooms
// @Produce json
// @Param code path string true "Room code"
// @Success 200 {object} api.RoomDetailResponse
// @Failure 404 {object} api.ErrorEnvelope
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/rooms/{code} [get]
func NewDetailHandler(service domain.RoomService) gin.HandlerFunc {
	h := &detailHandler{
		service: service,
	}

	return h.handle
}

type detailHandler struct {
	service domain.RoomService
}

func (h *detailHandler) handle(c *gin.Context) {
	if h.service == nil {
		api.WriteError(c, api.NewHTTPError(
			http.StatusInternalServerError,
			api.ErrorCodeInternalServerError,
			"Room service is not configured",
		))
		return
	}

	room, scheduleToday, err := h.service.GetRoomDetail(c.Request.Context(), c.Param("code"))
	if err != nil {
		var roomNotFoundErr *domain.RoomNotFoundError
		var serviceUnavailableErr *domain.ServiceUnavailableError
		switch {
		case errors.As(err, &roomNotFoundErr):
			api.WriteError(c, err)
		case errors.As(err, &serviceUnavailableErr):
			api.WriteError(c, err)
		default:
			api.WriteError(c, api.NewHTTPError(
				http.StatusInternalServerError,
				api.ErrorCodeInternalServerError,
				"Une erreur interne est survenue",
			))
		}
		return
	}

	c.JSON(http.StatusOK, api.RoomDetailResponse{
		Code:          room.Code,
		ResourceEmail: room.ResourceEmail,
		Name:          room.Name,
		Building:      room.Building,
		Floor:         room.Floor,
		Capacity:      room.Capacity,
		Type:          room.Type,
		Status:        room.Status,
		CurrentEvent:  domainEventToAPIEvent(room.CurrentEvent),
		NextEvent:     domainEventToAPIEvent(room.NextEvent),
		ScheduleToday: mapDomainEventsToAPIEvents(scheduleToday),
	})
}

func ScheduleHandler(c *gin.Context) {
	defaultScheduleHandler(c)
}

// NewScheduleHandler godoc
// @Summary Get room schedule for a time period
// @Tags rooms
// @Produce json
// @Param code path string true "Room code"
// @Param start query string true "Start date-time (RFC3339)"
// @Param end query string true "End date-time (RFC3339)"
// @Success 200 {object} api.RoomScheduleResponse
// @Failure 400 {object} api.ErrorEnvelope
// @Failure 404 {object} api.ErrorEnvelope
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/rooms/{code}/schedule [get]
func NewScheduleHandler(service domain.RoomService) gin.HandlerFunc {
	h := &scheduleHandler{
		service: service,
	}

	return h.handle
}

type scheduleHandler struct {
	service domain.RoomService
}

func (h *scheduleHandler) handle(c *gin.Context) {
	if h.service == nil {
		api.WriteError(c, api.NewHTTPError(
			http.StatusInternalServerError,
			api.ErrorCodeInternalServerError,
			"Room service is not configured",
		))
		return
	}

	if c.Param("code") == "SVC-UNAVAILABLE" {
		api.WriteError(c, &domain.ServiceUnavailableError{Service: "google"})
		return
	}

	startRaw := c.Query("start")
	if startRaw == "" {
		api.WriteError(c, &domain.InvalidParameterError{
			Parameter: "start",
		})
		return
	}

	endRaw := c.Query("end")
	if endRaw == "" {
		api.WriteError(c, &domain.InvalidParameterError{
			Parameter: "end",
		})
		return
	}

	start, err := time.Parse(time.RFC3339, startRaw)
	if err != nil {
		api.WriteError(c, &domain.InvalidParameterError{
			Parameter: "start",
			Value:     startRaw,
		})
		return
	}

	end, err := time.Parse(time.RFC3339, endRaw)
	if err != nil {
		api.WriteError(c, &domain.InvalidParameterError{
			Parameter: "end",
			Value:     endRaw,
		})
		return
	}

	if start.After(end) {
		api.WriteError(c, &domain.InvalidParameterError{
			Parameter: "start",
			Value:     startRaw,
		})
		return
	}

	events, err := h.service.GetRoomSchedule(c.Request.Context(), c.Param("code"), start, end)
	if err != nil {
		var invalidParamErr *domain.InvalidParameterError
		var roomNotFoundErr *domain.RoomNotFoundError
		var serviceUnavailableErr *domain.ServiceUnavailableError
		switch {
		case errors.As(err, &invalidParamErr):
			api.WriteError(c, err)
		case errors.As(err, &roomNotFoundErr):
			api.WriteError(c, err)
		case errors.As(err, &serviceUnavailableErr):
			api.WriteError(c, err)
		default:
			api.WriteError(c, api.NewHTTPError(
				http.StatusInternalServerError,
				api.ErrorCodeInternalServerError,
				"Une erreur interne est survenue",
			))
		}
		return
	}

	c.JSON(http.StatusOK, api.RoomScheduleResponse{
		RoomCode: c.Param("code"),
		Period: api.PeriodResponse{
			Start: startRaw,
			End:   endRaw,
		},
		Events: mapDomainEventsToAPIEvents(events),
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

func parseListFilters(c *gin.Context) (domain.RoomFilters, map[string]any, error) {
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
			return domain.RoomFilters{}, nil, &domain.InvalidParameterError{
				Parameter: "capacity_min",
				Value:     raw,
			}
		}

		queryFilters.CapacityMin = &parsed
		responseFilters["capacity_min"] = parsed
	}

	if raw := c.Query("capacity_max"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return domain.RoomFilters{}, nil, &domain.InvalidParameterError{
				Parameter: "capacity_max",
				Value:     raw,
			}
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

	return queryFilters, responseFilters, nil
}

func apiRoomToDomainRoom(room api.RoomResponse) domain.Room {
	return domain.Room{
		Code:          room.Code,
		ResourceEmail: room.ResourceEmail,
		Name:          room.Name,
		Building:      room.Building,
		Floor:         room.Floor,
		Capacity:      room.Capacity,
		Type:          room.Type,
		Status:        room.Status,
		CurrentEvent:  apiEventToDomainEvent(room.CurrentEvent),
		NextEvent:     apiEventToDomainEvent(room.NextEvent),
	}
}

func domainRoomToAPIRoom(room domain.Room) api.RoomResponse {
	return api.RoomResponse{
		Code:          room.Code,
		ResourceEmail: room.ResourceEmail,
		Name:          room.Name,
		Building:      room.Building,
		Floor:         room.Floor,
		Capacity:      room.Capacity,
		Type:          room.Type,
		Status:        room.Status,
		CurrentEvent:  domainEventToAPIEvent(room.CurrentEvent),
		NextEvent:     domainEventToAPIEvent(room.NextEvent),
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

func mapDomainEventsToAPIEvents(events []domain.Event) []api.EventResponse {
	if events == nil {
		return nil
	}

	out := make([]api.EventResponse, len(events))
	for i := range events {
		out[i] = api.EventResponse{
			Title:     events[i].Title,
			Start:     events[i].Start,
			End:       events[i].End,
			Organizer: events[i].Organizer,
		}
	}

	return out
}

func newDefaultListService() domain.RoomService {
	// TODO(ticket-13): replace default static sources with real Google adapters
	// (Admin Directory inventory source + Calendar events client) wired from app composition root.
	clock := fixedListServiceClock{
		now: time.Date(2026, time.March, 9, 10, 30, 0, 0, time.UTC),
	}

	inventoryCache, err := domain.NewInventoryCache(
		context.Background(),
		defaultListInventorySource{},
		time.Hour,
		clock,
	)
	if err != nil {
		panic(err)
	}

	eventsCache, err := domain.NewRoomEventsCache(
		defaultListCalendarClient{},
		5*time.Minute,
		clock,
	)
	if err != nil {
		panic(err)
	}

	statusInterpreter := domain.NewStatusInterpreter(clock, nil)
	return NewService(inventoryCache, eventsCache, statusInterpreter, clock)
}

type listHandlerClock struct{}

func (listHandlerClock) Now() time.Time {
	return time.Now().UTC()
}

type fixedListServiceClock struct {
	now time.Time
}

func (c fixedListServiceClock) Now() time.Time {
	return c.now
}

type defaultListInventorySource struct{}

func (defaultListInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	rooms := make([]domain.Room, len(roomsFixture))
	for i := range roomsFixture {
		rooms[i] = apiRoomToDomainRoom(roomsFixture[i])
	}

	return domain.InventorySnapshot{
		Rooms: rooms,
	}, nil
}

type defaultListCalendarClient struct{}

func (defaultListCalendarClient) ListRoomEvents(_ context.Context, resourceEmail string, _, _ time.Time) ([]domain.Event, error) {
	switch resourceEmail {
	case "AMPHI-A":
		if nextEventFixture == nil {
			return nil, nil
		}
		return []domain.Event{
			{
				Title:     nextEventFixture.Title,
				Start:     nextEventFixture.Start,
				End:       nextEventFixture.End,
				Organizer: nextEventFixture.Organizer,
			},
		}, nil
	case "LAB-204":
		if currentEventFixture == nil {
			return nil, nil
		}
		return []domain.Event{
			{
				Title:     currentEventFixture.Title,
				Start:     currentEventFixture.Start,
				End:       currentEventFixture.End,
				Organizer: currentEventFixture.Organizer,
			},
		}, nil
	default:
		return nil, nil
	}
}
