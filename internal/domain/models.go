package domain

import "time"

type Building struct {
	ID      string
	Name    string
	Address string
	Floors  []int
}

type Event struct {
	Title     string
	Start     time.Time
	End       time.Time
	Organizer string
}

type Room struct {
	Code         string
	Name         string
	Building     string
	Floor        int
	Capacity     int
	Type         string
	Status       string
	CurrentEvent *Event
	NextEvent    *Event
}

type RoomFilters struct {
	Building    *string
	Floor       *int
	Type        *string
	Status      *string
	CapacityMin *int
}

type HealthStatus struct {
	Status                     string
	Version                    string
	GoogleAdminAPIConnected    bool
	GoogleCalendarAPIConnected bool
	LastSync                   *time.Time
	ResponseTimeMS             int64
}
