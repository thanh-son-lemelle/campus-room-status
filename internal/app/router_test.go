package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"campus-room-status/internal/domain"
	"campus-room-status/internal/google/adminsdk"
	gcalendar "campus-room-status/internal/google/calendar"
	goauth "campus-room-status/internal/google/oauth"
	"github.com/gin-gonic/gin"
)

func TestNewRuntimeInventorySource_UsesStaticSourceByDefault(t *testing.T) {
	clearOAuthEnv(t)
	clearDataSourceEnv(t)
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "test-token")

	source := newRuntimeInventorySource()

	if _, ok := source.(staticInventorySource); !ok {
		t.Fatalf("expected staticInventorySource by default, got %T", source)
	}
}

func TestNewRuntimeInventorySource_UsesStaticSourceWhenConfigured(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "test-token")
	t.Setenv("DATA_SOURCE", "static")

	source := newRuntimeInventorySource()

	if _, ok := source.(staticInventorySource); !ok {
		t.Fatalf("expected staticInventorySource when DATA_SOURCE=static, got %T", source)
	}
}

func TestNewRuntimeInventorySource_UsesAdminSDKSourceWhenGoogleSourceAndTokenPresent(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "test-token")
	t.Setenv("GOOGLE_ADMIN_BASE_URL", "https://admin.googleapis.com")
	t.Setenv("GOOGLE_ADMIN_CUSTOMER", "my_customer")
	t.Setenv("GOOGLE_ADMIN_TIMEOUT", "3s")
	t.Setenv("GOOGLE_ADMIN_PAGE_SIZE", "50")

	source := newRuntimeInventorySource()

	wrapped, ok := source.(oauthBootstrapInventorySource)
	if !ok {
		t.Fatalf("expected oauthBootstrapInventorySource when DATA_SOURCE=google and GOOGLE_ADMIN_BEARER_TOKEN is present, got %T", source)
	}
	if _, ok := wrapped.primary.(*adminsdk.InventorySource); !ok {
		t.Fatalf("expected wrapped primary source to be *adminsdk.InventorySource, got %T", wrapped.primary)
	}
}

func TestNewRuntimeInventorySource_UsesAdminSDKSourceWhenGoogleSourceAndServiceAccountJSONIsPresent(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", makeRouterTestServiceAccountJSON(t))
	t.Setenv("GOOGLE_ADMIN_IMPERSONATED_USER", "admin@example.org")

	source := newRuntimeInventorySource()

	wrapped, ok := source.(oauthBootstrapInventorySource)
	if !ok {
		t.Fatalf("expected oauthBootstrapInventorySource when DATA_SOURCE=google and GOOGLE_SERVICE_ACCOUNT_JSON is present, got %T", source)
	}
	if _, ok := wrapped.primary.(*adminsdk.InventorySource); !ok {
		t.Fatalf("expected wrapped primary source to be *adminsdk.InventorySource, got %T", wrapped.primary)
	}
}

func TestNewRuntimeInventorySource_DoesNotFallbackToStaticWhenGoogleSourceAndNoProvider(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON_BASE64", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_ADMIN_IMPERSONATED_USER", "")

	source := newRuntimeInventorySource()

	if _, ok := source.(staticInventorySource); ok {
		t.Fatalf("expected non-static source when DATA_SOURCE=google")
	}
	if _, ok := source.(unavailableInventorySource); !ok {
		t.Fatalf("expected unavailableInventorySource when DATA_SOURCE=google and no provider, got %T", source)
	}
}

func TestNewRuntimeCalendarClient_UsesStaticClientByDefault(t *testing.T) {
	clearOAuthEnv(t)
	clearDataSourceEnv(t)
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "test-token")

	client := newRuntimeCalendarClient()

	if _, ok := client.(staticCalendarClient); !ok {
		t.Fatalf("expected staticCalendarClient by default, got %T", client)
	}
}

