package router

import (
	"os"

	"digital-memorial/api/internal/handlers"
	"digital-memorial/api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// New wires the full middleware chain and route groups. Order matters:
// SetupGate runs first so an unconfigured install only ever exposes
// /api/setup/** and /api/health; auth + CSRF are then layered onto the
// protected group only.
func New() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	allowedOrigin := os.Getenv("WEB_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}
	r.Use(middleware.CORS(allowedOrigin))
	r.Use(middleware.SetupGate())

	r.GET("/api/health", handlers.Health)
	r.GET("/api/setup/status", handlers.SetupStatus)

	setupGroup := r.Group("/api/setup")
	setupGroup.Use(middleware.RequireSetupIncompleteOrSuperAdmin())
	{
		setupGroup.POST("/db", handlers.SetupDB)
		setupGroup.POST("/admin", handlers.SetupAdmin)
		setupGroup.POST("/settings", handlers.SetupSettings)
	}

	authGroup := r.Group("/api/auth")
	authGroup.Use(middleware.RateLimit(1, 5))
	{
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/logout", handlers.Logout)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired(), middleware.CSRF())
	{
		protected.GET("/auth/me", handlers.Me)
		protected.GET("/settings", handlers.GetSettings)
		protected.PUT("/settings", middleware.RequireRole("super_admin", "admin"), handlers.UpdateSettingsHandler)
	}

	return r
}
