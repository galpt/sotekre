package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"github.com/galpt/sotekre/backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MenuSuite struct {
	suite.Suite
	r *gin.Engine
}

func (s *MenuSuite) SetupTest() {
	// in-memory DB per-suite
	dsn := "file:memtest_handlers_suite?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)
	config.DB = db
	require.NoError(s.T(), config.DB.AutoMigrate(&models.Menu{}))
}

func (s *MenuSuite) TearDownTest() {
	_ = config.CloseDB()
}

func (s *MenuSuite) TestCreateAndGetMenu_viaRouter() {
	r := routes.SetupRouter()

	// create root
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/menus/", bytesFromString(`{"title":"suite-root"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(s.T(), http.StatusCreated, rec.Code)

	// fetch list
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/menus/", nil)
	r.ServeHTTP(rec, req)
	require.Equal(s.T(), http.StatusOK, rec.Code)
}

func TestMenuSuite(t *testing.T) {
	suite.Run(t, new(MenuSuite))
}

// helper: small convenience to create a *bytes.Reader from string without extra imports
func bytesFromString(s string) *bytes.Reader { return bytes.NewReader([]byte(s)) }
