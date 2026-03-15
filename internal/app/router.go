package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"

	"campus-room-status/internal/buildings"
	"campus-room-status/internal/docs"
	"campus-room-status/internal/domain"
	"campus-room-status/internal/google/adminsdk"
	gcalendar "campus-room-status/internal/google/calendar"
	goauth "campus-room-status/internal/google/oauth"
	"campus-room-status/internal/health"
	"campus-room-status/internal/rooms"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type runtimeDataSource string

const (
	runtimeDataSourceStatic runtimeDataSource = "static"
	runtimeDataSourceGoogle runtimeDataSource = "google"
)

type runtimeServices struct {
	buildingService domain.BuildingService
	roomService     domain.RoomService
	healthService   domain.HealthService
	inventoryCache  *domain.InventoryCache
	eventsCache     *domain.RoomEventsCache
}

// NewRouter godoc
// @Summary Get Swagger specification
// @Tags docs
// @Produce json
// @Success 200 {string} string "Swagger JSON document"
// @Router /api/v1/docs/openapi.json [get]
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

	services := newRuntimeServices()
	oauthFlow := newRuntimeOAuthFlow()

	apiGroup := r.Group("/api/v1")
	apiGroup.GET("/docs/openapi.json", docs.NewOpenAPIHandler())
	apiGroup.GET("/docs/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/api/v1/docs/openapi.json"),
	))
	// TODO(prod): protect or disable OAuth consent endpoints outside trusted admin network.
	apiGroup.GET("/auth/google/start", goauth.NewStartHandler(oauthFlow))
	apiGroup.GET("/auth/google/callback", goauth.NewCallbackHandlerWithHook(oauthFlow, services.refreshCachesAfterOAuth))
	apiGroup.GET("/buildings", buildings.NewHandler(services.buildingService, nil))
	apiGroup.GET("/health", health.NewHandler(services.healthService))
	apiGroup.GET("/rooms", rooms.NewListHandler(services.roomService, nil))
	apiGroup.GET("/rooms/:code", rooms.NewDetailHandler(services.roomService))
	apiGroup.GET("/rooms/:code/schedule", rooms.NewScheduleHandler(services.roomService))

	return r
}

func newRuntimeServices() runtimeServices {
	cache, err := domain.NewInventoryCache(
		context.Background(),
		newRuntimeInventorySource(),
		time.Hour,
		nil,
	)
	if err != nil {
		panic(err)
	}

	eventsCache, err := domain.NewRoomEventsCache(
		newRuntimeCalendarClient(),
		5*time.Minute,
		nil,
	)
	if err != nil {
		panic(err)
	}

	buildingService := buildings.NewService(cache)
	roomService := rooms.NewService(cache, eventsCache, nil, nil)
	// TODO(prod): inject version from build metadata/env instead of hardcoded "dev".
	healthService := health.NewService(cache, eventsCache, nil, "dev")

	return runtimeServices{
		buildingService: buildingService,
		roomService:     roomService,
		healthService:   healthService,
		inventoryCache:  cache,
		eventsCache:     eventsCache,
	}
}

func (s runtimeServices) refreshCachesAfterOAuth(ctx context.Context) {
	if s.inventoryCache != nil {
		_ = s.inventoryCache.ForceRefresh(ctx)
	}
	if s.eventsCache != nil {
		s.eventsCache.Clear()
	}
}

func newRuntimeInventorySource() domain.InventorySource {
	if runtimeDataSourceFromEnv() != runtimeDataSourceGoogle {
		return staticInventorySource{}
	}

	tokenProvider, ok := newRuntimeAdminTokenProvider()
	if !ok {
		return unavailableInventorySource{
			err: errors.New("google inventory source selected but no Google token provider is configured"),
		}
	}

	source, err := adminsdk.NewInventorySource(
		nil,
		tokenProvider,
		adminsdk.InventorySourceConfig{
			BaseURL:  strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_BASE_URL")),
			Customer: strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_CUSTOMER")),
			PageSize: envInt("GOOGLE_ADMIN_PAGE_SIZE"),
			Timeout:  envDuration("GOOGLE_ADMIN_TIMEOUT"),
		},
	)
	if err != nil {
		return unavailableInventorySource{
			err: fmt.Errorf("create Google inventory source: %w", err),
		}
	}

	return oauthBootstrapInventorySource{
		primary: source,
	}
}