func TestNewRuntimeCalendarClient_UsesStaticClientWhenConfigured(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("DATA_SOURCE", "static")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "test-token")

	client := newRuntimeCalendarClient()

	if _, ok := client.(staticCalendarClient); !ok {
		t.Fatalf("expected staticCalendarClient when DATA_SOURCE=static, got %T", client)
	}
}

func TestNewRuntimeCalendarClient_UsesGoogleClientWhenGoogleSourceAndTokenPresent(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "test-token")
	t.Setenv("GOOGLE_CALENDAR_BASE_URL", "https://www.googleapis.com")
	t.Setenv("GOOGLE_CALENDAR_TIMEOUT", "3s")
	t.Setenv("GOOGLE_CALENDAR_PAGE_SIZE", "100")

	client := newRuntimeCalendarClient()

	if _, ok := client.(*gcalendar.Client); !ok {
		t.Fatalf("expected *calendar.Client when DATA_SOURCE=google and GOOGLE_ADMIN_BEARER_TOKEN is present, got %T", client)
	}
}

func TestNewRuntimeCalendarClient_UsesGoogleClientWhenGoogleSourceAndServiceAccountJSONIsPresent(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", makeRouterTestServiceAccountJSON(t))
	t.Setenv("GOOGLE_ADMIN_IMPERSONATED_USER", "admin@example.org")

	client := newRuntimeCalendarClient()

	if _, ok := client.(*gcalendar.Client); !ok {
		t.Fatalf("expected *calendar.Client when DATA_SOURCE=google and GOOGLE_SERVICE_ACCOUNT_JSON is present, got %T", client)
	}
}

func TestNewRuntimeOAuthTokenProvider_FailsWhenConfigMissing(t *testing.T) {
	clearOAuthEnv(t)

	provider, ok := newRuntimeOAuthTokenProvider()
	if ok {
		t.Fatalf("expected oauth provider loading to fail without required env, got %T", provider)
	}
}

func TestNewRuntimeOAuthTokenProvider_LoadsWhenConfigAndRefreshTokenPresent(t *testing.T) {
	clearOAuthEnv(t)

	tokenFile := filepath.Join(t.TempDir(), "oauth-refresh.json")
	store := goauth.NewFileRefreshTokenStore(tokenFile)
	if err := store.Save(t.Context(), "stored-refresh-token"); err != nil {
		t.Fatalf("seed refresh token file: %v", err)
	}

	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "http://localhost:8080/api/v1/auth/google/callback")
	t.Setenv("GOOGLE_OAUTH_SCOPES", "scope-a,scope-b")
	t.Setenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE", tokenFile)

	provider, ok := newRuntimeOAuthTokenProvider()
	if !ok {
		t.Fatalf("expected oauth provider loading to succeed")
	}
	if _, typed := provider.(*goauth.TokenProvider); !typed {
		t.Fatalf("expected *oauth.TokenProvider, got %T", provider)
	}
}

func TestNewRuntimeOAuthTokenProvider_LoadsWhenConfigPresentEvenIfRefreshTokenMissing(t *testing.T) {
	clearOAuthEnv(t)

	tokenFile := filepath.Join(t.TempDir(), "oauth-refresh.json")

	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "http://localhost:8080/api/v1/auth/google/callback")
	t.Setenv("GOOGLE_OAUTH_SCOPES", "scope-a,scope-b")
	t.Setenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE", tokenFile)

	provider, ok := newRuntimeOAuthTokenProvider()
	if !ok {
		t.Fatalf("expected oauth provider loading to succeed when config is present")
	}
	if _, typed := provider.(*goauth.TokenProvider); !typed {
		t.Fatalf("expected *oauth.TokenProvider, got %T", provider)
	}
}

func TestOAuthBootstrapInventorySource_ReturnsEmptySnapshotOnMissingRefreshToken(t *testing.T) {
	source := oauthBootstrapInventorySource{
		primary: inventorySourceFunc(func(context.Context) (domain.InventorySnapshot, error) {
			return domain.InventorySnapshot{}, errors.New("retrieve access token: missing refresh token; run OAuth consent flow first")
		}),
	}

	snapshot, err := source.LoadInventory(context.Background())
	if err != nil {
		t.Fatalf("expected empty bootstrap snapshot, got error: %v", err)
	}
	if len(snapshot.Rooms) != 0 {
		t.Fatalf("expected no rooms during oauth bootstrap, got %d", len(snapshot.Rooms))
	}
	if len(snapshot.Buildings) != 0 {
		t.Fatalf("expected no buildings during oauth bootstrap, got %d", len(snapshot.Buildings))
	}
}

