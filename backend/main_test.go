package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/galpt/sotekre/backend/config"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRun_startsAndGracefullyShutsDown(t *testing.T) {
	origOpen := config.OpenGorm
	origPing := config.PingFn
	defer func() { config.OpenGorm = origOpen; config.PingFn = origPing }()

	// make InitDB use sqlite in-memory and a no-op ping
	config.OpenGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open("file:memtest_main?mode=memory&cache=shared"), opts...)
	}
	config.PingFn = func(db *sql.DB) error { return nil }

	// find a free port and export it so run() binds to it
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	os.Setenv("PORT", fmt.Sprintf("%d", port))
	defer os.Unsetenv("PORT")
	// avoid docs file resolution during the test
	os.Setenv("SOTEKRE_TEST_NO_DOCS", "1")
	defer os.Unsetenv("SOTEKRE_TEST_NO_DOCS")

	quit := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() { done <- run(quit) }()

	// wait for server to become responsive (openapi.json may be 200 or 404)
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("server did not start in time")
		default:
			res, err := http.Get(fmt.Sprintf("http://localhost:%d/openapi.json", port))
			if err == nil {
				_ = res.Body.Close()
				if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotFound {
					goto ready
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
ready:
	// trigger graceful shutdown
	quit <- os.Interrupt
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("run did not return after shutdown")
	}
}

func TestRun_withDBHostDocker_warningAndInit(t *testing.T) {
	origOpen := config.OpenGorm
	origPing := config.PingFn
	defer func() { config.OpenGorm = origOpen; config.PingFn = origPing }()

	config.OpenGorm = func(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open("file:memtest_main2?mode=memory&cache=shared"), opts...)
	}
	config.PingFn = func(db *sql.DB) error { return nil }

	os.Setenv("DB_HOST", "db")
	os.Unsetenv("DB_PASS")
	defer os.Unsetenv("DB_HOST")

	// use a quick-run: provide quit channel and close it immediately after server ready
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	os.Setenv("PORT", fmt.Sprintf("%d", port))
	defer os.Unsetenv("PORT")
	os.Setenv("SOTEKRE_TEST_NO_DOCS", "1")
	defer os.Unsetenv("SOTEKRE_TEST_NO_DOCS")

	quit := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() { done <- run(quit) }()

	// wait for server then signal shutdown
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("server did not start in time")
		default:
			res, err := http.Get(fmt.Sprintf("http://localhost:%d/openapi.json", port))
			if err == nil {
				_ = res.Body.Close()
				if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotFound {
					quit <- os.Interrupt
					goto waitdone
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
waitdone:
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("run did not return after shutdown")
	}
}