func newRuntimeCalendarClient() domain.CalendarClient {
	if runtimeDataSourceFromEnv() != runtimeDataSourceGoogle {
		return staticCalendarClient{}
	}

	tokenProvider, ok := newRuntimeAdminTokenProvider()
	if !ok {
		return unavailableCalendarClient{
			err: errors.New("google calendar client selected but no Google token provider is configured"),
		}
	}

	client, err := gcalendar.NewClient(
		nil,
		tokenProvider,
		gcalendar.ClientConfig{
			BaseURL:  strings.TrimSpace(os.Getenv("GOOGLE_CALENDAR_BASE_URL")),
			Timeout:  envDuration("GOOGLE_CALENDAR_TIMEOUT"),
			PageSize: envInt("GOOGLE_CALENDAR_PAGE_SIZE"),
		},
	)
	if err != nil {
		return unavailableCalendarClient{
			err: fmt.Errorf("create Google calendar client: %w", err),
		}
	}

	return client
}

func newRuntimeAdminTokenProvider() (adminsdk.TokenProvider, bool) {
	if provider, ok := newRuntimeOAuthTokenProvider(); ok {
		return provider, true
	}

	credentialsJSON, hasCredentials := readServiceAccountCredentials()
	if hasCredentials {
		provider, err := adminsdk.NewServiceAccountTokenProvider(adminsdk.ServiceAccountTokenProviderConfig{
			CredentialsJSON: credentialsJSON,
			Subject:         strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_IMPERSONATED_USER")),
		})
		if err == nil {
			return provider, true
		}
	}

	token := strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_BEARER_TOKEN"))
	if token != "" {
		return staticTokenProvider{token: token}, true
	}

	return nil, false
}

func newRuntimeOAuthTokenProvider() (adminsdk.TokenProvider, bool) {
	cfg, err := goauth.LoadConfigFromEnv()
	if err != nil {
		return nil, false
	}

	// TODO(prod): replace file token store with secret manager or env-backed secure store.
	store := goauth.NewFileRefreshTokenStore(cfg.RefreshTokenFile)

	provider, err := goauth.NewTokenProvider(cfg, store)
	if err != nil {
		return nil, false
	}

	return provider, true
}

func newRuntimeOAuthFlow() *goauth.AuthorizationFlow {
	cfg, err := goauth.LoadConfigFromEnv()
	if err != nil {
		return nil
	}

	// TODO(prod): replace file token store with secret manager or env-backed secure store.
	store := goauth.NewFileRefreshTokenStore(cfg.RefreshTokenFile)
	flow, err := goauth.NewAuthorizationFlow(cfg, store)
	if err != nil {
		return nil
	}

	return flow
}

func readServiceAccountCredentials() ([]byte, bool) {
	rawJSON := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON"))
	if rawJSON != "" {
		return []byte(rawJSON), true
	}

	rawBase64 := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON_BASE64"))
	if rawBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(rawBase64)
		if err == nil && len(decoded) > 0 {
			return decoded, true
		}
	}

	filePath := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE"))
	if filePath == "" {
		return nil, false
	}

	credentialsJSON, err := os.ReadFile(filePath)
	if err != nil || len(credentialsJSON) == 0 {
		return nil, false
	}

	return credentialsJSON, true
}

func runtimeDataSourceFromEnv() runtimeDataSource {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("DATA_SOURCE")))
	if raw == string(runtimeDataSourceGoogle) {
		return runtimeDataSourceGoogle
	}

	return runtimeDataSourceStatic
}

func envInt(name string) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}

	return value
}

func envDuration(name string) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0
	}

	return value
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

type staticTokenProvider struct {
	token string
}

func (p staticTokenProvider) Token(context.Context) (string, error) {
	return p.token, nil
}

type oauthBootstrapInventorySource struct {
	primary domain.InventorySource
}

func (s oauthBootstrapInventorySource) LoadInventory(ctx context.Context) (domain.InventorySnapshot, error) {
	snapshot, err := s.primary.LoadInventory(ctx)
	if err == nil {
		return snapshot, nil
	}
	if !isMissingRefreshTokenError(err) {
		return domain.InventorySnapshot{}, err
	}

	// Keep API bootstrappable for OAuth consent flow without serving static fixtures.
	return domain.InventorySnapshot{}, nil
}

func isMissingRefreshTokenError(err error) bool {
	if err == nil {
		return false
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "missing refresh token")
}

type unavailableInventorySource struct {
	err error
}

func (s unavailableInventorySource) LoadInventory(context.Context) (domain.InventorySnapshot, error) {
	return domain.InventorySnapshot{}, s.err
}

type unavailableCalendarClient struct {
	err error
}

func (c unavailableCalendarClient) ListRoomEvents(context.Context, string, time.Time, time.Time) ([]domain.Event, error) {
	return nil, c.err
}
