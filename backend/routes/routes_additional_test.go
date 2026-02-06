package routes_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/galpt/sotekre/backend/routes"
)

func TestOpenAPI_and_DOCS_404_when_missing(t *testing.T) {
	// locate repo's backend/docs directory (same logic as SetupRouter)
	// simulate "docs not present" without mutating the repo on-disk
	os.Setenv("SOTEKRE_TEST_NO_DOCS", "1")
	defer os.Unsetenv("SOTEKRE_TEST_NO_DOCS")

	r := routes.SetupRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /openapi.json when docs missing, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /docs when docs missing, got %d", rec.Code)
	}
}

func TestCORS_Respects_ENV_ALLOW_ORIGINS(t *testing.T) {
	os.Setenv("CORS_ALLOW_ORIGINS", "https://example.test")
	defer os.Unsetenv("CORS_ALLOW_ORIGINS")

	r := routes.SetupRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/menus/", nil)
	req.Header.Set("Origin", "https://example.test")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("expected 204/200 for CORS preflight, got %d", rec.Code)
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "https://example.test" {
		t.Fatalf("unexpected Access-Control-Allow-Origin: %q", got)
	}
}
