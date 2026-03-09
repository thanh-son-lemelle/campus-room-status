package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter_ExposesHealthAtAPIV1Path(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/health to return %d, got %d", http.StatusOK, w.Code)
	}
}
