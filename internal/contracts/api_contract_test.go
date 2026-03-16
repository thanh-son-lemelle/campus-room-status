package contracts

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIContract_ResponseModelsJSONTags(t *testing.T) {
	pkg := parsePackage(t, filepath.Join("..", "api"))

	assertTags(t, pkg, "BuildingResponse", "json", map[string]string{
		"ID":      "id",
		"Name":    "name",
		"Address": "address",
		"Floors":  "floors",
	})

	assertTags(t, pkg, "EventResponse", "json", map[string]string{
		"Title":     "title",
		"Start":     "start",
		"End":       "end",
		"Organizer": "organizer",
	})

	assertTags(t, pkg, "RoomResponse", "json", map[string]string{
		"Code":         "code",
		"Name":         "name",
		"Building":     "building",
		"Floor":        "floor",
		"Capacity":     "capacity",
		"Type":         "type",
		"Status":       "status",
		"CurrentEvent": "current_event",
		"NextEvent":    "next_event",
	})

	assertTags(t, pkg, "HealthResponse", "json", map[string]string{
		"Status":                     "status",
		"Version":                    "version",
		"GoogleAdminAPIConnected":    "google_admin_api_connected",
		"GoogleCalendarAPIConnected": "google_calendar_api_connected",
		"LastSync":                   "last_sync",
		"ResponseTimeMS":             "response_time_ms",
	})

	assertTags(t, pkg, "ErrorResponse", "json", map[string]string{
		"Code":      "code",
		"Message":   "message",
		"Timestamp": "timestamp",
	})

	assertTags(t, pkg, "ErrorEnvelope", "json", map[string]string{
		"Error": "error",
	})
}

func TestAPIContract_RoomsQueryHasOptionalFilters(t *testing.T) {
	pkg := parsePackage(t, filepath.Join("..", "api"))

	fields := requireStructFields(t, pkg, "RoomsQuery", map[string]string{})

	formTags := make(map[string]struct{})
	for _, field := range fields {
		tag := primaryTagValue(fieldTagValue(field, "form"))
		if tag == "" {
			continue
		}
		formTags[tag] = struct{}{}
	}

	required := []string{"building", "type", "status"}
	for _, tag := range required {
		if _, ok := formTags[tag]; !ok {
			t.Fatalf("expected RoomsQuery to expose optional %q filter via form tag", tag)
		}
	}

	hasCapacityFilter := false
	for tag := range formTags {
		if strings.Contains(tag, "capacity") {
			hasCapacityFilter = true
			break
		}
	}

	if !hasCapacityFilter {
		t.Fatalf("expected RoomsQuery to expose an optional capacity filter via form tag")
	}
}

func assertTags(t *testing.T, pkg parsedPackage, typeName string, tagKey string, expected map[string]string) {
	t.Helper()

	fields := requireStructFields(t, pkg, typeName, map[string]string{})

	for fieldName, expectedTag := range expected {
		field, ok := fields[fieldName]
		if !ok {
			t.Fatalf("expected %q.%s field to exist", typeName, fieldName)
		}

		got := primaryTagValue(fieldTagValue(field, tagKey))
		if got != expectedTag {
			t.Fatalf("expected %q.%s %s tag to be %q, got %q", typeName, fieldName, tagKey, expectedTag, got)
		}
	}
}
