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

	buildingService := newBuildingService()

	apiGroup := r.Group("/api/v1")
	apiGroup.GET("/buildings", buildings.NewHandler(buildingService, nil))
	apiGroup.GET("/health", health.Handler)
	apiGroup.GET("/rooms", rooms.ListHandler)
	apiGroup.GET("/rooms/:code", rooms.DetailHandler)
	apiGroup.GET("/rooms/:code/schedule", rooms.ScheduleHandler)

	return r
}

func newBuildingService() domain.BuildingService {
	cache, err := domain.NewInventoryCache(
		context.Background(),
		staticInventorySource{},
		time.Hour,
		nil,
	)
	if err != nil {
		panic(err)
	}

	return buildings.NewService(cache)
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
		},
	}, nil
}
