package adminsdk

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestInventorySource_LoadInventory_MapsGoogleResponseToBuildingsAndRooms(t *testing.T) {
	t.Parallel()

	requestedPaths := make([]string, 0, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected Authorization header to contain bearer token, got %q", got)
		}

		switch r.URL.Path {
		case "/admin/directory/v1/customer/test-customer/resources/buildings":
			_, _ = w.Write([]byte(`{
				"buildings": [{
					"buildingId": "B1",
					"buildingName": "Main Campus",
					"address": {
						"addressLines": ["1 Campus Street"],
						"locality": "Paris",
						"postalCode": "75000",
						"regionCode": "FR"
					},
					"floorNames": ["0", "1", "RDC"]
				}]
			}`))
		case "/admin/directory/v1/customer/test-customer/resources/calendars":
			_, _ = w.Write([]byte(`{
				"items": [
					{
						"resourceName": "AMPHI-A",
						"resourceEmail": "amphi-a@example.org",
						"capacity": 180,
						"resourceType": "amphitheater",
						"buildingId": "B1",
						"floorName": "1"
					},
					{
						"resourceName": "LAB-204",
						"resourceEmail": "lab-204@example.org",
						"capacity": 30,
						"resourceCategory": "lab",
						"buildingId": "B2",
						"floorName": "2"
					}
				]
			}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	source := newTestInventorySource(t, &http.Client{}, InventorySourceConfig{
		BaseURL:  server.URL,
		Customer: "test-customer",
	})

	snapshot, err := source.LoadInventory(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !slices.Contains(requestedPaths, "/admin/directory/v1/customer/test-customer/resources/buildings") {
		t.Fatalf("expected buildings endpoint to be called, got paths: %v", requestedPaths)
	}
	if !slices.Contains(requestedPaths, "/admin/directory/v1/customer/test-customer/resources/calendars") {
		t.Fatalf("expected calendars endpoint to be called, got paths: %v", requestedPaths)
	}

	if len(snapshot.Buildings) != 1 {
		t.Fatalf("expected 1 building, got %d", len(snapshot.Buildings))
	}
	building := snapshot.Buildings[0]
	if building.ID != "B1" {
		t.Fatalf("expected building id B1, got %q", building.ID)
	}
	if building.Name != "Main Campus" {
		t.Fatalf("expected building name Main Campus, got %q", building.Name)
	}
	if building.Address == "" {
		t.Fatalf("expected building address to be mapped")
	}
	if !reflect.DeepEqual(building.Floors, []string{"0", "1", "RDC"}) {
		t.Fatalf("expected parsed floors [0 1 RDC], got %v", building.Floors)
	}

	if len(snapshot.Rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(snapshot.Rooms))
	}

	first := snapshot.Rooms[0]
	if first.Code != "AMPHI-A" {
		t.Fatalf("expected first room code AMPHI-A, got %q", first.Code)
	}
	if first.ResourceEmail != "amphi-a@example.org" {
		t.Fatalf("expected first room resourceEmail amphi-a@example.org, got %q", first.ResourceEmail)
	}
	if first.Name != "AMPHI-A" {
		t.Fatalf("expected first room name AMPHI-A, got %q", first.Name)
	}
	if first.Building != "B1" {
		t.Fatalf("expected first room building B1, got %q", first.Building)
	}
	if first.Floor != 1 {
		t.Fatalf("expected first room floor 1, got %d", first.Floor)
	}
	if first.Capacity != 180 {
		t.Fatalf("expected first room capacity 180, got %d", first.Capacity)
	}
	if first.Type != "amphitheater" {
		t.Fatalf("expected first room type amphitheater, got %q", first.Type)
	}

	second := snapshot.Rooms[1]
	if second.Code != "LAB-204" {
		t.Fatalf("expected second room code LAB-204, got %q", second.Code)
	}
	if second.ResourceEmail != "lab-204@example.org" {
		t.Fatalf("expected second room resourceEmail lab-204@example.org, got %q", second.ResourceEmail)
	}
	if second.Type != "lab" {
		t.Fatalf("expected second room type from resourceCategory fallback, got %q", second.Type)
	}
}

func TestInventorySource_LoadInventory_TracksAdditionalResourceFieldsWhenPresent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/directory/v1/customer/test-customer/resources/buildings":
			_, _ = w.Write([]byte(`{"buildings":[]}`))
		case "/admin/directory/v1/customer/test-customer/resources/calendars":
			_, _ = w.Write([]byte(`{
				"items": [{
					"generatedResourceName": "LAB-ADV-01",
					"resourceName": "",
					"resourceEmail": "lab-adv-01@example.org",
					"capacity": 24,
					"resourceCategory": "lab",
					"buildingId": "B9",
					"floorName": "3",
					"userVisibleDescription": "Room closed every Friday evening",
					"featureInstances": [{"feature": {"name": "PROJECTOR"}}]
				}]
			}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	source := newTestInventorySource(t, &http.Client{}, InventorySourceConfig{
		BaseURL:  server.URL,
		Customer: "test-customer",
	})

	snapshot, err := source.LoadInventory(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snapshot.Rooms) != 1 {
		t.Fatalf("expected one room, got %d", len(snapshot.Rooms))
	}

	room := snapshot.Rooms[0]
	if room.Code != "LAB-ADV-01" {
		t.Fatalf("expected code from generatedResourceName, got %q", room.Code)
	}
	if room.ResourceEmail != "lab-adv-01@example.org" {
		t.Fatalf("expected resourceEmail from payload, got %q", room.ResourceEmail)
	}
	if room.Building != "B9" {
		t.Fatalf("expected buildingId to be preserved, got %q", room.Building)
	}
	if room.Floor != 3 {
		t.Fatalf("expected floor parsed from floorName, got %d", room.Floor)
	}

	fields := source.ObservedResourceFields()
	expectedFields := []string{
		"generatedResourceName",
		"featureInstances",
		"userVisibleDescription",
		"resourceEmail",
	}
	for _, field := range expectedFields {
		if !slices.Contains(fields, field) {
			t.Fatalf("expected observed resource fields to contain %q, got %v", field, fields)
		}
	}
}

func TestInventorySource_LoadInventory_UsesEmailFallbackForCodeWhenNamesMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/directory/v1/customer/test-customer/resources/buildings":
			_, _ = w.Write([]byte(`{"buildings":[]}`))
		case "/admin/directory/v1/customer/test-customer/resources/calendars":
			_, _ = w.Write([]byte(`{
				"items": [{
					"resourceName": "",
					"resourceEmail": "room-alpha@example.org",
					"capacity": 12,
					"resourceType": "meeting_room"
				}]
			}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	source := newTestInventorySource(t, &http.Client{}, InventorySourceConfig{
		BaseURL:  server.URL,
		Customer: "test-customer",
	})

	snapshot, err := source.LoadInventory(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snapshot.Rooms) != 1 {
		t.Fatalf("expected one room, got %d", len(snapshot.Rooms))
	}
	if snapshot.Rooms[0].Code != "room-alpha" {
		t.Fatalf("expected room code fallback from resourceEmail local part, got %q", snapshot.Rooms[0].Code)
	}
	if snapshot.Rooms[0].ResourceEmail != "room-alpha@example.org" {
		t.Fatalf("expected room resourceEmail to be preserved, got %q", snapshot.Rooms[0].ResourceEmail)
	}
}

func TestInventorySource_LoadInventory_HandlesEmptyResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/directory/v1/customer/test-customer/resources/buildings":
			_, _ = w.Write([]byte(`{}`))
		case "/admin/directory/v1/customer/test-customer/resources/calendars":
			_, _ = w.Write([]byte(`{"items":[]}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	source := newTestInventorySource(t, &http.Client{}, InventorySourceConfig{
		BaseURL:  server.URL,
		Customer: "test-customer",
	})

	snapshot, err := source.LoadInventory(context.Background())
	if err != nil {
		t.Fatalf("expected no error on empty payload, got %v", err)
	}
	if len(snapshot.Buildings) != 0 {
		t.Fatalf("expected no buildings, got %d", len(snapshot.Buildings))
	}
	if len(snapshot.Rooms) != 0 {
		t.Fatalf("expected no rooms, got %d", len(snapshot.Rooms))
	}
}

func TestInventorySource_LoadInventory_HandlesTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 25 * time.Millisecond,
	}
	source := newTestInventorySource(t, client, InventorySourceConfig{
		BaseURL:  server.URL,
		Customer: "test-customer",
	})

	_, err := source.LoadInventory(context.Background())
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "timeout") &&
		!strings.Contains(strings.ToLower(err.Error()), "deadline") {
		t.Fatalf("expected timeout/deadline error, got %v", err)
	}
}

func TestInventorySource_LoadInventory_ReturnsErrorOnQuotaResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/directory/v1/customer/test-customer/resources/buildings":
			_, _ = w.Write([]byte(`{"buildings":[]}`))
		case "/admin/directory/v1/customer/test-customer/resources/calendars":
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":429,"message":"quota exceeded"}}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	source := newTestInventorySource(t, &http.Client{}, InventorySourceConfig{
		BaseURL:  server.URL,
		Customer: "test-customer",
	})

	_, err := source.LoadInventory(context.Background())
	if err == nil {
		t.Fatalf("expected quota error, got nil")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected error to mention HTTP 429, got %v", err)
	}
}

func TestInventorySource_LoadInventory_ReturnsErrorOnNetworkFailure(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial tcp 127.0.0.1:443: connectex: connection refused")
		}),
	}
	source := newTestInventorySource(t, client, InventorySourceConfig{
		BaseURL:  "http://example.invalid",
		Customer: "test-customer",
	})

	_, err := source.LoadInventory(context.Background())
	if err == nil {
		t.Fatalf("expected network error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "connection refused") {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func newTestInventorySource(t *testing.T, client *http.Client, cfg InventorySourceConfig) *InventorySource {
	t.Helper()

	source, err := NewInventorySource(client, staticTokenProvider{token: "test-token"}, cfg)
	if err != nil {
		t.Fatalf("expected source creation to succeed, got %v", err)
	}

	return source
}

type staticTokenProvider struct {
	token string
	err   error
}

func (p staticTokenProvider) Token(context.Context) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	return p.token, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
