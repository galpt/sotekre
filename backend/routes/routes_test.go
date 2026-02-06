package routes_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/galpt/sotekre/backend/routes"
	"github.com/stretchr/testify/require"
)

func TestSetupRouter_OpenAPI_and_Docs_Present(t *testing.T) {
	r := routes.SetupRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSetupRouter_NoDocsEnv_returns404(t *testing.T) {
	// emulate environment where docs are intentionally hidden
	os.Setenv("SOTEKRE_TEST_NO_DOCS", "1")
	defer os.Unsetenv("SOTEKRE_TEST_NO_DOCS")

	r := routes.SetupRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRootStaticServesIndex(t *testing.T) {
	// ensure the expected frontend index exists for the test (make the test hermetic)
	p := filepath.Join("..", "frontend")
	require.NoError(t, os.MkdirAll(p, 0o755))
	defer os.RemoveAll(p)
	f := filepath.Join(p, "index.html")
	require.NoError(t, os.WriteFile(f, []byte("<html>ok</html>"), 0o644))

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	// basic sanity for content-type / body present
	require.NotEmpty(t, rec.Body.String())
}

func TestCORSAllowOriginsEnv_applies(t *testing.T) {
	os.Setenv("CORS_ALLOW_ORIGINS", "https://example.test")
	defer os.Unsetenv("CORS_ALLOW_ORIGINS")

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/menus/", nil)
	req.Header.Set("Origin", "https://example.test")
	r.ServeHTTP(rec, req)
	// preflight should be allowed for that origin (some servers return 200 or 204)
	require.Contains(t, []int{http.StatusOK, http.StatusNoContent}, rec.Code)
}

func TestFindDocsFallback_findsSwaggerJSON(t *testing.T) {
	p := filepath.Join("routes", "docs")
	require.NoError(t, os.MkdirAll(p, 0o755))
	defer os.RemoveAll(p)
	f := filepath.Join(p, "swagger.json")
	require.NoError(t, os.WriteFile(f, []byte(`{"openapi":"3.0.0"}`), 0o644))

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestFindDocsFallback_findsSwaggerHTML(t *testing.T) {
	p := filepath.Join("routes", "docs")
	require.NoError(t, os.MkdirAll(p, 0o755))
	defer os.RemoveAll(p)
	f := filepath.Join(p, "swagger.html")
	require.NoError(t, os.WriteFile(f, []byte(`<html></html>`), 0o644))

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
