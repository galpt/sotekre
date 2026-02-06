package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"github.com/galpt/sotekre/backend/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInMemoryDBForDocs(t *testing.T) {
	dsn := "file:memtest_docs?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory: %v", err)
	}
	config.DB = db
	if err := config.DB.AutoMigrate(&models.Menu{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
}

func TestDocsEndpointsAvailable(t *testing.T) {
	setupInMemoryDBForDocs(t)
	defer config.CloseDB()

	r := routes.SetupRouter()

	// /openapi.json (existing static OpenAPI)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /openapi.json, got %d", rec.Code)
	}

	// /docs (legacy static page)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /docs, got %d", rec.Code)
	}

	// /swagger/index.html (gin-swagger UI should be registered)
	// swaggerFiles serves embedded UI and is pointed at /openapi.json by default in router
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /swagger/index.html, got %d", rec.Code)
	}

	// quick sanity: the UI should be returned quickly (not a redirect loop)
	// allow a tiny amount of time to emulate integration behaviour
	time.Sleep(5 * time.Millisecond)
}
