package adminsdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"campus-room-status/internal/domain"
	"campus-room-status/internal/google/httpauth"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const (
	defaultEndpoint = "https://admin.googleapis.com/"
	defaultCustomer = "my_customer"
	defaultPageSize = 100
	defaultTimeout  = 10 * time.Second
)

type TokenProvider = httpauth.TokenProvider

type InventorySourceConfig struct {
	BaseURL  string
	Customer string
	PageSize int
	Timeout  time.Duration
}

type InventorySource struct {
	service  *admin.Service
	customer string
	pageSize int64

	observedFieldsMu sync.RWMutex
	observedFields   map[string]struct{}
}

var _ domain.InventorySource = (*InventorySource)(nil)

// NewInventorySource creates a new inventory source.
//
// Summary:
// - Creates a new inventory source.
//
// Attributes:
// - client (*http.Client): Input parameter.
// - tokenProvider (TokenProvider): Input parameter.
// - cfg (InventorySourceConfig): Input parameter.
//
// Returns:
// - value1 (*InventorySource): Returned value.
// - value2 (error): Returned value.
func NewInventorySource(client *http.Client, tokenProvider TokenProvider, cfg InventorySourceConfig) (*InventorySource, error) {
	if tokenProvider == nil {
		return nil, errors.New("token provider is required")
	}

	customer := strings.TrimSpace(cfg.Customer)
	if customer == "" {
		customer = defaultCustomer
	}

	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	baseClient := client
	if baseClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		baseClient = &http.Client{Timeout: timeout}
	}

	authorizedClient := httpauth.NewAuthorizedHTTPClient(baseClient, tokenProvider)
	service, err := admin.NewService(
		context.Background(),
		option.WithHTTPClient(authorizedClient),
		option.WithEndpoint(normalizeEndpoint(cfg.BaseURL)),
		option.WithUserAgent("campus-room-status/adminsdk"),
	)
	if err != nil {
		return nil, fmt.Errorf("create admin directory service: %w", err)
	}

	return &InventorySource{
		service:        service,
		customer:       customer,
		pageSize:       int64(pageSize),
		observedFields: make(map[string]struct{}),
	}, nil
}

// LoadInventory loads inventory.
//
// Summary:
// - Loads inventory.
//
// Attributes:
// - ctx (context.Context): Input parameter.
//
// Returns:
// - value1 (domain.InventorySnapshot): Returned value.
// - value2 (error): Returned value.
func (s *InventorySource) LoadInventory(ctx context.Context) (domain.InventorySnapshot, error) {
	buildings, err := s.listBuildings(ctx)
	if err != nil {
		return domain.InventorySnapshot{}, fmt.Errorf("list buildings from admin directory: %w", err)
	}

	rooms, err := s.listCalendarResources(ctx)
	if err != nil {
		return domain.InventorySnapshot{}, fmt.Errorf("list calendar resources from admin directory: %w", err)
	}

	if len(buildings) == 0 {
		buildings = deriveBuildingsFromRooms(rooms)
	}

	return domain.InventorySnapshot{
		Buildings: buildings,
		Rooms:     rooms,
	}, nil
}

// ObservedResourceFields observeds resource fields.
//
// Summary:
// - Observeds resource fields.
//
// Attributes:
// - None.
//
// Returns:
// - value1 ([]string): Returned value.
func (s *InventorySource) ObservedResourceFields() []string {
	s.observedFieldsMu.RLock()
	defer s.observedFieldsMu.RUnlock()

	fields := make([]string, 0, len(s.observedFields))
	for field := range s.observedFields {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}