func TestOAuthBootstrapInventorySource_PropagatesNonRefreshTokenErrors(t *testing.T) {
	source := oauthBootstrapInventorySource{
		primary: inventorySourceFunc(func(context.Context) (domain.InventorySnapshot, error) {
			return domain.InventorySnapshot{}, errors.New("quota exceeded")
		}),
	}

	_, err := source.LoadInventory(context.Background())
	if err == nil {
		t.Fatalf("expected non-refresh-token errors to be propagated")
	}
}

func TestNewRouter_ExposesGoogleOAuthStartPath(t *testing.T) {
	clearOAuthEnv(t)
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/start", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected /api/v1/auth/google/start to return %d when OAuth is not configured, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestNewRouter_ExposesOpenAPIDocPath(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs/openapi.json", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/docs/openapi.json to return %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON OpenAPI body, got error: %v", err)
	}

	version, ok := payload["openapi"].(string)
	if !ok {
		version, ok = payload["swagger"].(string)
	}
	if !ok {
		t.Fatalf("expected swagger/openapi version field in doc response")
	}
	if version != "2.0" {
		t.Fatalf("expected Swagger version 2.0, got %q", version)
	}
}

func TestNewRouter_ExposesSwaggerUIPath(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs/swagger/index.html", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/docs/swagger/index.html to return %d, got %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(strings.ToLower(w.Body.String()), "swagger") {
		t.Fatalf("expected swagger UI HTML response body")
	}
}

func TestNewRouter_ExposesGoogleOAuthCallbackPath(t *testing.T) {
	clearOAuthEnv(t)
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?state=test&code=test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected /api/v1/auth/google/callback to return %d when OAuth is not configured, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestNewRouter_ExposesHealthAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/health to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_ExposesBuildingsAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/buildings", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/buildings to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_ExposesRoomDetailAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/AMPHI-A", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected /api/v1/rooms/AMPHI-A to return %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestNewRouter_ExposesRoomsAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/rooms to return %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNewRouter_ExposesRoomScheduleAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/rooms/AMPHI-A/schedule?start=2026-03-09&end=2026-03-09",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected /api/v1/rooms/AMPHI-A/schedule to return %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestNewRouter_HealthResponseMatchesContract(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/health to return %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	requiredKeys := []string{
		"status",
		"version",
		"google_admin_api_connected",
		"google_calendar_api_connected",
		"last_sync",
		"response_time_ms",
	}

	for _, key := range requiredKeys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in /api/v1/health response", key)
		}
	}
}

func TestNewRouter_Error400UsesStandardFormat(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms?capacity_min=not-a-number", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "INVALID_PARAMETERS" {
		t.Fatalf("expected error.code %q, got %q", "INVALID_PARAMETERS", code)
	}
}

func TestNewRouter_Error404UsesStandardFormat(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown-endpoint", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	assertStandardErrorResponse(t, w.Body.Bytes())
}

func TestNewRouter_Error503UsesStandardFormat(t *testing.T) {
	clearOAuthEnv(t)
	r := NewRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/auth/google/start",
		nil,
	)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "GOOGLE_SERVICE_UNAVAILABLE" {
		t.Fatalf("expected error.code %q, got %q", "GOOGLE_SERVICE_UNAVAILABLE", code)
	}
}

