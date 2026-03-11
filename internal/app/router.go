package app

import (
	"context"
	"net/http"
	"time"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"

	"campus-room-status/internal/buildings"
	"campus-room-status/internal/domain"
	"campus-room-status/internal/health"
	"campus-room-status/internal/rooms"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Logger(),
		gin.CustomRecovery(func(c *gin.Context, _ any) {
			api.WriteError(c, api.NewHTTPError(
				http.StatusInternalServerError,
				api.ErrorCodeInternalServerError,
				"Une erreur interne est survenue",
			))
		}),
	)

	r.NoRoute(func(c *gin.Context) {
		api.WriteError(c, api.NewHTTPError(
			http.StatusNotFound,
			api.ErrorCodeNotFound,
			"La route demandee n'existe pas",
		))
	})
	r.NoMethod(func(c *gin.Context) {
		api.WriteError(c, api.NewHTTPError(
			http.StatusNotFound,
			api.ErrorCodeNotFound,
			"La route demandee n'existe pas",
		))
	})

	buildingService, roomService, healthService := newRuntimeServices()

	apiGroup := r.Group("/api/v1")
	apiGroup.GET("/buildings", buildings.NewHandler(buildingService, nil))
	apiGroup.GET("/health", health.NewHandler(healthService))
	apiGroup.GET("/rooms", rooms.NewListHandler(roomService, nil))
	apiGroup.GET("/rooms/:code", rooms.NewDetailHandler(roomService))
	apiGroup.GET("/rooms/:code/schedule", rooms.NewScheduleHandler(roomService))

	return r
}

func newRuntimeServices() (domain.BuildingService, domain.RoomService, domain.HealthService) {
	cache, err := domain.NewInventoryCache(
		context.Background(),
		staticInventorySource{},
		time.Hour,
		nil,
	)
	if err != nil {
		panic(err)
	}

	eventsCache, err := domain.NewRoomEventsCache(
		staticCalendarClient{},
		5*time.Minute,
		nil,
	)
	if err != nil {
		panic(err)
	}

	buildingService := buildings.NewService(cache)
	roomService := rooms.NewService(cache, eventsCache, nil, nil)
	healthService := health.NewService(cache, eventsCache, nil, "dev")

	return buildingService, roomService, healthService
}

type staticInventorySource struct{}

func (staticInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	return domain.InventorySnapshot{
		Buildings: []domain.Building{
			{
				ID:      "B1",
				Name:    "Building A",
				Address: "1 Campus Street",
				Floors:  []int{0, 1, 2},
			},
			{
				ID:      "B2",
				Name:    "Building B",
				Address: "2 Campus Street",
				Floors:  []int{0, 1, 2, 3},
			},
		},
		Rooms: []domain.Room{
			{
				Code:     "AMPHI-A",
				Name:     "Amphitheater A",
				Building: "B1",
				Floor:    1,
				Capacity: 180,
				Type:     "amphitheater",
				Status:   "available",
			},
			{
				Code:     "LAB-204",
				Name:     "Computer Lab 204",
				Building: "B2",
				Floor:    2,
				Capacity: 30,
				Type:     "lab",
				Status:   "available",
			},
		},
	}, nil
}

type staticCalendarClient struct{}

func (staticCalendarClient) ListRoomEvents(_ context.Context, resourceEmail string, _, _ time.Time) ([]domain.Event, error) {
	switch resourceEmail {
	case "AMPHI-A":
		return []domain.Event{
			{
				Title:     "Capstone Review",
				Start:     time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC),
				End:       time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
				Organizer: "Academic Board",
			},
		}, nil
	case "LAB-204":
		return []domain.Event{
			{
				Title:     "OS Lab Session",
				Start:     time.Date(2026, time.March, 9, 10, 0, 0, 0, time.UTC),
				End:       time.Date(2026, time.March, 9, 12, 0, 0, 0, time.UTC),
				Organizer: "Systems Team",
			},
		}, nil
	default:
		return nil, nil
	}
}
