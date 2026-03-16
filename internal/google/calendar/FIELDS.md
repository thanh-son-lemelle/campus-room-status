# Google Calendar API fields used by the adapter

This package reads Google Calendar API endpoints:

- `POST /calendar/v3/freeBusy`
- `GET /calendar/v3/calendars/{calendarId}/events`

Implementation note: the adapter uses the official Go SDK
`google.golang.org/api/calendar/v3`.

Mapped FreeBusy fields:

- `calendars.{id}.busy[].start` -> synthetic `domain.Event.Start`
- `calendars.{id}.busy[].end` -> synthetic `domain.Event.End`
- synthetic fallback title: `Busy`

Mapped Events fields:

- `summary` fallback `description` fallback synthetic `Busy` -> `domain.Event.Title`
- `start.dateTime` fallback `start.date` -> `domain.Event.Start`
- `end.dateTime` fallback `end.date` -> `domain.Event.End`
- `organizer.displayName` fallback `organizer.email` fallback `creator.*` fallback synthetic `Google Calendar` -> `domain.Event.Organizer`
- cancelled events are ignored

Behavior:

- detailed events are fetched for rich payloads (`current_event`, `next_event`, `schedule_today`)
- freeBusy intervals are merged as fallback busy events when no detailed event overlaps
- malformed/partial event items are ignored without failing the whole request
