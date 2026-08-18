package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CSRF enforces the double-submit cookie pattern on mutating requests: the
// value in the readable csrf_token cookie must match the X-CSRF-Token header.
// Only mounted on the authenticated route group — login/logout/setup happen
// before a session (and therefore a CSRF cookie) exists.
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing csrf token", "code": "csrf_missing"})
			return
		}

		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken == "" || headerToken != cookieToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid csrf token", "code": "csrf_invalid"})
			return
		}

		c.Next()
	}
}
