package middleware

import (
	"net/http"
	"time"

	"digital-memorial/api/internal/services"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "session"
	CSRFCookieName    = "csrf_token"
)

// AuthRequired verifies the session cookie and, on success, implements the
// sliding idle-timeout: it re-issues the cookie with a renewed expiry on
// every authenticated request, so "N minutes of inactivity" is measured from
// the last request rather than from login time. It also throttles a
// last_active_at write for activity tracing.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated", "code": "unauthenticated"})
			return
		}

		claims, err := services.ParseToken(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired", "code": "session_expired"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		timeoutMinutes := services.GetSettingInt("session_timeout_minutes", 120)
		ttl := time.Duration(timeoutMinutes) * time.Minute
		newToken, expiresAt, issueErr := services.IssueToken(claims.UserID, claims.Email, claims.Role, ttl)
		if issueErr == nil {
			SetSessionCookie(c, newToken, expiresAt)
		}

		services.TouchLastActive(claims.UserID)

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")
		roleStr, _ := role.(string)
		if !allowed[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "code": "forbidden"})
			return
		}
		c.Next()
	}
}

func SetSessionCookie(c *gin.Context, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	c.SetCookie(SessionCookieName, token, maxAge, "/", "", isSecureRequest(c), true)
}

// SetCSRFCookie issues a readable (non-HttpOnly) CSRF token cookie alongside
// the session cookie so the frontend can echo it back as X-CSRF-Token.
func SetCSRFCookie(c *gin.Context, token string, maxAge int) {
	c.SetCookie(CSRFCookieName, token, maxAge, "/", "", isSecureRequest(c), false)
}

func ClearAuthCookies(c *gin.Context) {
	c.SetCookie(SessionCookieName, "", -1, "/", "", isSecureRequest(c), true)
	c.SetCookie(CSRFCookieName, "", -1, "/", "", isSecureRequest(c), false)
}

func isSecureRequest(c *gin.Context) bool {
	return c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
}
