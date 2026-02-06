package routes

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/galpt/sotekre/backend/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures routes and middleware
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Disable automatic redirect for trailing slashes to prevent CORS issues
	r.RedirectTrailingSlash = false

	// Simple CORS (adjust for production)
	cfg := cors.DefaultConfig()
	allow := os.Getenv("CORS_ALLOW_ORIGINS")
	if allow == "" {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = []string{allow}
	}
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(cfg))

	api := r.Group("/api")
	{
		menus := api.Group("/menus")
		{
			// Register both with and without trailing slash for compatibility
			// Frontend calls without trailing slash, tests call with trailing slash
			menus.GET("", handlers.GetMenus)
			menus.GET("/", handlers.GetMenus)
			menus.POST("", handlers.CreateMenu)
			menus.POST("/", handlers.CreateMenu)
			menus.PUT("/:id", handlers.UpdateMenu)
			menus.PATCH("/:id/reorder", handlers.ReorderMenu)
			menus.PATCH("/:id/move", handlers.MoveMenu)
			menus.DELETE("/:id", handlers.DeleteMenu)
		}
	}

	// serve static frontend (simple SPA)
	r.StaticFile("/", "../frontend/index.html")
	r.Static("/static", "../frontend")

	// Serve OpenAPI + Swagger UI (resolve files relative to source so tests and
	// `go test` don't depend on working directory)
	findDocsFile := func(name string) (string, bool) {
		_, callerFile, _, ok := runtime.Caller(0)
		if !ok {
			return "", false
		}
		base := filepath.Dir(callerFile) // backend/routes
		candidates := []string{
			filepath.Join(base, "..", "docs", name),
			filepath.Join(base, "docs", name),
			filepath.Join(base, "routes", "docs", name), // support tests that create `routes/docs` inside the package
			filepath.Join("routes", "docs", name),       // support working-dir-relative `routes/docs`
			filepath.Join("docs", name),
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				return p, true
			}
		}
		return "", false
	}

	// prefer OpenAPI v3 `openapi.json` (checked in by the repo) then fall back
	// to the swag-generated `swagger.json` if present. In tests we can force the
	// "docs missing" branch by setting SOTEKRE_TEST_NO_DOCS=1.
	skipDocs := os.Getenv("SOTEKRE_TEST_NO_DOCS") == "1"
	if !skipDocs {
		if p, ok := findDocsFile("openapi.json"); ok {
			r.GET("/openapi.json", func(c *gin.Context) { c.File(p) })
		} else if p, ok := findDocsFile("swagger.json"); ok {
			r.GET("/openapi.json", func(c *gin.Context) { c.File(p) })
		} else {
			// helpful 404 so users know how to generate docs.
			r.GET("/openapi.json", func(c *gin.Context) {
				c.JSON(404, gin.H{"error": "openapi.json not found; run `go generate ./...` in backend"})
			})
		}
	} else {
		// test-only: simulate missing docs
		r.GET("/openapi.json", func(c *gin.Context) {
			c.JSON(404, gin.H{"error": "openapi.json not found; run `go generate ./...` in backend"})
		})
	}

	if !skipDocs {
		if p, ok := findDocsFile("swagger.html"); ok {
			r.GET("/docs", func(c *gin.Context) { c.File(p) })
		} else {
			r.GET("/docs", func(c *gin.Context) {
				c.String(404, "API docs not generated — run `go generate ./...` in backend")
			})
		}
	} else {
		r.GET("/docs", func(c *gin.Context) {
			c.String(404, "API docs not generated — run `go generate ./...` in backend")
		})
	}

	// Swagger UI (dynamic) — serves embedded Swagger UI and points it at
	// `/openapi.json` (which above resolves to repo file or generated file).
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.json")))

	return r
}
