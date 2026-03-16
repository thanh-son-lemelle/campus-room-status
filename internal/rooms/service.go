package rooms

import (
	"context"
	"time"

	"campus-room-status/internal/domain"
)

type inventoryReader interface {
	GetInventory(ctx context.Context) (domain.InventorySnapshot, error)
}

type eventsReader interface {
	Get(ctx context.Context, key domain.RoomEventsKey) ([]domain.Event, error)
}

type service struct {
	inventory         inventoryReader
	events            eventsReader
	statusInterpreter domain.StatusInterpreter
	clock             domain.Clock
	eventsWindow      time.Duration
	eventsBucket      time.Duration
	maxParallelFetch  int
}

var _ domain.RoomService = (*service)(nil)

// NewService creates a new service.
//
// Summary:
// - Creates a new service.
//
// Attributes:
// - inventory (inventoryReader): Input parameter.
// - events (eventsReader): Input parameter.
// - statusInterpreter (domain.StatusInterpreter): Input parameter.
// - clock (domain.Clock): Input parameter.
//
// Returns:
// - value1 (domain.RoomService): Returned value.
func NewService(
	inventory inventoryReader,
	events eventsReader,
	statusInterpreter domain.StatusInterpreter,
	clock domain.Clock,
) domain.RoomService {
	if clock == nil {
		clock = serviceClock{}
	}
	if statusInterpreter == nil {
		statusInterpreter = domain.NewStatusInterpreter(clock, nil)
	}

	return &service{
		inventory:         inventory,
		events:            events,
		statusInterpreter: statusInterpreter,
		clock:             clock,
		eventsWindow:      7 * 24 * time.Hour,
		eventsBucket:      5 * time.Minute,
		maxParallelFetch:  8,
	}
}

type serviceClock struct{}

// Now nows function behavior.
//
// Summary:
// - Nows function behavior.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (time.Time): Returned value.
func (serviceClock) Now() time.Time {
	return time.Now().UTC()
}
