package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/galpt/sotekre/backend/routes"
)

func TestUpdateMenu_InvalidID_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/menus/invalid-id", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", rec.Code)
	}
}

func TestReorderMenu_MissingOrNegativeNewOrder(t *testing.T) {
	r := routes.SetupRouter()

	// missing new_order
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/1/reorder", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when new_order missing, got %d", rec.Code)
	}

	// negative new_order
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/menus/1/reorder", bytes.NewReader([]byte(`{"new_order": -1}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when new_order negative, got %d", rec.Code)
	}
}

func TestCreateMenu_BadRequest_MissingTitle(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/menus/", bytes.NewReader([]byte(`{"url":"/x"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when title missing, got %d", rec.Code)
	}
}

func TestUpdateMenu_NoUpdatableFields_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/menus/1", bytes.NewReader([]byte(`{"foo":"bar"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when no updatable fields provided, got %d", rec.Code)
	}
}

func TestMoveMenu_NewOrderNegative_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/1/move", bytes.NewReader([]byte(`{"new_parent_id": null, "new_order": -1}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when new_order is negative, got %d", rec.Code)
	}
}

func TestUpdateMenu_BadJSON_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/menus/1", bytes.NewReader([]byte(`{"title":`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestReorderMenu_BadJSON_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/1/reorder", bytes.NewReader([]byte(`{"new_order":`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestMoveMenu_BadJSON_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/1/move", bytes.NewReader([]byte(`{"new_parent_id":`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestMoveMenu_InvalidID_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/invalid/move", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", rec.Code)
	}
}

func TestDeleteMenu_InvalidID_Returns400(t *testing.T) {
	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/menus/invalid-id", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", rec.Code)
	}
}
