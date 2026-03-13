package calendar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"campus-room-status/internal/domain"
	gcalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	defaultEndpoint = "https://www.googleapis.com/calendar/v3/"
	defaultTimeout  = 10 * time.Second
	defaultPageSize = 250
)

type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

type ClientConfig struct {
	BaseURL  string
	Timeout  time.Duration
	PageSize int
}

type Client struct {
	service  *gcalendar.Service
	pageSize int64
}

var _ domain.CalendarClient = (*Client)(nil)

func NewClient(httpClient *http.Client, tokenProvider TokenProvider, cfg ClientConfig) (*Client, error) {
	if tokenProvider == nil {
		return nil, errors.New("token provider is required")
	}

	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	baseClient := httpClient
	if baseClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		baseClient = &http.Client{Timeout: timeout}
	}

	authorizedClient := newAuthorizedHTTPClient(baseClient, tokenProvider)
	service, err := gcalendar.NewService(
		context.Background(),
		option.WithHTTPClient(authorizedClient),
		option.WithEndpoint(normalizeEndpoint(cfg.BaseURL)),
		option.WithUserAgent("campus-room-status/calendar"),
	)
	if err != nil {
		return nil, fmt.Errorf("create calendar service: %w", err)
	}

	return &Client{
		service:  service,
		pageSize: int64(pageSize),
	}, nil
}

func (c *Client) ListRoomEvents(ctx context.Context, resourceEmail string, start, end time.Time) ([]domain.Event, error) {
	roomID := strings.TrimSpace(resourceEmail)
	if roomID == "" {
		return nil, errors.New("room email is required")
	}
	start = start.UTC()
	end = end.UTC()
	if !end.After(start) {
		return nil, errors.New("end must be after start")
	}

	busyIntervals, err := c.listBusyIntervals(ctx, roomID, start, end)
	if err != nil {
		return nil, fmt.Errorf("freebusy request failed: %w", err)
	}

	events, err := c.listDetailedEvents(ctx, roomID, start, end)
	if err != nil {
		return nil, fmt.Errorf("events list request failed: %w", err)
	}

	return mergeBusyAndDetailedEvents(busyIntervals, events), nil
}

type busyInterval struct {
	Start time.Time
	End   time.Time
}

func (c *Client) listBusyIntervals(ctx context.Context, roomID string, start, end time.Time) ([]busyInterval, error) {
	request := &gcalendar.FreeBusyRequest{
		TimeMin: start.Format(time.RFC3339),
		TimeMax: end.Format(time.RFC3339),
		Items: []*gcalendar.FreeBusyRequestItem{
			{Id: roomID},
		},
	}

	response, err := c.service.Freebusy.Query(request).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	if response == nil || len(response.Calendars) == 0 {
		return nil, nil
	}

	calendar, found := response.Calendars[roomID]
	if !found && len(response.Calendars) == 1 {
		for _, item := range response.Calendars {
			calendar = item
			found = true
		}
	}
	if !found {
		return nil, nil
	}

	out := make([]busyInterval, 0, len(calendar.Busy))
	for _, interval := range calendar.Busy {
		if interval == nil {
			continue
		}

		intervalStart, err := time.Parse(time.RFC3339, interval.Start)
		if err != nil {
			continue
		}
		intervalEnd, err := time.Parse(time.RFC3339, interval.End)
		if err != nil {
			continue
		}
		intervalStart = intervalStart.UTC()
		intervalEnd = intervalEnd.UTC()
		if !intervalEnd.After(intervalStart) {
			continue
		}

		out = append(out, busyInterval{
			Start: intervalStart,
			End:   intervalEnd,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].End.Before(out[j].End)
		}
		return out[i].Start.Before(out[j].Start)
	})

	return out, nil
}

func (c *Client) listDetailedEvents(ctx context.Context, roomID string, start, end time.Time) ([]domain.Event, error) {
	pageToken := ""
	visitedTokens := make(map[string]struct{})
	out := make([]domain.Event, 0, c.pageSize)

	for {
		call := c.service.Events.List(roomID).
			TimeMin(start.Format(time.RFC3339)).
			TimeMax(end.Format(time.RFC3339)).
			SingleEvents(true).
			OrderBy("startTime").
			MaxResults(c.pageSize)

		if pageToken != "" {
			if _, exists := visitedTokens[pageToken]; exists {
				return nil, fmt.Errorf("detected repeated page token %q on events endpoint", pageToken)
			}
			visitedTokens[pageToken] = struct{}{}
			call = call.PageToken(pageToken)
		}

		response, err := call.Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		if response == nil {
			break
		}

		for _, item := range response.Items {
			mapped, ok := mapCalendarEvent(item)
			if !ok {
				continue
			}
			out = append(out, mapped)
		}

		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].End.Before(out[j].End)
		}
		return out[i].Start.Before(out[j].Start)
	})

	return out, nil
}

func mapCalendarEvent(item *gcalendar.Event) (domain.Event, bool) {
	if item == nil {
		return domain.Event{}, false
	}
	if strings.EqualFold(strings.TrimSpace(item.Status), "cancelled") {
		return domain.Event{}, false
	}

	start, ok := parseEventDateTime(item.Start)
	if !ok {
		return domain.Event{}, false
	}
	end, ok := parseEventDateTime(item.End)
	if !ok || !end.After(start) {
		return domain.Event{}, false
	}

	organizer := ""
	if item.Organizer != nil {
		organizer = firstNonEmpty(item.Organizer.DisplayName, item.Organizer.Email)
	}

	return domain.Event{
		Title:     firstNonEmpty(item.Summary, item.Id, "Busy"),
		Start:     start,
		End:       end,
		Organizer: organizer,
	}, true
}

func parseEventDateTime(dateTime *gcalendar.EventDateTime) (time.Time, bool) {
	if dateTime == nil {
		return time.Time{}, false
	}

	if strings.TrimSpace(dateTime.DateTime) != "" {
		parsed, err := time.Parse(time.RFC3339, dateTime.DateTime)
		if err != nil {
			return time.Time{}, false
		}
		return parsed.UTC(), true
	}

	if strings.TrimSpace(dateTime.Date) != "" {
		parsed, err := time.Parse("2006-01-02", dateTime.Date)
		if err != nil {
			return time.Time{}, false
		}
		return parsed.UTC(), true
	}

	return time.Time{}, false
}

func mergeBusyAndDetailedEvents(busy []busyInterval, detailed []domain.Event) []domain.Event {
	out := make([]domain.Event, 0, len(detailed)+len(busy))
	out = append(out, detailed...)

	for _, interval := range busy {
		if busyIntervalCoveredByDetailedEvents(interval, detailed) {
			continue
		}

		out = append(out, domain.Event{
			Title:     "Busy",
			Start:     interval.Start,
			End:       interval.End,
			Organizer: "Google Calendar",
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].End.Before(out[j].End)
		}
		return out[i].Start.Before(out[j].Start)
	})

	return out
}

func busyIntervalCoveredByDetailedEvents(interval busyInterval, detailed []domain.Event) bool {
	for _, event := range detailed {
		if !event.End.After(interval.Start) {
			continue
		}
		if !event.Start.Before(interval.End) {
			continue
		}
		return true
	}

	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeEndpoint(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return defaultEndpoint
	}

	trimmed = strings.TrimRight(trimmed, "/")
	if !strings.HasSuffix(trimmed, "/calendar/v3") {
		trimmed += "/calendar/v3"
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
