package config

// Lightweight GORM + MySQL connector with retry and sensible pool defaults.
// References:
// - GORM Quickstart: https://gorm.io/docs/
// - go-sql-driver DSN options: https://github.com/go-sql-driver/mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Test hooks (overridable in tests).
// - OpenGorm wraps gorm.Open so tests can inject failures or alternate drivers.
// - SleepFn is used for retry backoff and can be replaced to avoid slow tests.
// These defaults are identical to the real implementations in production.
var (
	OpenGorm = gorm.Open
	SleepFn  = time.Sleep
	// test hook: allow ping behavior to be overridden in tests
	PingFn = func(db *sql.DB) error { return db.Ping() }
)

func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// InitDB opens a GORM connection (MySQL) and configures connection pool.
func InitDB() error {
	host := envOr("DB_HOST", "127.0.0.1")
	port := envOr("DB_PORT", "3306")
	user := envOr("DB_USER", "root")
	pass := os.Getenv("DB_PASS")
	name := envOr("DB_NAME", "sotekre_dev")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		user, pass, host, port, name)

	var err error
	// Retry loop for transient DB startup (useful with docker-compose)
	for i := 0; i < 6; i++ {
		DB, err = OpenGorm(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			// configure underlying sql.DB
			sqlDB, derr := DB.DB()
			if derr != nil {
				return derr
			}
			sqlDB.SetMaxOpenConns(25)
			sqlDB.SetMaxIdleConns(5)
			sqlDB.SetConnMaxLifetime(5 * time.Minute)
			// quick ping (testable via PingFn)
			if pingErr := PingFn(sqlDB); pingErr != nil {
				err = pingErr
			} else {
				log.Printf("connected to database %s@%s:%s", user, host, port)
				return nil
			}
		}

		log.Printf("db connect attempt %d failed: %v — retrying in 2s", i+1, err)
		SleepFn(2 * time.Second)
	}

	return fmt.Errorf("could not connect to database: %w", err)
}

// CloseDB closes the underlying connection pool.
func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
