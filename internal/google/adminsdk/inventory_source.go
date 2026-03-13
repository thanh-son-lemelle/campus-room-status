package adminsdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"campus-room-status/internal/domain"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const (
	defaultEndpoint = "https://admin.googleapis.com/"
	defaultCustomer = "my_customer"
	defaultPageSize = 100
	defaultTimeout  = 10 * time.Second
)

type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

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

	authorizedClient := newAuthorizedHTTPClient(baseClient, tokenProvider)
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

func (s *InventorySource) listBuildings(ctx context.Context) ([]domain.Building, error) {
	pageToken := ""
	visitedTokens := make(map[string]struct{})
	buildingsByID := make(map[string]domain.Building)

	for {
		call := s.service.Resources.Buildings.List(s.customer).MaxResults(s.pageSize)
		if pageToken != "" {
			if _, exists := visitedTokens[pageToken]; exists {
				return nil, fmt.Errorf("detected repeated page token %q on buildings endpoint", pageToken)
			}
			visitedTokens[pageToken] = struct{}{}
			call = call.PageToken(pageToken)
		}

		response, err := call.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("buildings list request failed: %w", err)
		}
		if response == nil {
			break
		}

		for _, building := range response.Buildings {
			mapped, ok := mapBuilding(building)
			if !ok {
				continue
			}
			buildingsByID[mapped.ID] = mapped
		}

		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	buildings := make([]domain.Building, 0, len(buildingsByID))
	for _, building := range buildingsByID {
		buildings = append(buildings, building)
	}
	sort.SliceStable(buildings, func(i, j int) bool {
		return buildings[i].ID < buildings[j].ID
	})

	return buildings, nil
}

func (s *InventorySource) listCalendarResources(ctx context.Context) ([]domain.Room, error) {
	pageToken := ""
	visitedTokens := make(map[string]struct{})
	roomsByCode := make(map[string]domain.Room)

	for {
		call := s.service.Resources.Calendars.List(s.customer).MaxResults(s.pageSize)
		if pageToken != "" {
			if _, exists := visitedTokens[pageToken]; exists {
				return nil, fmt.Errorf("detected repeated page token %q on calendars endpoint", pageToken)
			}
			visitedTokens[pageToken] = struct{}{}
			call = call.PageToken(pageToken)
		}

		response, err := call.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("calendars list request failed: %w", err)
		}
		if response == nil {
			break
		}

		for _, resource := range response.Items {
			mapped, fields, ok := mapCalendarResource(resource)
			if !ok {
				continue
			}
			s.recordObservedFields(fields)
			roomsByCode[mapped.Code] = mapped
		}

		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	rooms := make([]domain.Room, 0, len(roomsByCode))
	for _, room := range roomsByCode {
		rooms = append(rooms, room)
	}
	sort.SliceStable(rooms, func(i, j int) bool {
		return rooms[i].Code < rooms[j].Code
	})

	return rooms, nil
}

func (s *InventorySource) recordObservedFields(fields []string) {
	s.observedFieldsMu.Lock()
	defer s.observedFieldsMu.Unlock()

	for _, field := range fields {
		s.observedFields[field] = struct{}{}
	}
}

func mapBuilding(building *admin.Building) (domain.Building, bool) {
	if building == nil {
		return domain.Building{}, false
	}

	id := strings.TrimSpace(building.BuildingId)
	if id == "" {
		return domain.Building{}, false
	}

	name := firstNonEmpty(building.BuildingName, id)
	floors := parseNumericFloors(building.FloorNames)
	address := mapBuildingAddress(building.Address)

	return domain.Building{
		ID:      id,
		Name:    name,
		Address: address,
		Floors:  floors,
	}, true
}

func mapBuildingAddress(address *admin.BuildingAddress) string {
	if address == nil {
		return ""
	}

	parts := make([]string, 0, 8)
	parts = append(parts, trimNonEmpty(address.AddressLines)...)
	parts = appendIfNotEmpty(parts, address.Locality)
	parts = appendIfNotEmpty(parts, address.AdministrativeArea)
	parts = appendIfNotEmpty(parts, address.PostalCode)
	parts = appendIfNotEmpty(parts, address.RegionCode)

	return strings.Join(parts, ", ")
}

func mapCalendarResource(resource *admin.CalendarResource) (domain.Room, []string, bool) {
	if resource == nil {
		return domain.Room{}, nil, false
	}

	resourceName := strings.TrimSpace(resource.ResourceName)
	resourceEmail := strings.TrimSpace(resource.ResourceEmail)
	generatedResourceName := strings.TrimSpace(resource.GeneratedResourceName)

	code := firstNonEmpty(
		generatedResourceName,
		resourceName,
		codeFromEmail(resourceEmail),
	)
	if code == "" {
		return domain.Room{}, nil, false
	}

	floor, _ := strconv.Atoi(strings.TrimSpace(resource.FloorName))
	resourceType := firstNonEmpty(resource.ResourceType, resource.ResourceCategory, "room")

	return domain.Room{
		Code:          code,
		ResourceEmail: resourceEmail,
		Name:          firstNonEmpty(resourceName, generatedResourceName, code),
		Building:      strings.TrimSpace(resource.BuildingId),
		Floor:         floor,
		Capacity:      int(resource.Capacity),
		Type:          resourceType,
		Status:        "available",
	}, observedFieldsFromResource(resource), true
}

