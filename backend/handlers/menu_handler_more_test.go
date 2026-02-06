package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"github.com/galpt/sotekre/backend/routes"
	"github.com/stretchr/testify/require"
)

func TestCreateMenu_withOptionalFields_viaHTTP(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()

	r := routes.SetupRouter()
	payload := map[string]interface{}{"title": "Root", "url": "/x", "order": 3}
	b, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/menus/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	data := res["data"].(map[string]any)
	require.Equal(t, "/x", data["url"].(string))
	require.Equal(t, float64(3), data["order"].(float64))
}

func TestUpdateMenu_updateParentAndOrder_viaHTTP(t *testing.T) {
	setupInMemoryDB(t)
	defer config.CloseDB()
	// create parent and child
	parent := models.Menu{Title: "P"}
	config.DB.Create(&parent)
	child := models.Menu{Title: "C"}
	config.DB.Create(&child)

	r := routes.SetupRouter()
	update := map[string]interface{}{"parent_id": parent.ID, "order": 5}
	b, _ := json.Marshal(update)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/menus/"+strconv.Itoa(int(child.ID)), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var got models.Menu
	require.NoError(t, config.DB.First(&got, child.ID).Error)
	require.NotNil(t, got.ParentID)
	require.Equal(t, 5, got.Order)
}
