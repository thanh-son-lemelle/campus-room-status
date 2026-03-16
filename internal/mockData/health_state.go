package mockdata

import "time"

type InventoryCacheHealthState struct {
	Degraded       bool
	LastRefresh    time.Time
	LastAdminError *time.Time
}

type RoomEventsCacheHealthState struct {
	Degraded                bool
	LastCalendarErrorAt     *time.Time
	LastSuccessfulRefreshAt *time.Time
}
