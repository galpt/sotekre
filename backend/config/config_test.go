package config

import (
	"database/sql"
	"errors"
	"fmt"
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
	origOpen := OpenGorm
	origSleep := SleepFn
	defer func() { OpenGorm = origOpen; SleepFn = origSleep }()

	// simulate permanent open failure
	OpenGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return nil, errors.New("open-failure")
	}
	// avoid waiting during retries
	SleepFn = func(d time.Duration) {}

	err := InitDB()
	require.Error(t, err)

	// simulate success via sqlite
	OpenGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open("file:memtest_init?mode=memory&cache=shared"), opts...)
	}
	SleepFn = func(d time.Duration) {}
	require.NoError(t, InitDB())
	require.NotNil(t, DB)
	_ = CloseDB()
}

func TestInitDB_pingFailure_retries_thenSucceeds(t *testing.T) {
	origOpen := OpenGorm
	origPing := PingFn
	origSleep := SleepFn
	defer func() { OpenGorm = origOpen; PingFn = origPing; SleepFn = origSleep }()

	// open succeeds, but ping fails once then succeeds
	OpenGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open("file:memtest_ping?mode=memory&cache=shared"), opts...)
	}
	calls := 0
	PingFn = func(db *sql.DB) error {
		calls++
		if calls == 1 {
			return fmt.Errorf("transient ping")
		}
		return nil
	}
	SleepFn = func(d time.Duration) {}

	require.NoError(t, InitDB())
	_ = CloseDB()
}

func TestInitDB_pingPermanentFailure_returnsError(t *testing.T) {
	origOpen := OpenGorm
	origPing := PingFn
	origSleep := SleepFn
	defer func() { OpenGorm = origOpen; PingFn = origPing; SleepFn = origSleep }()

	OpenGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open("file:memtest_ping2?mode=memory&cache=shared"), opts...)
	}
	PingFn = func(db *sql.DB) error { return fmt.Errorf("unreachable") }
	SleepFn = func(d time.Duration) {}

	reqErr := InitDB()
	require.Error(t, reqErr)
}
