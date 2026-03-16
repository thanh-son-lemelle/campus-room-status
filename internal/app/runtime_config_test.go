package app

import (
	"testing"
	"time"
)

func TestLoadRuntimeConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("DATA_SOURCE", "")
	t.Setenv("APP_VERSION", "")
	t.Setenv("INVENTORY_CACHE_TTL", "")
	t.Setenv("ROOM_EVENTS_CACHE_TTL", "")
	t.Setenv("GOOGLE_ADMIN_BASE_URL", "")
	t.Setenv("GOOGLE_ADMIN_CUSTOMER", "")
	t.Setenv("GOOGLE_ADMIN_PAGE_SIZE", "")
	t.Setenv("GOOGLE_ADMIN_TIMEOUT", "")
	t.Setenv("GOOGLE_ADMIN_IMPERSONATED_USER", "")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "")
	t.Setenv("GOOGLE_CALENDAR_BASE_URL", "")
	t.Setenv("GOOGLE_CALENDAR_TIMEOUT", "")
	t.Setenv("GOOGLE_CALENDAR_PAGE_SIZE", "")

	cfg := loadRuntimeConfigFromEnv()

	if cfg.dataSource != runtimeDataSourceStatic {
		t.Fatalf("expected default data source %q, got %q", runtimeDataSourceStatic, cfg.dataSource)
	}
	if cfg.version != defaultRuntimeVersion {
		t.Fatalf("expected default version %q, got %q", defaultRuntimeVersion, cfg.version)
	}
	if cfg.inventoryCacheTTL != defaultInventoryCacheTTL {
		t.Fatalf("expected default inventory TTL %v, got %v", defaultInventoryCacheTTL, cfg.inventoryCacheTTL)
	}
	if cfg.roomEventsCacheTTL != defaultRoomEventsCacheTTL {
		t.Fatalf("expected default room events TTL %v, got %v", defaultRoomEventsCacheTTL, cfg.roomEventsCacheTTL)
	}
}

func TestLoadRuntimeConfigFromEnv_OverridesAndTTLParsing(t *testing.T) {
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("APP_VERSION", "1.2.3")
	t.Setenv("INVENTORY_CACHE_TTL", "2h")
	t.Setenv("ROOM_EVENTS_CACHE_TTL", "11m")
	t.Setenv("GOOGLE_ADMIN_BASE_URL", "https://admin.example.com")
	t.Setenv("GOOGLE_ADMIN_CUSTOMER", "my_customer")
	t.Setenv("GOOGLE_ADMIN_PAGE_SIZE", "77")
	t.Setenv("GOOGLE_ADMIN_TIMEOUT", "8s")
	t.Setenv("GOOGLE_ADMIN_IMPERSONATED_USER", "admin@example.org")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "token")
	t.Setenv("GOOGLE_CALENDAR_BASE_URL", "https://calendar.example.com")
	t.Setenv("GOOGLE_CALENDAR_TIMEOUT", "9s")
	t.Setenv("GOOGLE_CALENDAR_PAGE_SIZE", "88")

	cfg := loadRuntimeConfigFromEnv()

	if cfg.dataSource != runtimeDataSourceGoogle {
		t.Fatalf("expected data source %q, got %q", runtimeDataSourceGoogle, cfg.dataSource)
	}
	if cfg.version != "1.2.3" {
		t.Fatalf("expected version %q, got %q", "1.2.3", cfg.version)
	}
	if cfg.inventoryCacheTTL != 2*time.Hour {
		t.Fatalf("expected inventory TTL 2h, got %v", cfg.inventoryCacheTTL)
	}
	if cfg.roomEventsCacheTTL != 11*time.Minute {
		t.Fatalf("expected room events TTL 11m, got %v", cfg.roomEventsCacheTTL)
	}
	if cfg.googleAdminBaseURL != "https://admin.example.com" {
		t.Fatalf("expected admin base URL override, got %q", cfg.googleAdminBaseURL)
	}
	if cfg.googleAdminCustomer != "my_customer" {
		t.Fatalf("expected admin customer override, got %q", cfg.googleAdminCustomer)
	}
	if cfg.googleAdminPageSize != 77 {
		t.Fatalf("expected admin page size override 77, got %d", cfg.googleAdminPageSize)
	}
	if cfg.googleAdminTimeout != 8*time.Second {
		t.Fatalf("expected admin timeout override 8s, got %v", cfg.googleAdminTimeout)
	}
	if cfg.googleAdminImpersonatedUser != "admin@example.org" {
		t.Fatalf("expected impersonated user override, got %q", cfg.googleAdminImpersonatedUser)
	}
	if cfg.googleAdminBearerToken != "token" {
		t.Fatalf("expected bearer token override, got %q", cfg.googleAdminBearerToken)
	}
	if cfg.googleCalendarBaseURL != "https://calendar.example.com" {
		t.Fatalf("expected calendar base URL override, got %q", cfg.googleCalendarBaseURL)
	}
	if cfg.googleCalendarTimeout != 9*time.Second {
		t.Fatalf("expected calendar timeout override 9s, got %v", cfg.googleCalendarTimeout)
	}
	if cfg.googleCalendarPageSize != 88 {
		t.Fatalf("expected calendar page size override 88, got %d", cfg.googleCalendarPageSize)
	}
}

func TestLoadRuntimeConfigFromEnv_InvalidTTLFallsBackToDefault(t *testing.T) {
	t.Setenv("INVENTORY_CACHE_TTL", "invalid")
	t.Setenv("ROOM_EVENTS_CACHE_TTL", "0s")

	cfg := loadRuntimeConfigFromEnv()

	if cfg.inventoryCacheTTL != defaultInventoryCacheTTL {
		t.Fatalf("expected fallback inventory TTL %v, got %v", defaultInventoryCacheTTL, cfg.inventoryCacheTTL)
	}
	if cfg.roomEventsCacheTTL != defaultRoomEventsCacheTTL {
		t.Fatalf("expected fallback room events TTL %v, got %v", defaultRoomEventsCacheTTL, cfg.roomEventsCacheTTL)
	}
}
