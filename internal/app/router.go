package app

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"

	"campus-room-status/internal/buildings"
	"campus-room-status/internal/domain"
	"campus-room-status/internal/google/adminsdk"
	gcalendar "campus-room-status/internal/google/calendar"
	goauth "campus-room-status/internal/google/oauth"
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
	oauthFlow := newRuntimeOAuthFlow()

	apiGroup := r.Group("/api/v1")
	apiGroup.GET("/auth/google/start", goauth.NewStartHandler(oauthFlow))
	apiGroup.GET("/auth/google/callback", goauth.NewCallbackHandler(oauthFlow))
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
	healthService := health.NewService(cache, eventsCache, nil, "dev")

	return buildingService, roomService, healthService
}

func newRuntimeInventorySource() domain.InventorySource {
	tokenProvider, ok := newRuntimeAdminTokenProvider()
	if !ok {
		return staticInventorySource{}
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
		return staticInventorySource{}
	}

	return source
}

func newRuntimeCalendarClient() domain.CalendarClient {
	tokenProvider, ok := newRuntimeAdminTokenProvider()
	if !ok {
		return staticCalendarClient{}
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
		return staticCalendarClient{}
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

	store := goauth.NewFileRefreshTokenStore(cfg.RefreshTokenFile)
	refreshToken, err := store.Load(context.Background())
	if err != nil {
		return nil, false
	}
	if strings.TrimSpace(refreshToken) == "" {
		return nil, false
	}

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
