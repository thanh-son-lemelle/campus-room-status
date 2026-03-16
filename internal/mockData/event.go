package mockdata

import "time"

type Event struct {
	Title     string
	Start     time.Time
	End       time.Time
	Organizer string
}

// RoomServiceEventsByRoom rooms service events by room.
//
// Summary:
// - Rooms service events by room.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 (map[string][]Event): Returned value.
func RoomServiceEventsByRoom(now time.Time) map[string][]Event {
	return map[string][]Event{
		"AMPHI-A": {
			{
				Title: "Algorithms",
				Start: now.Add(-10 * time.Minute),
				End:   now.Add(20 * time.Minute),
			},
			{
				Title: "Security",
				Start: now.Add(40 * time.Minute),
				End:   now.Add(100 * time.Minute),
			},
		},
		"LAB-204": {
			{
				Title: "OS Lab",
				Start: now.Add(20 * time.Minute),
				End:   now.Add(80 * time.Minute),
			},
		},
		"LAB-101": {},
	}
}

// DetailOrderedEvents details ordered events.
//
// Summary:
// - Details ordered events.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func DetailOrderedEvents(now time.Time) []Event {
	return []Event{
		{Title: "Third", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
		{Title: "First", Start: now.Add(-30 * time.Minute), End: now.Add(15 * time.Minute)},
		{Title: "Second", Start: now.Add(time.Hour), End: now.Add(90 * time.Minute)},
	}
}

// ScheduleOrderedEvents schedules ordered events.
//
// Summary:
// - Schedules ordered events.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func ScheduleOrderedEvents(now time.Time) []Event {
	return []Event{
		{Title: "Third", Start: now.Add(4 * time.Hour), End: now.Add(5 * time.Hour)},
		{Title: "First", Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)},
		{Title: "Second", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
	}
}

// CacheEvents caches events.
//
// Summary:
// - Caches events.
//
// Attributes:
// - title (string): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func CacheEvents(title string) []Event {
	start := time.Date(2026, time.March, 10, 8, 0, 0, 0, time.UTC)
	return []Event{
		{
			Title:     title,
			Start:     start,
			End:       start.Add(90 * time.Minute),
			Organizer: "Academic Office",
		},
	}
}

// OccupiedEvent occupieds event.
//
// Summary:
// - Occupieds event.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func OccupiedEvent(now time.Time) []Event {
	return []Event{{
		Title: "Current Event",
		Start: now.Add(-5 * time.Minute),
		End:   now.Add(5 * time.Minute),
	}}
}

// UpcomingEvent upcomings event.
//
// Summary:
// - Upcomings event.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func UpcomingEvent(now time.Time) []Event {
	return []Event{{
		Title: "Distributed Systems",
		Start: now.Add(29*time.Minute + 30*time.Second),
		End:   now.Add(89 * time.Minute),
	}}
}

// FutureEvent futures event.
//
// Summary:
// - Futures event.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func FutureEvent(now time.Time) []Event {
	return []Event{{
		Title: "Future Session",
		Start: now.Add(45 * time.Minute),
		End:   now.Add(2 * time.Hour),
	}}
}

// NoonSessionEvent noons session event.
//
// Summary:
// - Noons session event.
//
// Attributes:
// - now (time.Time): Input parameter.
//
// Returns:
// - value1 ([]Event): Returned value.
func NoonSessionEvent(now time.Time) []Event {
	return []Event{{
		Title: "Noon Session",
		Start: now.Add(3 * time.Hour),
		End:   now.Add(4 * time.Hour),
	}}
}