func TestNewRouter_DoesNotPanicOnRuntimeBootstrapFailure(t *testing.T) {
	clearOAuthEnv(t)
	t.Setenv("DATA_SOURCE", "google")
	t.Setenv("GOOGLE_ADMIN_BEARER_TOKEN", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_JSON_BASE64", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("GOOGLE_ADMIN_IMPERSONATED_USER", "")

	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "GOOGLE_SERVICE_UNAVAILABLE" {
		t.Fatalf("expected error.code %q, got %q", "GOOGLE_SERVICE_UNAVAILABLE", code)
	}
}

func TestNewRouter_ErrorResponseIncludesTimestamp(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms?capacity_max=invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())
	if _, ok := errObj["timestamp"]; !ok {
		t.Fatalf("expected error.timestamp to be present")
	}
}

func TestNewRouter_RoomNotFoundReturnsRoomNotFoundCode(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/NOPE-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "ROOM_NOT_FOUND" {
		t.Fatalf("expected error.code %q, got %q", "ROOM_NOT_FOUND", code)
	}

	message, ok := errObj["message"].(string)
	if !ok {
		t.Fatalf("expected error.message to be a string")
	}
	if !strings.Contains(message, "NOPE-404") {
		t.Fatalf("expected error.message to contain missing room code, got %q", message)
	}
}

func TestNewRouter_RecoveryUsesStandardErrorFormat(t *testing.T) {
	r := NewRouter()
	r.GET("/api/v1/panic", func(_ *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	errObj := assertStandardErrorResponse(t, w.Body.Bytes())

	code, ok := errObj["code"].(string)
	if !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if code != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("expected error.code %q, got %q", "INTERNAL_SERVER_ERROR", code)
	}

	message, ok := errObj["message"].(string)
	if !ok {
		t.Fatalf("expected error.message to be a string")
	}
	if strings.Contains(strings.ToLower(message), "panic") {
		t.Fatalf("expected panic details to stay hidden, got %q", message)
	}
}

func assertStandardErrorResponse(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}

	if len(payload) != 1 {
		t.Fatalf("expected only error envelope at top level, got %d fields", len(payload))
	}

	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected payload.error to be an object, got %T", payload["error"])
	}

	if len(errObj) != 3 {
		t.Fatalf("expected payload.error to contain exactly 3 fields, got %d", len(errObj))
	}

	if _, ok := errObj["code"].(string); !ok {
		t.Fatalf("expected error.code to be a string")
	}
	if _, ok := errObj["message"].(string); !ok {
		t.Fatalf("expected error.message to be a string")
	}

	timestamp, ok := errObj["timestamp"].(string)
	if !ok {
		t.Fatalf("expected error.timestamp to be a string")
	}
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		t.Fatalf("expected error.timestamp RFC3339, got %q: %v", timestamp, err)
	}

	return errObj
}

func makeRouterTestServiceAccountJSON(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	return fmt.Sprintf(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "test-private-key-id",
		"private_key": %q,
		"client_email": "service-account@test-project.iam.gserviceaccount.com",
		"client_id": "123456789012345678901",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service-account%%40test-project.iam.gserviceaccount.com"
	}`, string(privateKeyPEM))
}

func clearOAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URI", "")
	t.Setenv("GOOGLE_OAUTH_SCOPES", "")
	t.Setenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE", "")
	t.Setenv("DATA_SOURCE", "")

	// Ensure external shell state cannot interfere with tests.
	_ = os.Unsetenv("GOOGLE_OAUTH_CLIENT_ID")
	_ = os.Unsetenv("GOOGLE_OAUTH_CLIENT_SECRET")
	_ = os.Unsetenv("GOOGLE_OAUTH_REDIRECT_URI")
	_ = os.Unsetenv("GOOGLE_OAUTH_SCOPES")
	_ = os.Unsetenv("GOOGLE_OAUTH_REFRESH_TOKEN_FILE")
	_ = os.Unsetenv("DATA_SOURCE")
}

func clearDataSourceEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATA_SOURCE", "")
	_ = os.Unsetenv("DATA_SOURCE")
}

type inventorySourceFunc func(context.Context) (domain.InventorySnapshot, error)

func (f inventorySourceFunc) LoadInventory(ctx context.Context) (domain.InventorySnapshot, error) {
	return f(ctx)
}
