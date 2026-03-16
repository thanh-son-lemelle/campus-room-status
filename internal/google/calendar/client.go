package calendar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"campus-room-status/internal/domain"
	"campus-room-status/internal/google/httpauth"
	gcalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	defaultEndpoint = "https://www.googleapis.com/calendar/v3/"
	defaultTimeout  = 10 * time.Second
	defaultPageSize = 250
)

type TokenProvider = httpauth.TokenProvider

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

// NewClient creates a new client.
//
// Summary:
// - Creates a new client.
//
// Attributes:
// - httpClient (*http.Client): Input parameter.
// - tokenProvider (TokenProvider): Input parameter.
// - cfg (ClientConfig): Input parameter.
//
// Returns:
// - value1 (*Client): Returned value.
// - value2 (error): Returned value.
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

	authorizedClient := httpauth.NewAuthorizedHTTPClient(baseClient, tokenProvider)
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

// ListRoomEvents lists room events.
//
// Summary:
// - Lists room events.
//
// Attributes:
// - ctx (context.Context): Input parameter.
// - resourceEmail (string): Input parameter.
// - start (time.Time): Input parameter.
// - end (time.Time): Input parameter.
//
// Returns:
// - value1 ([]domain.Event): Returned value.
// - value2 (error): Returned value.
func (c *Client) ListRoomEvents(ctx context.Context, resourceEmail string, start, end time.Time) ([]domain.Event, error) {
	roomID := normalizeRoomID(resourceEmail)
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
		if isRecoverableEventsListError(err) {
			return mergeBusyAndDetailedEvents(busyIntervals, nil), nil
		}
		return nil, fmt.Errorf("events list request failed: %w", err)
	}

	return mergeBusyAndDetailedEvents(busyIntervals, events), nil
}
