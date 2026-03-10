package domain

import (
	"context"
	"time"
)

const (
	RoomStatusAvailable   = "available"
	RoomStatusOccupied    = "occupied"
	RoomStatusUpcoming    = "upcoming"
	RoomStatusMaintenance = "maintenance"
)

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

type roomStatusInterpreter struct {
	clock              Clock
	unavailability     UnavailabilitySource
	upcomingWindowSize time.Duration
}

func NewStatusInterpreter(clock Clock, unavailability UnavailabilitySource) StatusInterpreter {
	if clock == nil {
		clock = systemClock{}
	}

	return &roomStatusInterpreter{
		clock:              clock,
		unavailability:     unavailability,
		upcomingWindowSize: 30 * time.Minute,
	}
}

// Resolve determines room status from:
// 1) optional external unavailability source (maintenance only if explicitly confirmed)
// 2) calendar events for occupied/upcoming
// 3) fallback to available
//
// Priority order: maintenance > occupied > upcoming > available.
func (i *roomStatusInterpreter) Resolve(ctx context.Context, room DirectoryRoom, events []Event) string {
	now := i.clock.Now()

	if i.isUnavailable(ctx, room, now) {
		return RoomStatusMaintenance
	}

	if hasRunningEvent(events, now) {
		return RoomStatusOccupied
	}

	if hasImminentEvent(events, now, i.upcomingWindowSize) {
		return RoomStatusUpcoming
	}

	return RoomStatusAvailable
}

func (i *roomStatusInterpreter) isUnavailable(ctx context.Context, room DirectoryRoom, now time.Time) bool {
	if i.unavailability == nil {
		return false
	}

	roomRef := room.ResourceEmail
	if roomRef == "" {
		roomRef = room.ResourceName
	}

	unavailable, err := i.unavailability.IsRoomUnavailable(ctx, roomRef, now)
	if err != nil {
		return false
	}

	return unavailable
}

func hasRunningEvent(events []Event, now time.Time) bool {
	for _, event := range events {
		if !now.Before(event.Start) && now.Before(event.End) {
			return true
		}
	}
	return false
}

func hasImminentEvent(events []Event, now time.Time, window time.Duration) bool {
	for _, event := range events {
		if !event.Start.After(now) {
			continue
		}
		if event.Start.Sub(now) < window {
			return true
		}
	}
	return false
}
