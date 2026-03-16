package domain

import "time"

type Building struct {
	ID      string
	Name    string
	Address string
	Floors  []string
}

type Event struct {
	Title     string
	Start     time.Time
	End       time.Time
	Organizer string
}

// DirectoryRoom is a provider-agnostic room snapshot coming from an external directory.
// It intentionally mirrors common fields (resourceName/resourceEmail/capacity/category)
// without importing any provider-specific SDK type.
type DirectoryRoom struct {
	ResourceName     string
	ResourceEmail    string
	Capacity         int
	ResourceType     string
	ResourceCategory string
}

type Room struct {
	Code          string
	ResourceEmail string
	Name          string
	Building      string
	Floor         int
	Capacity      int
	Type          string
	Status        string
	CurrentEvent  *Event
	NextEvent     *Event
}

type RoomFilters struct {
	Building    *string
	Floor       *int
	Type        *string
	Status      *string
	CapacityMin *int
	CapacityMax *int
	Sort        *string
	Order       *string
}

type HealthStatus struct {
	Status                     string
	Version                    string
	GoogleAdminAPIConnected    bool
	GoogleCalendarAPIConnected bool
	LastSync                   *time.Time
	ResponseTimeMS             int64
}

type APIError struct {
	Code      string
	Message   string
	Timestamp time.Time
}
