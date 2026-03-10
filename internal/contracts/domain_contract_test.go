package contracts

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDomainContract_Models(t *testing.T) {
	pkg := parsePackage(t, filepath.Join("..", "domain"))

	requireStructFields(t, pkg, "Building", map[string]string{
		"ID":      "string",
		"Name":    "string",
		"Address": "string",
		"Floors":  "[]int",
	})

	requireStructFields(t, pkg, "Event", map[string]string{
		"Title":     "string",
		"Start":     "time.Time",
		"End":       "time.Time",
		"Organizer": "string",
	})

	roomFields := requireStructFields(t, pkg, "Room", map[string]string{
		"Code":         "string",
		"Name":         "string",
		"Building":     "",
		"Floor":        "int",
		"Capacity":     "int",
		"Type":         "string",
		"Status":       "string",
		"CurrentEvent": "",
		"NextEvent":    "",
	})

	buildingType := exprString(pkg.fset, roomFields["Building"].Type)
	if buildingType != "string" && buildingType != "Building" && buildingType != "*Building" {
		t.Fatalf("expected Room.Building type to model a building identifier/reference, got %q", buildingType)
	}

	currentEventType := exprString(pkg.fset, roomFields["CurrentEvent"].Type)
	if currentEventType != "Event" && currentEventType != "*Event" {
		t.Fatalf("expected Room.CurrentEvent type to be Event or *Event, got %q", currentEventType)
	}

	nextEventType := exprString(pkg.fset, roomFields["NextEvent"].Type)
	if nextEventType != "Event" && nextEventType != "*Event" {
		t.Fatalf("expected Room.NextEvent type to be Event or *Event, got %q", nextEventType)
	}

	filterFields := requireStructFields(t, pkg, "RoomFilters", map[string]string{
		"Building": "*string",
		"Floor":    "*int",
		"Type":     "*string",
		"Status":   "*string",
	})

	capacityField, ok := filterFields["CapacityMin"]
	if !ok {
		capacityField, ok = filterFields["MinCapacity"]
	}
	if !ok {
		t.Fatalf("expected RoomFilters to expose CapacityMin or MinCapacity for optional capacity filtering")
	}

	if got := exprString(pkg.fset, capacityField.Type); got != "*int" {
		t.Fatalf("expected RoomFilters capacity filter to be *int, got %q", got)
	}

	requireStructFields(t, pkg, "HealthStatus", map[string]string{
		"Status":                     "string",
		"Version":                    "string",
		"GoogleAdminAPIConnected":    "bool",
		"GoogleCalendarAPIConnected": "bool",
		"LastSync":                   "*time.Time",
		"ResponseTimeMS":             "int64",
	})

	requireStructFields(t, pkg, "APIError", map[string]string{
		"Code":      "string",
		"Message":   "string",
		"Timestamp": "time.Time",
	})
}

func TestDomainContract_Interfaces(t *testing.T) {
	pkg := parsePackage(t, filepath.Join("..", "domain"))

	requireInterfaceMethodContains(
		t,
		pkg,
		"RoomService",
		"ListRooms",
		[]string{"context.Context", "RoomFilters", "[]Room", "error"},
	)

	requireInterfaceMethodContains(
		t,
		pkg,
		"HealthService",
		"GetHealth",
		[]string{"context.Context", "HealthStatus", "error"},
	)

	requireInterfaceMethodContains(
		t,
		pkg,
		"AdminDirectoryClient",
		"ListRooms",
		[]string{"context.Context", "[]DirectoryRoom", "error"},
	)

	requireInterfaceMethodContains(
		t,
		pkg,
		"CalendarClient",
		"ListRoomEvents",
		[]string{"context.Context", "string", "time.Time", "[]Event", "error"},
	)

	requireInterfaceMethodContains(
		t,
		pkg,
		"StatusInterpreter",
		"Resolve",
		[]string{"Room", "string"},
	)

	requireInterfaceMethodContains(
		t,
		pkg,
		"Clock",
		"Now",
		[]string{"time.Time"},
	)
}

func TestDomainContract_HasNoGoogleImports(t *testing.T) {
	pkg := parsePackage(t, filepath.Join("..", "domain"))

	for _, file := range pkg.files {
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if strings.Contains(strings.ToLower(importPath), "google") {
				t.Fatalf("domain package must stay Google-agnostic, found import %q", importPath)
			}
		}
	}
}
