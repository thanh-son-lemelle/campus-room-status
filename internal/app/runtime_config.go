package app

import (
	"encoding/base64"
	"os"
	"strconv"
	"strings"
	"time"
)

type runtimeDataSource string

const (
	runtimeDataSourceStatic runtimeDataSource = "static"
	runtimeDataSourceGoogle runtimeDataSource = "google"
)

const (
	defaultRuntimeVersion     = "dev"
	defaultInventoryCacheTTL  = time.Hour
	defaultRoomEventsCacheTTL = 5 * time.Minute
)

type runtimeConfig struct {
	dataSource runtimeDataSource

	version            string
	inventoryCacheTTL  time.Duration
	roomEventsCacheTTL time.Duration

	googleAdminBaseURL          string
	googleAdminCustomer         string
	googleAdminPageSize         int
	googleAdminTimeout          time.Duration
	googleAdminImpersonatedUser string
	googleAdminBearerToken      string

	googleCalendarBaseURL  string
	googleCalendarTimeout  time.Duration
	googleCalendarPageSize int
}

// loadRuntimeConfigFromEnv loads runtime config from env.
//
// Summary:
// - Loads runtime config from env.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (runtimeConfig): Returned value.
func loadRuntimeConfigFromEnv() runtimeConfig {
	version := strings.TrimSpace(os.Getenv("APP_VERSION"))
	if version == "" {
		version = defaultRuntimeVersion
	}

	return runtimeConfig{
		dataSource:                  runtimeDataSourceFromEnv(),
		version:                     version,
		inventoryCacheTTL:           envDurationOrDefault("INVENTORY_CACHE_TTL", defaultInventoryCacheTTL),
		roomEventsCacheTTL:          envDurationOrDefault("ROOM_EVENTS_CACHE_TTL", defaultRoomEventsCacheTTL),
		googleAdminBaseURL:          strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_BASE_URL")),
		googleAdminCustomer:         strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_CUSTOMER")),
		googleAdminPageSize:         envInt("GOOGLE_ADMIN_PAGE_SIZE"),
		googleAdminTimeout:          envDuration("GOOGLE_ADMIN_TIMEOUT"),
		googleAdminImpersonatedUser: strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_IMPERSONATED_USER")),
		googleAdminBearerToken:      strings.TrimSpace(os.Getenv("GOOGLE_ADMIN_BEARER_TOKEN")),
		googleCalendarBaseURL:       strings.TrimSpace(os.Getenv("GOOGLE_CALENDAR_BASE_URL")),
		googleCalendarTimeout:       envDuration("GOOGLE_CALENDAR_TIMEOUT"),
		googleCalendarPageSize:      envInt("GOOGLE_CALENDAR_PAGE_SIZE"),
	}
}

// readServiceAccountCredentials reads service account credentials.
//
// Summary:
// - Reads service account credentials.
//
// Attributes:
// - None.
//
// Returns:
// - value1 ([]byte): Returned value.
// - value2 (bool): Returned value.
func readServiceAccountCredentials() ([]byte, bool) {
	rawJSON := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON"))
	if rawJSON != "" {
		return []byte(rawJSON), true
	}

	rawBase64 := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON_BASE64"))
	if rawBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(rawBase64)
		if err == nil && len(decoded) > 0 {
			return decoded, true
		}
	}

	filePath := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE"))
	if filePath == "" {
		return nil, false
	}

	// #nosec G304,G703 -- file path is controlled by trusted deployment configuration.
	credentialsJSON, err := os.ReadFile(filePath)
	if err != nil || len(credentialsJSON) == 0 {
		return nil, false
	}

	return credentialsJSON, true
}

// runtimeDataSourceFromEnv handles runtime data source from env.
//
// Summary:
// - Handles runtime data source from env.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (runtimeDataSource): Returned value.
func runtimeDataSourceFromEnv() runtimeDataSource {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("DATA_SOURCE")))
	if raw == string(runtimeDataSourceGoogle) {
		return runtimeDataSourceGoogle
	}

	return runtimeDataSourceStatic
}

// envInt envs int.
//
// Summary:
// - Envs int.
//
// Attributes:
// - name (string): Input parameter.
//
// Returns:
// - value1 (int): Returned value.
func envInt(name string) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}

	return value
}

// envDuration envs duration.
//
// Summary:
// - Envs duration.
//
// Attributes:
// - name (string): Input parameter.
//
// Returns:
// - value1 (time.Duration): Returned value.
func envDuration(name string) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return 0
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0
	}

	return value
}

// envDurationOrDefault envs duration or default.
//
// Summary:
// - Envs duration or default.
//
// Attributes:
// - name (string): Input parameter.
// - defaultValue (time.Duration): Input parameter.
//
// Returns:
// - value1 (time.Duration): Returned value.
func envDurationOrDefault(name string, defaultValue time.Duration) time.Duration {
	value := envDuration(name)
	if value <= 0 {
		return defaultValue
	}

	return value
}
