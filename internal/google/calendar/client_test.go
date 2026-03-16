package calendar

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestClient_ListRoomEvents_MapsFreeBusyIntervalsToBusyEvents(t *testing.T) {
	t.Parallel()

	requestedPaths := make([]string, 0, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected Authorization header to contain bearer token, got %q", got)
		}

		switch r.URL.Path {
		case "/calendar/v3/freeBusy":
			_, _ = w.Write([]byte(`{
				"calendars": {
					"room@example.org": {
						"busy": [{
							"start": "2026-03-10T09:00:00Z",
							"end": "2026-03-10T10:30:00Z"
						}]
					}
				}
			}`))
		case "/calendar/v3/calendars/room@example.org/events":
			_, _ = w.Write([]byte(`{"items":[]}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestCalendarClient(t, &http.Client{}, ClientConfig{
		BaseURL: server.URL,
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)

	events, err := client.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !slices.Contains(requestedPaths, "/calendar/v3/freeBusy") {
		t.Fatalf("expected freeBusy endpoint call, got %v", requestedPaths)
	}
	if !slices.Contains(requestedPaths, "/calendar/v3/calendars/room@example.org/events") {
		t.Fatalf("expected events endpoint call, got %v", requestedPaths)
	}

	if len(events) != 1 {
		t.Fatalf("expected one busy event, got %d", len(events))
	}
	if events[0].Title != "Unknown event" {
		t.Fatalf("expected unknown fallback title, got %q", events[0].Title)
	}
	if !events[0].Start.Equal(time.Date(2026, time.March, 10, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected busy event start: %s", events[0].Start)
	}
	if !events[0].End.Equal(time.Date(2026, time.March, 10, 10, 30, 0, 0, time.UTC)) {
		t.Fatalf("unexpected busy event end: %s", events[0].End)
	}
}

func TestClient_ListRoomEvents_MapsEventsListToDomainEvents(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/calendar/v3/freeBusy":
			_, _ = w.Write([]byte(`{
				"calendars": {
					"room@example.org": { "busy": [] }
				}
			}`))
		case "/calendar/v3/calendars/room@example.org/events":
			_, _ = w.Write([]byte(`{
				"items": [
					{
						"id": "evt-2",
						"summary": "Distributed Systems",
						"start": {"dateTime": "2026-03-10T13:00:00Z"},
						"end": {"dateTime": "2026-03-10T15:00:00Z"},
						"organizer": {"email": "teacher@example.org"}
					},
					{
						"id": "evt-1",
						"summary": "Algorithms",
						"start": {"dateTime": "2026-03-10T09:00:00Z"},
						"end": {"dateTime": "2026-03-10T10:00:00Z"},
						"organizer": {"displayName": "Academic Office"}
					}
				]
			}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestCalendarClient(t, &http.Client{}, ClientConfig{
		BaseURL: server.URL,
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC)

	events, err := client.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Title != "Algorithms" {
		t.Fatalf("expected first event Algorithms after sorting, got %q", events[0].Title)
	}
	if events[0].Organizer != "Academic Office" {
		t.Fatalf("expected displayName organizer fallback, got %q", events[0].Organizer)
	}
	if events[1].Title != "Distributed Systems" {
		t.Fatalf("expected second event Distributed Systems, got %q", events[1].Title)
	}
	if events[1].Organizer != "teacher@example.org" {
		t.Fatalf("expected organizer email fallback, got %q", events[1].Organizer)
	}
}

func TestClient_ListRoomEvents_UsesReadableFallbacksWhenSummaryAndOrganizerAreMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/calendar/v3/freeBusy":
			_, _ = w.Write([]byte(`{
				"calendars": {
					"room@example.org": { "busy": [] }
				}
			}`))
		case "/calendar/v3/calendars/room@example.org/events":
			_, _ = w.Write([]byte(`{
				"items": [{
					"id": "opaque-event-id",
					"start": {"dateTime": "2026-03-10T09:00:00Z"},
					"end": {"dateTime": "2026-03-10T10:00:00Z"}
				}]
			}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestCalendarClient(t, &http.Client{}, ClientConfig{
		BaseURL: server.URL,
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC)

	events, err := client.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Unknown event" {
		t.Fatalf("expected title fallback Unknown event, got %q", events[0].Title)
	}
	if events[0].Organizer != "Google Calendar" {
		t.Fatalf("expected organizer fallback Google Calendar, got %q", events[0].Organizer)
	}
}

func TestClient_ListRoomEvents_HandlesPartialResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/calendar/v3/freeBusy":
			_, _ = w.Write([]byte(`{
				"calendars": {
					"room@example.org": {
						"busy": [{
							"start": "2026-03-10T09:00:00Z",
							"end": "2026-03-10T10:30:00Z"
						}]
					}
				}
			}`))
		case "/calendar/v3/calendars/room@example.org/events":
			_, _ = w.Write([]byte(`{
				"items": [{
					"id": "evt-1",
					"summary": "Malformed Event",
					"start": {"dateTime": "2026-03-10T09:00:00Z"},
					"end": {}
				}]
			}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestCalendarClient(t, &http.Client{}, ClientConfig{
		BaseURL: server.URL,
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)

	events, err := client.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected fallback busy event when detailed event is malformed, got %d events", len(events))
	}
	if events[0].Title != "Unknown event" {
		t.Fatalf("expected unknown fallback title, got %q", events[0].Title)
	}
}

func TestClient_ListRoomEvents_FallsBackToFreeBusyWhenEventsListIsForbidden(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/calendar/v3/freeBusy":
			_, _ = w.Write([]byte(`{
				"calendars": {
					"room@example.org": {
						"busy": [{
							"start": "2026-03-10T09:00:00Z",
							"end": "2026-03-10T10:30:00Z"
						}]
					}
				}
			}`))
		case "/calendar/v3/calendars/room@example.org/events":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":403,"message":"insufficient permissions"}}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestCalendarClient(t, &http.Client{}, ClientConfig{
		BaseURL: server.URL,
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)

	events, err := client.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected busy fallback event, got %d", len(events))
	}
	if events[0].Title != "Unknown event" {
		t.Fatalf("expected unknown fallback title, got %q", events[0].Title)
	}
}

func TestClient_ListRoomEvents_ReturnsErrorOnRateLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/calendar/v3/freeBusy" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":429,"message":"quota exceeded"}}`))
			return
		}
		t.Fatalf("unexpected path %q", r.URL.Path)
	}))
	defer server.Close()

	client := newTestCalendarClient(t, &http.Client{}, ClientConfig{
		BaseURL: server.URL,
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC)

	_, err := client.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err == nil {
		t.Fatalf("expected quota error, got nil")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected error to mention HTTP 429, got %v", err)
	}
}

func TestClient_ListRoomEvents_HandlesTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client := newTestCalendarClient(t, nil, ClientConfig{
		BaseURL: server.URL,
		Timeout: 25 * time.Millisecond,
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC)

	_, err := client.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "timeout") &&
		!strings.Contains(strings.ToLower(err.Error()), "deadline") {
		t.Fatalf("expected timeout/deadline error, got %v", err)
	}
}

func TestClient_ListRoomEvents_ReturnsErrorOnNetworkFailure(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial tcp 127.0.0.1:443: connectex: connection refused")
		}),
	}
	adapter := newTestCalendarClient(t, client, ClientConfig{
		BaseURL: "http://example.invalid",
	})

	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC)

	_, err := adapter.ListRoomEvents(context.Background(), "room@example.org", start, end)
	if err == nil {
		t.Fatalf("expected network error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "connection refused") {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func newTestCalendarClient(t *testing.T, httpClient *http.Client, cfg ClientConfig) *Client {
	t.Helper()

	client, err := NewClient(httpClient, staticTokenProvider{token: "test-token"}, cfg)
	if err != nil {
		t.Fatalf("expected client creation to succeed, got %v", err)
	}

	return client
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
