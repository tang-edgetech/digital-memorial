package middleware

import (
	"net/http"
	"strings"

	"digital-memorial/api/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupGate blocks every route except /api/setup/** and /api/health until the
// first-run setup wizard has completed, so the app can't be used against a
// half-configured database.
func SetupGate() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/api/health" || strings.HasPrefix(path, "/api/setup/") {
			c.Next()
			return
		}
		if !services.IsSetupCompleted() {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "setup required", "code": "setup_required"})
			return
		}
		c.Next()
	}
}

// RequireSetupIncompleteOrSuperAdmin lets /api/setup/** run freely before
// setup has completed (no admin exists yet to authenticate as); once setup is
// done, those same endpoints require an authenticated Super Admin so the
// wizard can't be re-run anonymously.
func RequireSetupIncompleteOrSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !services.IsSetupCompleted() {
			c.Next()
			return
		}
		AuthRequired()(c)
		if c.IsAborted() {
			return
		}
		RequireRole(string("super_admin"))(c)
	}
}
