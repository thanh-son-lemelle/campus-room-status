package adminsdk

import (
	"reflect"
	"strings"
)

// codeFromEmail codes from email.
//
// Summary:
// - Codes from email.
//
// Attributes:
// - resourceEmail (string): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
func codeFromEmail(resourceEmail string) string {
	at := strings.Index(resourceEmail, "@")
	if at <= 0 {
		return strings.TrimSpace(resourceEmail)
	}
	return strings.TrimSpace(resourceEmail[:at])
}

// firstNonEmpty firsts non empty.
//
// Summary:
// - Firsts non empty.
//
// Attributes:
// - values (...string): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// appendIfNotEmpty appends if not empty.
//
// Summary:
// - Appends if not empty.
//
// Attributes:
// - values ([]string): Input parameter.
// - value (string): Input parameter.
//
// Returns:
// - value1 ([]string): Returned value.
func appendIfNotEmpty(values []string, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return values
	}
	return append(values, trimmed)
}

// trimNonEmpty trims non empty.
//
// Summary:
// - Trims non empty.
//
// Attributes:
// - values ([]string): Input parameter.
//
// Returns:
// - value1 ([]string): Returned value.
func trimNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

// hasAnyValue has any value.
//
// Summary:
// - Has any value.
//
// Attributes:
// - value (any): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
func hasAnyValue(value any) bool {
	if value == nil {
		return false
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return reflected.Len() > 0
	case reflect.Bool:
		return reflected.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflected.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return reflected.Float() != 0
	case reflect.Interface, reflect.Pointer:
		return !reflected.IsNil()
	default:
		return true
	}
}

// slicesContainsString sliceses contains string.
//
// Summary:
// - Sliceses contains string.
//
// Attributes:
// - values ([]string): Input parameter.
// - target (string): Input parameter.
//
// Returns:
// - value1 (bool): Returned value.
func slicesContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// normalizeEndpoint normalizes endpoint.
//
// Summary:
// - Normalizes endpoint.
//
// Attributes:
// - baseURL (string): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
func normalizeEndpoint(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return defaultEndpoint
	}

	trimmed = strings.TrimRight(trimmed, "/")
	trimmed = strings.TrimSuffix(trimmed, "/admin/directory/v1")

	return trimmed + "/"
}
