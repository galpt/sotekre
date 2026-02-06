// @title Sotekre — Menu Tree API
// @version 0.1.0
// @description Minimal OpenAPI for the Menu Tree MVP
// @contact.name API Support
// @contact.email dev@example.com
// @host localhost:8080
// @BasePath /api
//
//go:generate swag init -g main.go -o ./docs --outputTypes go,json,yaml
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"github.com/galpt/sotekre/backend/routes"
	"github.com/joho/godotenv"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// load .env if present
	_ = godotenv.Load()

	// Helpful runtime hint: if the app is started in Docker (DB_HOST=db) but the
	// DB password is empty, the official MySQL image will fail to start because
	// it requires a non-empty MYSQL_ROOT_PASSWORD. This is *only* a warning.
	if os.Getenv("DB_HOST") == "db" && os.Getenv("DB_PASS") == "" {
		log.Println("WARNING: DB_HOST=db but DB_PASS is empty — Docker MySQL requires a non-empty MYSQL_ROOT_PASSWORD. Use .env.docker or set MYSQL_ROOT_PASSWORD when running docker-compose.")
	}

	if err := config.InitDB(); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer config.CloseDB()

	// Auto-migrate schema (safe for interview / MVP)
	if err := config.DB.AutoMigrate(&models.Menu{}); err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server stopped")
}
