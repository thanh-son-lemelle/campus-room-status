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
	"google.golang.org/api/googleapi"
)

func isRecoverableEventsListError(err error) bool {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.Code == http.StatusForbidden || apiErr.Code == http.StatusNotFound
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
			MaxResults(c.pageSize).
			AlwaysIncludeEmail(true)

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
	if organizer == "" && item.Creator != nil {
		organizer = firstNonEmpty(item.Creator.DisplayName, item.Creator.Email)
	}
	if organizer == "" {
		organizer = "Google Calendar"
	}

	title := firstNonEmpty(item.Summary, item.Description)
	if title == "" {
		title = "Busy"
	}

	return domain.Event{
		Title:     title,
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
