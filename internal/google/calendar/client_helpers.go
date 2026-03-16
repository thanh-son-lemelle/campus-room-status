package calendar

import "strings"

func normalizeRoomID(resourceEmail string) string {
	return strings.TrimSpace(resourceEmail)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

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
