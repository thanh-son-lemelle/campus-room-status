package calendar

import "strings"

// normalizeRoomID normalizes room id.
//
// Summary:
// - Normalizes room id.
//
// Attributes:
// - resourceEmail (string): Input parameter.
//
// Returns:
// - value1 (string): Returned value.
func normalizeRoomID(resourceEmail string) string {
	return strings.TrimSpace(resourceEmail)
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
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
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
	if !strings.HasSuffix(trimmed, "/calendar/v3") {
		trimmed += "/calendar/v3"
	}

	return trimmed + "/"
}
