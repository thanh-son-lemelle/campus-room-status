package docs

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"campus-room-status/internal/api"
)

func TestSpecJSON_IsValidSwagger2Document(t *testing.T) {
	spec := parseSpec(t)

	version, ok := spec["swagger"].(string)
	if !ok {
		t.Fatalf("expected swagger field to be a string")
	}
	if version != "2.0" {
		t.Fatalf("expected Swagger version 2.0, got %q", version)
	}

	if _, ok := spec["paths"].(map[string]any); !ok {
		t.Fatalf("expected paths object in Swagger document")
	}
}

func TestSpecJSON_ContainsRequiredPaths(t *testing.T) {
	spec := parseSpec(t)
	paths := mustMap(t, spec["paths"], "paths")

	required := []string{
		"/api/v1/docs/openapi.json",
		"/api/v1/health",
		"/api/v1/buildings",
		"/api/v1/rooms",
		"/api/v1/rooms/{code}",
		"/api/v1/rooms/{code}/schedule",
	}

	for _, path := range required {
		if _, ok := paths[path]; !ok {
			t.Fatalf("expected path %q in Swagger document", path)
		}
	}
}

func TestSpecJSON_DefinitionsAlignWithAPIModels(t *testing.T) {
	spec := parseSpec(t)
	definitions := mustMap(t, spec["definitions"], "definitions")

	assertDefinitionHasJSONFields(t, definitions, "BuildingsResponse", api.BuildingsResponse{})
	assertDefinitionHasJSONFields(t, definitions, "RoomResponse", api.RoomResponse{})
	assertDefinitionHasJSONFields(t, definitions, "RoomDetailResponse", api.RoomDetailResponse{})
	assertDefinitionHasJSONFields(t, definitions, "RoomScheduleResponse", api.RoomScheduleResponse{})
	assertDefinitionHasJSONFields(t, definitions, "HealthResponse", api.HealthResponse{})
	assertDefinitionHasJSONFields(t, definitions, "ErrorEnvelope", api.ErrorEnvelope{})
}

func TestSpecJSON_DocumentsRoomParametersAndErrors(t *testing.T) {
	spec := parseSpec(t)

	parameters, ok := getByPath(t, spec, "paths", "/api/v1/rooms", "get", "parameters").([]any)
	if !ok {
		t.Fatalf("expected /api/v1/rooms GET parameters array")
	}

	requiredParams := []string{"building", "type", "status", "capacity_min", "capacity_max", "sort", "order"}
	for _, name := range requiredParams {
		if !hasParameter(parameters, name) {
			t.Fatalf("expected /api/v1/rooms parameter %q", name)
		}
	}

	responses := mustMap(t, getByPath(t, spec, "paths", "/api/v1/rooms", "get", "responses"), "rooms responses")
	if _, ok := responses["400"]; !ok {
		t.Fatalf("expected /api/v1/rooms to document 400 response")
	}
	if _, ok := responses["503"]; !ok {
		t.Fatalf("expected /api/v1/rooms to document 503 response")
	}
}

func parseSpec(t *testing.T) map[string]any {
	t.Helper()

	specRaw := SpecJSON()
	if len(specRaw) == 0 {
		t.Fatalf("expected embedded Swagger spec to be non-empty")
	}

	var spec map[string]any
	if err := json.Unmarshal(specRaw, &spec); err != nil {
		t.Fatalf("expected valid JSON Swagger spec: %v", err)
	}
	return spec
}

func mustMap(t *testing.T, value any, name string) map[string]any {
	t.Helper()

	out, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected %s to be an object, got %T", name, value)
	}
	return out
}

func getByPath(t *testing.T, root map[string]any, segments ...string) any {
	t.Helper()

	var current any = root
	for i := range segments {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("expected object while resolving %q, got %T", strings.Join(segments[:i], "."), current)
		}
		next, ok := object[segments[i]]
		if !ok {
			t.Fatalf("expected key %q while resolving %q", segments[i], strings.Join(segments[:i+1], "."))
		}
		current = next
	}
	return current
}

func assertDefinitionHasJSONFields(t *testing.T, definitions map[string]any, definitionSuffix string, model any) {
	t.Helper()

	rawDefinition, definitionName, ok := findDefinition(definitions, definitionSuffix)
	if !ok {
		t.Fatalf("expected definition for %q", definitionSuffix)
	}

	definition := mustMap(t, rawDefinition, definitionName)
	properties := mustMap(t, definition["properties"], definitionName+".properties")

	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Struct {
		t.Fatalf("model for %s must be a struct, got %s", definitionSuffix, modelType.Kind())
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		jsonField := strings.Split(tag, ",")[0]
		if jsonField == "" || jsonField == "-" {
			continue
		}

		if _, ok := properties[jsonField]; !ok {
			t.Fatalf("expected definition %q to contain property %q", definitionName, jsonField)
		}
	}
}

func findDefinition(definitions map[string]any, suffix string) (any, string, bool) {
	if value, ok := definitions[suffix]; ok {
		return value, suffix, true
	}

	suffixWithDot := "." + suffix
	for name, value := range definitions {
		if strings.HasSuffix(name, suffixWithDot) {
			return value, name, true
		}
	}

	return nil, "", false
}

func hasParameter(params []any, name string) bool {
	for _, raw := range params {
		param, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rawName, ok := param["name"].(string)
		if !ok {
			continue
		}
		if rawName == name {
			return true
		}
	}
	return false
}
