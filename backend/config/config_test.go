package config

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnvOr(t *testing.T) {
	require.Equal(t, "bar", func() string {
		os.Setenv("FOO_TEST_ENVOR", "bar")
		defer os.Unsetenv("FOO_TEST_ENVOR")
		return envOr("FOO_TEST_ENVOR", "def")
	}(), "envOr should return the set env value")

	require.Equal(t, "def", envOr("MISSING_TEST_ENVOR", "def"), "envOr should return default when unset")
}

func TestCloseDB_NilAndSQLite(t *testing.T) {
	// nil DB -> no-op
	DB = nil
	require.NoError(t, CloseDB())

	// in-memory sqlite -> closes cleanly
	dsn := "file:memtest_config?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	require.NoError(t, CloseDB())
}

func TestInitDB_overridableHooks(t *testing.T) {
	// ensure InitDB propagates errors and uses sleep/open hooks without sleeping
	origOpen := openGorm
	origSleep := sleepFn
	defer func() { openGorm = origOpen; sleepFn = origSleep }()

	// simulate permanent open failure
	openGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return nil, errors.New("open-failure")
	}
	// avoid waiting during retries
	sleepFn = func(d time.Duration) {}

	err := InitDB()
	require.Error(t, err)

	// simulate success via sqlite
	openGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open("file:memtest_init?mode=memory&cache=shared"), opts...)
	}
	sleepFn = func(d time.Duration) {}
	require.NoError(t, InitDB())
	require.NotNil(t, DB)
	_ = CloseDB()
}