func observedFieldsFromResource(resource *admin.CalendarResource) []string {
	fields := make([]string, 0, 16)

	addField := func(name string, shouldAdd bool) {
		if shouldAdd {
			fields = append(fields, name)
		}
	}

	addField("generatedResourceName", strings.TrimSpace(resource.GeneratedResourceName) != "")
	addField("resourceName", strings.TrimSpace(resource.ResourceName) != "")
	addField("resourceEmail", strings.TrimSpace(resource.ResourceEmail) != "")
	addField("capacity", resource.Capacity > 0)
	addField("resourceType", strings.TrimSpace(resource.ResourceType) != "")
	addField("resourceCategory", strings.TrimSpace(resource.ResourceCategory) != "")
	addField("buildingId", strings.TrimSpace(resource.BuildingId) != "")
	addField("floorName", strings.TrimSpace(resource.FloorName) != "")
	addField("floorSection", strings.TrimSpace(resource.FloorSection) != "")
	addField("resourceDescription", strings.TrimSpace(resource.ResourceDescription) != "")
	addField("userVisibleDescription", strings.TrimSpace(resource.UserVisibleDescription) != "")
	addField("featureInstances", hasAnyValue(resource.FeatureInstances))

	sort.Strings(fields)
	return fields
}

func deriveBuildingsFromRooms(rooms []domain.Room) []domain.Building {
	byID := make(map[string]*domain.Building)

	for _, room := range rooms {
		if strings.TrimSpace(room.Building) == "" {
			continue
		}

		building, exists := byID[room.Building]
		if !exists {
			building = &domain.Building{
				ID:   room.Building,
				Name: room.Building,
			}
			byID[room.Building] = building
		}

		if room.Floor != 0 && !slicesContainsInt(building.Floors, room.Floor) {
			building.Floors = append(building.Floors, room.Floor)
		}
	}

	buildings := make([]domain.Building, 0, len(byID))
	for _, building := range byID {
		sort.Ints(building.Floors)
		buildings = append(buildings, *building)
	}
	sort.SliceStable(buildings, func(i, j int) bool {
		return buildings[i].ID < buildings[j].ID
	})

	return buildings
}

func parseNumericFloors(floorNames []string) []int {
	if len(floorNames) == 0 {
		return nil
	}

	floors := make([]int, 0, len(floorNames))
	for _, floorName := range floorNames {
		value, err := strconv.Atoi(strings.TrimSpace(floorName))
		if err != nil {
			continue
		}
		if !slicesContainsInt(floors, value) {
			floors = append(floors, value)
		}
	}
	sort.Ints(floors)
	return floors
}

func codeFromEmail(resourceEmail string) string {
	at := strings.Index(resourceEmail, "@")
	if at <= 0 {
		return strings.TrimSpace(resourceEmail)
	}
	return strings.TrimSpace(resourceEmail[:at])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func appendIfNotEmpty(values []string, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return values
	}
	return append(values, trimmed)
}

func trimNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func hasAnyValue(value any) bool {
	if value == nil {
		return false
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return reflected.Len() > 0
	case reflect.Bool:
		return reflected.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflected.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return reflected.Float() != 0
	case reflect.Interface, reflect.Pointer:
		return !reflected.IsNil()
	default:
		return true
	}
}

func slicesContainsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func normalizeEndpoint(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return defaultEndpoint
	}

	trimmed = strings.TrimRight(trimmed, "/")
	if strings.HasSuffix(trimmed, "/admin/directory/v1") {
		trimmed = strings.TrimSuffix(trimmed, "/admin/directory/v1")
	}

	return trimmed + "/"
}

func newAuthorizedHTTPClient(client *http.Client, tokenProvider TokenProvider) *http.Client {
	clone := *client
	transport := clone.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	clone.Transport = authorizedTransport{
		base:          transport,
		tokenProvider: tokenProvider,
	}
	return &clone
}

type authorizedTransport struct {
	base          http.RoundTripper
	tokenProvider TokenProvider
}

func (t authorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())

	token, err := t.tokenProvider.Token(req.Context())
	if err != nil {
		return nil, fmt.Errorf("retrieve access token: %w", err)
	}

	if strings.TrimSpace(token) != "" {
		cloned.Header.Set("Authorization", "Bearer "+token)
	}

	return t.base.RoundTrip(cloned)
}
