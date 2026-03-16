package calendar

import (
	"context"
	"sort"
	"time"

	gcalendar "google.golang.org/api/calendar/v3"
)

type busyInterval struct {
	Start time.Time
	End   time.Time
}

// listBusyIntervals retrieves and normalizes busy intervals for a room calendar.
//
// Summary:
// - Queries Google Calendar FreeBusy data for the provided room and time window.
// - Filters malformed intervals, normalizes times to UTC, and sorts output chronologically.
//
// Attributes:
// - ctx: Request-scoped context for cancellation and deadlines.
// - roomID: Google Calendar identifier for the room calendar.
// - start: Inclusive start of the query window.
// - end: Exclusive end of the query window.
//
// Returns:
// - []busyInterval: Sorted list of valid busy intervals; empty when none are found.
// - error: Google API query error.
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
