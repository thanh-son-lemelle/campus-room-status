package api

import "time"

type BuildingResponse struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Address string   `json:"address"`
	Floors  []string `json:"floors"`
}

type BuildingsResponse struct {
	Timestamp time.Time          `json:"timestamp"`
	Buildings []BuildingResponse `json:"buildings"`
}

type EventResponse struct {
	Title     string    `json:"title"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Organizer string    `json:"organizer"`
}

type RoomResponse struct {
	Code          string         `json:"code"`
	ResourceEmail string         `json:"resource_email,omitempty"`
	Name          string         `json:"name"`
	Building      string         `json:"building"`
	Floor         int            `json:"floor"`
	Capacity      int            `json:"capacity"`
	Type          string         `json:"type"`
	Status        string         `json:"status"`
	CurrentEvent  *EventResponse `json:"current_event"`
	NextEvent     *EventResponse `json:"next_event"`
}

type RoomDetailResponse struct {
	Code          string          `json:"code"`
	ResourceEmail string          `json:"resource_email,omitempty"`
	Name          string          `json:"name"`
	Building      string          `json:"building"`
	Floor         int             `json:"floor"`
	Capacity      int             `json:"capacity"`
	Type          string          `json:"type"`
	Status        string          `json:"status"`
	CurrentEvent  *EventResponse  `json:"current_event"`
	NextEvent     *EventResponse  `json:"next_event"`
	ScheduleToday []EventResponse `json:"schedule_today"`
}

type PeriodResponse struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type RoomScheduleResponse struct {
	RoomCode string          `json:"room_code"`
	Period   PeriodResponse  `json:"period"`
	Events   []EventResponse `json:"events"`
}

type RoomsListResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Filters   map[string]interface{} `json:"filters"`
	Count     int                    `json:"count"`
	Rooms     []RoomResponse         `json:"rooms"`
}

type HealthResponse struct {
	Status                     string     `json:"status"`
	Version                    string     `json:"version"`
	GoogleAdminAPIConnected    bool       `json:"google_admin_api_connected"`
	GoogleCalendarAPIConnected bool       `json:"google_calendar_api_connected"`
	LastSync                   *time.Time `json:"last_sync"`
	ResponseTimeMS             int64      `json:"response_time_ms"`
}

type RoomsQuery struct {
	Building    *string `form:"building"`
	Type        *string `form:"type"`
	Status      *string `form:"status"`
	CapacityMin *int    `form:"capacity_min"`
	CapacityMax *int    `form:"capacity_max"`
	Sort        *string `form:"sort"`
	Order       *string `form:"order"`
}

type ErrorResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorEnvelope struct {
	Error ErrorResponse `json:"error"`
}
