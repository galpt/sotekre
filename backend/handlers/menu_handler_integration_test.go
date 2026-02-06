package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"context"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"github.com/galpt/sotekre/backend/routes"
	"github.com/galpt/sotekre/backend/services"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInMemoryDB(t *testing.T) {
	// unique in-memory DSN per test to avoid interference when running whole package
	dsn := fmt.Sprintf("file:memtest_routes_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory: %v", err)
	}
	config.DB = db
	if err := config.DB.AutoMigrate(&models.Menu{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
}

func TestMenus_CRUD_viaHTTP(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()

	r := routes.SetupRouter()

	// create root
	payload := map[string]interface{}{"title": "Root X"}
	b, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/menus/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", rec.Code, rec.Body.String())
	}
	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	data := res["data"].(map[string]any)
	id := int(data["id"].(float64))

	// create child
	payload = map[string]interface{}{"title": "Child Y", "parent_id": id}
	b, _ = json.Marshal(payload)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/menus/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for child, got %d body=%s", rec.Code, rec.Body.String())
	}

	// fetch tree
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET, got %d", rec.Code)
	}
	// debug: show raw response for easier troubleshooting
	t.Logf("GET /api/menus/ response: %s", rec.Body.String())
	var listRes map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &listRes); err != nil {
		t.Fatalf("invalid json list: %v", err)
	}
	arr := listRes["data"].([]any)
	if len(arr) != 1 {
		t.Fatalf("expected 1 root, got %d", len(arr))
	}
	root := arr[0].(map[string]any)
	children := root["children"].([]any)
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	// delete root (should remove child too)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/menus/"+strconv.Itoa(id), nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for delete, got %d", rec.Code)
	}

	// confirm empty
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET after delete, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listRes); err != nil {
		t.Fatalf("invalid json list: %v", err)
	}
	v, exists := listRes["data"]
	if !exists || v == nil {
		// treat missing/null as empty list
		return
	}
	arr, ok := v.([]any)
	if !ok {
		t.Fatalf("expected data to be an array, got %T", v)
	}
	if len(arr) != 0 {
		t.Fatalf("expected 0 roots after delete, got %d", len(arr))
	}
}

func TestMenus_Move_Reorder_viaHTTP(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()

	r := routes.SetupRouter()

	// create three roots: A, B, C
	create := func(title string) int {
		payload := map[string]interface{}{"title": title}
		b, _ := json.Marshal(payload)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/menus/", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %s failed: %d %s", title, rec.Code, rec.Body.String())
		}
		var res map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("invalid json on create %s: %v", title, err)
		}
		return int(res["data"].(map[string]any)["id"].(float64))
	}

	idA := create("A")
	idB := create("B")
	idC := create("C")

	// reorder C to index 1 => expected A, C, B
	{
		rec := httptest.NewRecorder()
		body := bytes.NewReader([]byte(`{"new_order":1}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/menus/"+strconv.Itoa(idC)+"/reorder", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("reorder failed: %d %s", rec.Code, rec.Body.String())
		}
		// verify
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
		r.ServeHTTP(rec, req)
		var listRes map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &listRes); err != nil {
			t.Fatalf("invalid json list: %v", err)
		}
		arr := listRes["data"].([]any)
		if len(arr) != 3 {
			t.Fatalf("expected 3 roots after reorder, got %d", len(arr))
		}
		if int(arr[0].(map[string]any)["id"].(float64)) != idA || int(arr[1].(map[string]any)["id"].(float64)) != idC || int(arr[2].(map[string]any)["id"].(float64)) != idB {
			t.Fatalf("unexpected order after reorder: %v", []int{idA, idC, idB})
		}
	}

	// move B to be first child of A
	{
		rec := httptest.NewRecorder()
		body := bytes.NewReader([]byte(`{"new_parent_id": ` + strconv.Itoa(idA) + `, "new_order": 0}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/menus/"+strconv.Itoa(idB)+"/move", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("move failed: %d %s", rec.Code, rec.Body.String())
		}
		// verify structure: roots should be [A, C], and A should have child B
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
		r.ServeHTTP(rec, req)
		var listRes map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &listRes); err != nil {
			t.Fatalf("invalid json list: %v", err)
		}
		arr := listRes["data"].([]any)
		if len(arr) != 2 {
			t.Fatalf("expected 2 roots after move, got %d", len(arr))
		}
		rootA := arr[0].(map[string]any)
		children := rootA["children"].([]any)
		if len(children) != 1 || int(children[0].(map[string]any)["id"].(float64)) != idB {
			t.Fatalf("expected B as child of A, got %+v", children)
		}
	}

	// move B back to root at index 1 => A, B, C
	{
		rec := httptest.NewRecorder()
		body := bytes.NewReader([]byte(`{"new_parent_id": null, "new_order": 1}`))
		req := httptest.NewRequest(http.MethodPatch, "/api/menus/"+strconv.Itoa(idB)+"/move", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("move-to-root failed: %d %s", rec.Code, rec.Body.String())
		}
		// verify
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
		r.ServeHTTP(rec, req)
		var listRes map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &listRes); err != nil {
			t.Fatalf("invalid json list: %v", err)
		}
		arr := listRes["data"].([]any)
		if len(arr) != 3 {
			t.Fatalf("expected 3 roots after move-to-root, got %d", len(arr))
		}
		if int(arr[0].(map[string]any)["id"].(float64)) != idA || int(arr[1].(map[string]any)["id"].(float64)) != idB || int(arr[2].(map[string]any)["id"].(float64)) != idC {
			t.Fatalf("unexpected order after move-to-root: %v", []int{idA, idB, idC})
		}
	}
}

