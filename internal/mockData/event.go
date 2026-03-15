package mockdata

import "time"

type Event struct {
	Title     string
	Start     time.Time
	End       time.Time
	Organizer string
}

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

func DetailOrderedEvents(now time.Time) []Event {
	return []Event{
		{Title: "Third", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
		{Title: "First", Start: now.Add(-30 * time.Minute), End: now.Add(15 * time.Minute)},
		{Title: "Second", Start: now.Add(time.Hour), End: now.Add(90 * time.Minute)},
	}
}

func ScheduleOrderedEvents(now time.Time) []Event {
	return []Event{
		{Title: "Third", Start: now.Add(4 * time.Hour), End: now.Add(5 * time.Hour)},
		{Title: "First", Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)},
		{Title: "Second", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
	}
}

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

func OccupiedEvent(now time.Time) []Event {
	return []Event{{
		Title: "Current Event",
		Start: now.Add(-5 * time.Minute),
		End:   now.Add(5 * time.Minute),
	}}
}

func UpcomingEvent(now time.Time) []Event {
	return []Event{{
		Title: "Distributed Systems",
		Start: now.Add(29*time.Minute + 30*time.Second),
		End:   now.Add(89 * time.Minute),
	}}
}

func FutureEvent(now time.Time) []Event {
	return []Event{{
		Title: "Future Session",
		Start: now.Add(45 * time.Minute),
		End:   now.Add(2 * time.Hour),
	}}
}

func NoonSessionEvent(now time.Time) []Event {
	return []Event{{
		Title: "Noon Session",
		Start: now.Add(3 * time.Hour),
		End:   now.Add(4 * time.Hour),
	}}
}