func TestGetMenus_ServiceError_returns500_sqlmock(t *testing.T) {
	// replace DB with sqlmock and force the SELECT to return an error
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open(sqlmock): %v", err)
	}
	config.DB = gdb

	mock.ExpectQuery("SELECT .*FROM .*menus.*").WillReturnError(fmt.Errorf("boom"))

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMenus_returnsEmptyArray_whenNoMenus(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	data := res["data"].([]any)
	require.Len(t, data, 0)
}


func TestUpdateMenu_Success_viaHTTP(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()

	r := routes.SetupRouter()

	// create menu
	payload := map[string]interface{}{"title": "original"}
	b, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/menus/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	id := int(res["data"].(map[string]any)["id"].(float64))

	// update
	update := map[string]interface{}{"title": "updated"}
	b, _ = json.Marshal(update)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/menus/"+strconv.Itoa(id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// verify via list
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var listRes map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &listRes))
	arr := listRes["data"].([]any)
	require.Len(t, arr, 1)
	got := arr[0].(map[string]any)
	require.Equal(t, "updated", got["title"].(string))
}

func TestGetMenus_EmptyList_returns200_withEmptyArray(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	data, ok := res["data"].([]any)
	require.True(t, ok)
	require.Len(t, data, 0)
}

func TestMoveMenu_Handler_returns500_whenMovingIntoDescendant(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()

	// create a -> b -> c
	if err := config.DB.Create(&models.Menu{Title: "a"}).Error; err != nil {
		t.Fatalf("create a: %v", err)
	}
	var a models.Menu
	config.DB.First(&a, "title = ?", "a")
	if err := config.DB.Create(&models.Menu{Title: "b", ParentID: &a.ID}).Error; err != nil {
		t.Fatalf("create b: %v", err)
	}
	var b models.Menu
	config.DB.First(&b, "title = ?", "b")
	if err := config.DB.Create(&models.Menu{Title: "c", ParentID: &b.ID}).Error; err != nil {
		t.Fatalf("create c: %v", err)
	}
	var c models.Menu
	config.DB.First(&c, "title = ?", "c")

	r := routes.SetupRouter()
	body := bytes.NewReader([]byte(fmt.Sprintf(`{"new_parent_id": %d}`, c.ID)))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/"+strconv.Itoa(int(a.ID))+"/move", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteMenu_ServiceError_returns500_sqlmock(t *testing.T) {
	orig := services.DeleteMenuRecursiveFn
	defer func() { services.DeleteMenuRecursiveFn = orig }()
	services.DeleteMenuRecursiveFn = func(ctx context.Context, id uint) error { return fmt.Errorf("boom") }

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/menus/42", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestReorderMenu_ServiceError_returns500_sqlmock(t *testing.T) {
	// stub the service to return an error and assert the handler returns 500 —
	// keeps handler-level tests deterministic and avoids SQL-level brittleness.
	orig := services.ReorderMenuFn
	defer func() { services.ReorderMenuFn = orig }()
	services.ReorderMenuFn = func(ctx context.Context, id uint, newOrder int) error { return fmt.Errorf("boom") }

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/1/reorder", bytes.NewReader([]byte(`{"new_order": 1}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMoveMenu_Handler_sqlExecError_returns500_sqlmock(t *testing.T) {
	orig := services.MoveMenuFn
	defer func() { services.MoveMenuFn = orig }()
	services.MoveMenuFn = func(ctx context.Context, id uint, newParentID *uint, newOrder *int) error { return fmt.Errorf("boom") }

	r := routes.SetupRouter()
	rec := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{"new_parent_id": null, "new_order": 0}`))
	req := httptest.NewRequest(http.MethodPatch, "/api/menus/2/move", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateMenu_ServiceError_returns500_sqlmock(t *testing.T) {
	orig := services.CreateMenuFn
	defer func() { services.CreateMenuFn = orig }()
	services.CreateMenuFn = func(ctx context.Context, m *models.Menu) error { return fmt.Errorf("boom") }

	r := routes.SetupRouter()
	payload := map[string]interface{}{"title": "x"}
	b, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/menus/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateMenu_ServiceError_returns500_sqlmock(t *testing.T) {
	orig := services.UpdateMenuFn
	defer func() { services.UpdateMenuFn = orig }()
	services.UpdateMenuFn = func(ctx context.Context, id uint, upd map[string]interface{}) error { return fmt.Errorf("boom") }

	r := routes.SetupRouter()
	update := map[string]interface{}{"title": "updated"}
	b, _ := json.Marshal(update)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/menus/1", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
